package tree

import (
	"database/sql"
	"errors"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
)

// UpdatableTree is a tree that have updatable leaves, and doesn't need to have sequential inserts
type UpdatableTree struct {
	*Tree
}

// NewUpdatableTree returns an UpdatableTree
func NewUpdatableTree(db *sql.DB, dbPrefix string) *UpdatableTree {
	t := newTree(db, dbPrefix)
	ut := &UpdatableTree{
		Tree: t,
	}
	return ut
}

func (t *UpdatableTree) UpsertLeaf(tx db.Txer, blockNum, blockPosition uint64, leaf types.Leaf) error {
	var rootHash common.Hash
	root, err := t.getLastRootWithTx(tx)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			rootHash = t.zeroHashes[types.DefaultHeight]
		} else {
			return err
		}
	} else {
		rootHash = root.Hash
	}
	siblings, _, err := t.getSiblings(tx, leaf.Index, rootHash)
	if err != nil {
		return err
	}
	currentChildHash := leaf.Hash
	newNodes := []types.TreeNode{}
	for h := uint8(0); h < types.DefaultHeight; h++ {
		var parent types.TreeNode
		if leaf.Index&(1<<h) > 0 {
			// Add child to the right
			parent = newTreeNode(siblings[h], currentChildHash)
		} else {
			// Add child to the left
			parent = newTreeNode(currentChildHash, siblings[h])
		}
		currentChildHash = parent.Hash
		newNodes = append(newNodes, parent)
	}
	if err := t.storeRoot(tx, types.Root{
		Hash:          currentChildHash,
		Index:         leaf.Index,
		BlockNum:      blockNum,
		BlockPosition: blockPosition,
	}); err != nil {
		return err
	}
	if err := t.storeNodes(tx, newNodes); err != nil {
		return err
	}
	return nil
}
