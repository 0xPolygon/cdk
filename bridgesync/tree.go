package bridgesync

import (
	"errors"
	"fmt"

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"golang.org/x/crypto/sha3"
)

type tree struct {
	db             kv.RwDB
	lastIndex      int64
	height         uint8
	rightPathCache []common.Hash
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

func newTree() (*tree, error) {
	// TODO: init lastIndex & rightPathCache
	return &tree{}, errors.New("not implemented")
}

func (t *tree) addLeaf(index uint, hash common.Hash) error {
	if int64(index) != t.lastIndex+1 {
		return fmt.Errorf("mismatched index. Expected: %d, actual: %d", t.lastIndex+1, index)
	}

	currentChildHash := hash
	leftIsFilled := true
	newNodes := []treeNode{}
	for h := uint8(0); h < t.height; h++ {
		var parent treeNode
		if index&(1<<h) > 0 {
			// Add child to the right
			var child common.Hash
			copy(child[:], currentChildHash[:])
			parent = treeNode{
				left:  t.rightPathCache[h],
				right: child,
			}
		} else {
			// Add child to the left
			if leftIsFilled {
				// if at this level the left is filled, it means that the new node will be in the right path
				copy(t.rightPathCache[h][:], currentChildHash[:])
				leftIsFilled = false
			}
			var child common.Hash
			copy(child[:], currentChildHash[:])
			parent = treeNode{
				left:  child,
				right: common.Hash{},
			}
		}
		currentChildHash = parent.hash()
		newNodes = append(newNodes, parent)
	}

	// store root
	root := currentChildHash
	// store nodes

	t.lastIndex++
	return nil
}

// TODO: handle rerog: lastIndex & rightPathCache
