package tree

import (
	"context"
	"errors"
	"fmt"
	"math"

	dbCommon "github.com/0xPolygon/cdk/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"golang.org/x/crypto/sha3"
)

const (
	DefaultHeight   uint8 = 32
	rootTableSufix        = "-root"
	rhtTableSufix         = "-rht"
	indexTableSufix       = "-index"
)

var (
	EmptyProof  = [32]common.Hash{}
	ErrNotFound = errors.New("not found")
)

type Leaf struct {
	Index uint32
	Hash  common.Hash
}

type Tree struct {
	db         kv.RwDB
	rhtTable   string
	rootTable  string
	indexTable string
	zeroHashes []common.Hash
}

type treeNode struct {
	left  common.Hash
	right common.Hash
}

func (n *treeNode) hash() common.Hash {
	var hash common.Hash
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(n.left[:])
	hasher.Write(n.right[:])
	copy(hash[:], hasher.Sum(nil))
	return hash
}

func (n *treeNode) MarshalBinary() ([]byte, error) {
	return append(n.left[:], n.right[:]...), nil
}

func (n *treeNode) UnmarshalBinary(data []byte) error {
	const nodeDataLength = 64
	if len(data) != nodeDataLength {
		return fmt.Errorf("expected len %d, actual len %d", nodeDataLength, len(data))
	}
	n.left = common.Hash(data[:32])
	n.right = common.Hash(data[32:])
	return nil
}

// AddTables add the needed tables for the tree to work in a tableCfg
func AddTables(tableCfg map[string]kv.TableCfgItem, dbPrefix string) {
	rootTable := dbPrefix + rootTableSufix
	rhtTable := dbPrefix + rhtTableSufix
	indexTable := dbPrefix + indexTableSufix
	tableCfg[rootTable] = kv.TableCfgItem{}
	tableCfg[rhtTable] = kv.TableCfgItem{}
	tableCfg[indexTable] = kv.TableCfgItem{}
}

func newTree(db kv.RwDB, dbPrefix string) *Tree {
	rootTable := dbPrefix + rootTableSufix
	rhtTable := dbPrefix + rhtTableSufix
	indexTable := dbPrefix + indexTableSufix
	t := &Tree{
		rhtTable:   rhtTable,
		rootTable:  rootTable,
		indexTable: indexTable,
		db:         db,
		zeroHashes: generateZeroHashes(DefaultHeight),
	}

	return t
}

func (t *Tree) getRootByIndex(tx kv.Tx, index uint64) (common.Hash, error) {
	rootBytes, err := tx.GetOne(t.rootTable, dbCommon.Uint64ToBytes(index))
	if err != nil {
		return common.Hash{}, err
	}
	if rootBytes == nil {
		return common.Hash{}, ErrNotFound
	}
	return common.BytesToHash(rootBytes), nil
}

func (t *Tree) getIndexByRoot(tx kv.Tx, root common.Hash) (uint64, error) {
	indexBytes, err := tx.GetOne(t.indexTable, root[:])
	if err != nil {
		return 0, err
	}
	if indexBytes == nil {
		return 0, ErrNotFound
	}
	return dbCommon.BytesToUint64(indexBytes), nil
}

func (t *Tree) getSiblings(tx kv.Tx, index uint32, root common.Hash) (
	siblings [32]common.Hash,
	hasUsedZeroHashes bool,
	err error,
) {
	currentNodeHash := root
	// It starts in height-1 because 0 is the level of the leafs
	for h := int(DefaultHeight - 1); h >= 0; h-- {
		var currentNode *treeNode
		currentNode, err = t.getRHTNode(tx, currentNodeHash)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
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
			siblings[h] = currentNode.left
			currentNodeHash = currentNode.right
		} else {
			siblings[h] = currentNode.right
			currentNodeHash = currentNode.left
		}
	}

	return
}

// GetProof returns the merkle proof for a given index and root.
func (t *Tree) GetProof(ctx context.Context, index uint32, root common.Hash) ([DefaultHeight]common.Hash, error) {
	tx, err := t.db.BeginRw(ctx)
	if err != nil {
		return [DefaultHeight]common.Hash{}, err
	}
	defer tx.Rollback()
	siblings, isErrNotFound, err := t.getSiblings(tx, index, root)
	if err != nil {
		return [DefaultHeight]common.Hash{}, err
	}
	if isErrNotFound {
		return [DefaultHeight]common.Hash{}, ErrNotFound
	}
	return siblings, nil
}

func (t *Tree) getRHTNode(tx kv.Tx, nodeHash common.Hash) (*treeNode, error) {
	nodeBytes, err := tx.GetOne(t.rhtTable, nodeHash[:])
	if err != nil {
		return nil, err
	}
	if nodeBytes == nil {
		return nil, ErrNotFound
	}
	node := &treeNode{}
	err = node.UnmarshalBinary(nodeBytes)
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

func (t *Tree) storeNodes(tx kv.RwTx, nodes []treeNode) error {
	for _, node := range nodes {
		value, err := node.MarshalBinary()
		if err != nil {
			return err
		}
		if err := tx.Put(t.rhtTable, node.hash().Bytes(), value); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tree) storeRoot(tx kv.RwTx, rootIndex uint64, root common.Hash) error {
	if err := tx.Put(t.rootTable, dbCommon.Uint64ToBytes(rootIndex), root[:]); err != nil {
		return err
	}
	return tx.Put(t.indexTable, root[:], dbCommon.Uint64ToBytes(rootIndex))
}

// GetLastRoot returns the last processed root
func (t *Tree) GetLastRoot(ctx context.Context) (common.Hash, error) {
	tx, err := t.db.BeginRo(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	defer tx.Rollback()

	i, root, err := t.getLastIndexAndRootWithTx(tx)
	if err != nil {
		return common.Hash{}, err
	}
	if i == -1 {
		return common.Hash{}, ErrNotFound
	}
	return root, nil
}

// getLastIndexAndRootWithTx return the index and the root associated to the last leaf inserted.
// If index == -1, it means no leaf added yet
func (t *Tree) getLastIndexAndRootWithTx(tx kv.Tx) (int64, common.Hash, error) {
	iter, err := tx.RangeDescend(
		t.rootTable,
		dbCommon.Uint64ToBytes(math.MaxUint64),
		dbCommon.Uint64ToBytes(0),
		1,
	)
	if err != nil {
		return 0, common.Hash{}, err
	}

	lastIndexBytes, rootBytes, err := iter.Next()
	if err != nil {
		return 0, common.Hash{}, err
	}
	if lastIndexBytes == nil {
		return -1, common.Hash{}, nil
	}
	return int64(dbCommon.BytesToUint64(lastIndexBytes)), common.Hash(rootBytes), nil
}

func (t *Tree) GetLeaf(ctx context.Context, index uint32, root common.Hash) (common.Hash, error) {
	tx, err := t.db.BeginRo(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	defer tx.Rollback()

	currentNodeHash := root
	for h := int(DefaultHeight - 1); h >= 0; h-- {
		currentNode, err := t.getRHTNode(tx, currentNodeHash)
		if err != nil {
			return common.Hash{}, err
		}
		if index&(1<<h) > 0 {
			currentNodeHash = currentNode.right
		} else {
			currentNodeHash = currentNode.left
		}
	}

	return currentNodeHash, nil
}
