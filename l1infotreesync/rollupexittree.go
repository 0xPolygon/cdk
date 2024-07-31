package l1infotreesync

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"golang.org/x/crypto/sha3"
)

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

type rollupExitTree struct {
	height           uint8
	rhtTable         string
	lastExitTreeRoot common.Hash
}

func newRollupExitTree() *rollupExitTree {
	return &rollupExitTree{}
}

func (t *rollupExitTree) addLeaf(
	tx kv.RwTx,
	rollupID uint32,
	rollupExitRoot common.Hash,
	expectedRollupExitRoot *common.Hash,
) error {
	siblings, err := t.getProof(tx, rollupID, t.lastExitTreeRoot)
	if err != nil {
		return err
	}
	if expectedRollupExitRoot != nil && *expectedRollupExitRoot != t.lastExitTreeRoot {
		return fmt.Errorf(
			"expectedRollupExitRoot: %s, actual: %s",
			expectedRollupExitRoot.Hex(), t.lastExitTreeRoot.Hex(),
		)
	}
	return nil
}

func (t *rollupExitTree) getSiblings(tx kv.RwTx, rollupID uint32, root common.Hash) (bool, []common.Hash, error) {
	siblings := make([]common.Hash, int(t.height))

	currentNodeHash := root
	// It starts in height-1 because 0 is the level of the leafs
	for h := int(t.height - 1); h >= 0; h-- {
		currentNode, err := t.getRHTNode(tx, currentNodeHash)
		if err != nil {
			// handle not found for inserts and shit
			return false, nil, fmt.Errorf(
				"height: %d, currentNode: %s, error: %v",
				h, currentNodeHash.Hex(), err,
			)
		}
		if rollupID&(1<<h) > 0 {
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

	return false, siblings, nil

}

// getProof returns the merkle proof for a given deposit count and root.
func (t *rollupExitTree) getProof(tx kv.RwTx, rollupID uint32, root common.Hash) ([]common.Hash, error) {
	usedEmptyTree, siblings, err := t.getSiblings(tx, rollupID, root)
	if usedEmptyTree {
		return nil, ErrNotFound
	}
	return siblings, err
}

func (t *rollupExitTree) getRHTNode(tx kv.Tx, nodeHash common.Hash) (*treeNode, error) {
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
