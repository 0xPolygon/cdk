package tree

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/russross/meddler"
	"golang.org/x/crypto/sha3"
)

var (
	EmptyProof = types.Proof{}
)

type Tree struct {
	db         *sql.DB
	zeroHashes []common.Hash
	rhtTable   string
	rootTable  string
}

func newTreeNode(left, right common.Hash) types.TreeNode {
	var hash common.Hash
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(left[:])
	hasher.Write(right[:])
	copy(hash[:], hasher.Sum(nil))
	return types.TreeNode{
		Hash:  hash,
		Left:  left,
		Right: right,
	}
}

func newTree(db *sql.DB, tablePrefix string) *Tree {
	t := &Tree{
		db:         db,
		zeroHashes: generateZeroHashes(types.DefaultHeight),
		rhtTable:   tablePrefix + "rht",
		rootTable:  tablePrefix + "root",
	}

	return t
}

func (t *Tree) getSiblings(tx db.Querier, index uint32, root common.Hash) (
	siblings types.Proof,
	hasUsedZeroHashes bool,
	err error,
) {
	currentNodeHash := root
	// It starts in height-1 because 0 is the level of the leafs
	for h := int(types.DefaultHeight - 1); h >= 0; h-- {
		var currentNode *types.TreeNode
		currentNode, err = t.getRHTNode(tx, currentNodeHash)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				hasUsedZeroHashes = true
				siblings[h] = t.zeroHashes[h]
				err = nil
				continue
			} else {
				err = fmt.Errorf(
					"height: %d, currentNode: %s, error: %w",
					h, currentNodeHash.Hex(), err,
				)
				return
			}
		}
		/*
		*        Root                (level h=3 => height=4)
		*      /     \
		*	 O5       O6             (level h=2)
		*	/ \      / \
		*  O1  O2   O3  O4           (level h=1)
		*  /\   /\   /\ /\
		* 0  1 2  3 4 5 6 7 Leafs    (level h=0)
		* Example 1:
		* Choose index = 3 => 011 binary
		* Assuming we are in level 1 => h=1; 1<<h = 010 binary
		* Now, let's do AND operation => 011&010=010 which is higher than 0 so we need the left sibling (O1)
		* Example 2:
		* Choose index = 4 => 100 binary
		* Assuming we are in level 1 => h=1; 1<<h = 010 binary
		* Now, let's do AND operation => 100&010=000 which is not higher than 0 so we need the right sibling (O4)
		* Example 3:
		* Choose index = 4 => 100 binary
		* Assuming we are in level 2 => h=2; 1<<h = 100 binary
		* Now, let's do AND operation => 100&100=100 which is higher than 0 so we need the left sibling (O5)
		 */
		if index&(1<<h) > 0 {
			siblings[h] = currentNode.Left
			currentNodeHash = currentNode.Right
		} else {
			siblings[h] = currentNode.Right
			currentNodeHash = currentNode.Left
		}
	}

	return
}

// GetProof returns the merkle proof for a given index and root.
func (t *Tree) GetProof(ctx context.Context, index uint32, root common.Hash) (types.Proof, error) {
	siblings, isErrNotFound, err := t.getSiblings(t.db, index, root)
	if err != nil {
		return types.Proof{}, err
	}
	if isErrNotFound {
		return types.Proof{}, db.ErrNotFound
	}
	return siblings, nil
}

func (t *Tree) getRHTNode(tx db.Querier, nodeHash common.Hash) (*types.TreeNode, error) {
	node := &types.TreeNode{}
	err := meddler.QueryRow(
		tx, node,
		fmt.Sprintf(`select * from %s where hash = $1`, t.rhtTable),
		nodeHash.Hex(),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return node, db.ErrNotFound
		}
		return node, err
	}
	return node, err
}

func generateZeroHashes(height uint8) []common.Hash {
	var zeroHashes = []common.Hash{
		{},
	}
	// This generates a leaf = HashZero in position 0. In the rest of the positions that are
	// equivalent to the ascending levels, we set the hashes of the nodes.
	// So all nodes from level i=5 will have the same value and same children nodes.
	for i := 1; i <= int(height); i++ {
		hasher := sha3.NewLegacyKeccak256()
		hasher.Write(zeroHashes[i-1][:])
		hasher.Write(zeroHashes[i-1][:])
		thisHeightHash := common.Hash{}
		copy(thisHeightHash[:], hasher.Sum(nil))
		zeroHashes = append(zeroHashes, thisHeightHash)
	}
	return zeroHashes
}

func (t *Tree) storeNodes(tx db.Txer, nodes []types.TreeNode) error {
	for i := 0; i < len(nodes); i++ {
		if err := meddler.Insert(tx, t.rhtTable, &nodes[i]); err != nil {
			if sqliteErr, ok := db.SQLiteErr(err); ok {
				if sqliteErr.ExtendedCode == db.UniqueConstrain {
					// ignore repeated entries. This is likely to happen due to not
					// cleaning RHT when reorg
					continue
				}
			}
			return err
		}
	}
	return nil
}

func (t *Tree) storeRoot(tx db.Txer, root types.Root) error {
	return meddler.Insert(tx, t.rootTable, &root)
}

// GetLastRoot returns the last processed root
func (t *Tree) GetLastRoot(ctx context.Context) (types.Root, error) {
	return t.getLastRootWithTx(t.db)
}

func (t *Tree) getLastRootWithTx(tx db.Querier) (types.Root, error) {
	var root types.Root
	err := meddler.QueryRow(
		tx, &root,
		fmt.Sprintf(`SELECT * FROM %s ORDER BY block_num DESC, block_position DESC LIMIT 1;`, t.rootTable),
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return root, db.ErrNotFound
		}
		return root, err
	}
	return root, nil
}

// GetRootByIndex returns the root associated to the index
func (t *Tree) GetRootByIndex(ctx context.Context, index uint32) (types.Root, error) {
	var root types.Root
	if err := meddler.QueryRow(
		t.db, &root,
		fmt.Sprintf(`SELECT * FROM %s WHERE position = $1;`, t.rootTable),
		index,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return root, db.ErrNotFound
		}
		return root, err
	}
	return root, nil
}

// GetRootByHash returns the root associated to the hash
func (t *Tree) GetRootByHash(ctx context.Context, hash common.Hash) (*types.Root, error) {
	var root *types.Root
	if err := meddler.QueryRow(
		t.db, root,
		fmt.Sprintf(`SELECT * FROM %s WHERE hash = $1;`, t.rootTable),
		hash.Hex(),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrNotFound
		}
		return nil, err
	}
	return root, nil
}

func (t *Tree) GetLeaf(ctx context.Context, index uint32, root common.Hash) (common.Hash, error) {
	currentNodeHash := root
	for h := int(types.DefaultHeight - 1); h >= 0; h-- {
		currentNode, err := t.getRHTNode(t.db, currentNodeHash)
		if err != nil {
			return common.Hash{}, err
		}
		if index&(1<<h) > 0 {
			currentNodeHash = currentNode.Right
		} else {
			currentNodeHash = currentNode.Left
		}
	}

	return currentNodeHash, nil
}

// Reorg deletes all the data relevant from firstReorgedBlock (includded) and onwards
func (t *Tree) Reorg(tx db.Txer, firstReorgedBlock uint64) error {
	_, err := tx.Exec(
		fmt.Sprintf(`DELETE FROM %s WHERE block_num >= $1`, t.rootTable),
		firstReorgedBlock,
	)
	return err
	// NOTE: rht is not cleaned, this could be done in the future as optimization
}
