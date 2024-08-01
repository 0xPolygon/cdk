package tree

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"golang.org/x/crypto/sha3"
)

const (
	defaultHeight  uint8 = 32
	rootTableSufix       = "-root"
	rhtTableSufix        = "-rht"
)

var (
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
	height     uint8
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
	if len(data) != 64 {
		return fmt.Errorf("expected len %d, actual len %d", 64, len(data))
	}
	n.left = common.Hash(data[:32])
	n.right = common.Hash(data[32:])
	return nil
}

func AddTables(tableCfg map[string]kv.TableCfgItem, dbPrefix string) {
	rootTable := dbPrefix + rootTableSufix
	rhtTable := dbPrefix + rhtTableSufix
	tableCfg[rootTable] = kv.TableCfgItem{}
	tableCfg[rhtTable] = kv.TableCfgItem{}
}

func newTree(db kv.RwDB, dbPrefix string) *Tree {
	rootTable := dbPrefix + rootTableSufix
	rhtTable := dbPrefix + rhtTableSufix
	t := &Tree{
		rhtTable:   rhtTable,
		rootTable:  rootTable,
		db:         db,
		height:     defaultHeight,
		zeroHashes: generateZeroHashes(defaultHeight),
	}

	return t
}

// GetProof returns the merkle proof for a given index and root.
func (t *Tree) GetProof(ctx context.Context, index uint32, root common.Hash) ([]common.Hash, error) {
	tx, err := t.db.BeginRw(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	siblings := make([]common.Hash, int(t.height))

	currentNodeHash := root
	// It starts in height-1 because 0 is the level of the leafs
	for h := int(t.height - 1); h >= 0; h-- {
		currentNode, err := t.getRHTNode(tx, currentNodeHash)
		if err != nil {
			return nil, fmt.Errorf(
				"height: %d, currentNode: %s, error: %v",
				h, currentNodeHash.Hex(), err,
			)
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
			siblings = append(siblings, currentNode.left)
			currentNodeHash = currentNode.right
		} else {
			siblings = append(siblings, currentNode.right)
			currentNodeHash = currentNode.left
		}
	}

	// Reverse siblings to go from leafs to root
	for i, j := 0, len(siblings)-1; i < j; i, j = i+1, j-1 {
		siblings[i], siblings[j] = siblings[j], siblings[i]
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
	// This generates a leaf = HashZero in position 0. In the rest of the positions that are equivalent to the ascending levels,
	// we set the hashes of the nodes. So all nodes from level i=5 will have the same value and same children nodes.
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
