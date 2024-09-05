package tree

import (
	"database/sql"
	"fmt"

	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
)

// AppendOnlyTree is a tree where leaves are added sequentially (by index)
type AppendOnlyTree struct {
	*Tree
	lastLeftCache types.Proof
	lastIndex     int64
}

// NewAppendOnlyTree creates a AppendOnlyTree
func NewAppendOnlyTree(db *sql.DB, dbPrefix string) *AppendOnlyTree {
	t := newTree(db, dbPrefix)
	return &AppendOnlyTree{
		Tree:      t,
		lastIndex: -2,
	}
}

	if int64(leaf.Index) != t.lastIndex+1 {
			t.lastLeftCache[h] = currentChildHash
		}
		currentChildHash = parent.Hash
		newNodes = append(newNodes, parent)
	}

	// store root
	if err := t.storeRoot(tx, types.Root{
		Hash:          currentChildHash,
		Index:         leaf.Index,
		BlockNum:      blockNum,
		BlockPosition: blockPosition,
	}); err != nil {
		return err
	}

	// store nodes
	if err := t.storeNodes(tx, newNodes); err != nil {
		return err
	}
	t.lastIndex++
	return nil
}

func (t *AppendOnlyTree) initCache(tx *sql.Tx) error {
	siblings := [types.DefaultHeight]common.Hash{}
	lastRoot, err := t.getLastRootWithTx(tx)
	if err != nil {
		if err == ErrNotFound {
			t.lastIndex = -1
			t.lastLeftCache = siblings
			return nil
		}
		return err
	}
	t.lastIndex = int64(lastRoot.Index)
	currentNodeHash := lastRoot.Hash
	index := t.lastIndex
	// It starts in height-1 because 0 is the level of the leafs
	for h := int(types.DefaultHeight - 1); h >= 0; h-- {
		currentNode, err := t.getRHTNode(tx, currentNodeHash)
		if err != nil {
			return fmt.Errorf(
				"error getting node %s from the RHT at height %d with root %s: %v",
				currentNodeHash.Hex(), h, lastRoot.Hash.Hex(), err,
			)
		}
		if currentNode == nil {
			return ErrNotFound
		}
		siblings[h] = currentNode.Left
		if index&(1<<h) > 0 {
			currentNodeHash = currentNode.Right
		} else {
			currentNodeHash = currentNode.Left
		}
	}

	// Reverse the siblings to go from leafs to root
	for i, j := 0, len(siblings)-1; i < j; i, j = i+1, j-1 {
		siblings[i], siblings[j] = siblings[j], siblings[i]
	}

	t.lastLeftCache = siblings
	return nil
}
