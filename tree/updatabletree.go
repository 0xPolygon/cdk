package tree

import (
	"context"
	"math"

	dbCommon "github.com/0xPolygon/cdk/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

// UpdatableTree is a tree that have updatable leaves, and doesn't need to have sequential inserts
type UpdatableTree struct {
	*Tree
	lastRoot common.Hash
}

// NewUpdatableTree returns an UpdatableTree
func NewUpdatableTree(ctx context.Context, db kv.RwDB, dbPrefix string) (*UpdatableTree, error) {
	// TODO: Load last root
	t := newTree(db, dbPrefix)
	tx, err := t.db.BeginRw(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rootIndex, root, err := t.getLastIndexAndRootWithTx(tx)
	if err != nil {
		return nil, err
	}
	if rootIndex == -1 {
		root = t.zeroHashes[DefaultHeight]
	}
	ut := &UpdatableTree{
		Tree:     t,
		lastRoot: root,
	}
	return ut, nil
}

// UpseartLeaves inserts or updates a list of leaves. The root index will be used to index the resulting
// root after performing all the operations. Root index must be greater than the last used root index,
// but doesn't need to be sequential. Great for relating block nums and roots :)
// It returns a function that must be called to rollback the changes done by this interaction
func (t *UpdatableTree) UpseartLeaves(tx kv.RwTx, leaves []Leaf, rootIndex uint64) (func(), error) {
	if len(leaves) == 0 {
		return func() {}, nil
	}
	rootBackup := t.lastRoot
	rollback := func() {
		t.lastRoot = rootBackup
	}

	for _, l := range leaves {
		if err := t.upsertLeaf(tx, l); err != nil {
			return rollback, err
		}
	}

	if err := t.storeRoot(tx, rootIndex, t.lastRoot); err != nil {
		return rollback, err
	}
	return rollback, nil
}

func (t *UpdatableTree) upsertLeaf(tx kv.RwTx, leaf Leaf) error {
	siblings, _, err := t.getSiblings(tx, leaf.Index, t.lastRoot)
	if err != nil {
		return err
	}
	currentChildHash := leaf.Hash
	newNodes := []treeNode{}
	for h := uint8(0); h < DefaultHeight; h++ {
		var parent treeNode
		if leaf.Index&(1<<h) > 0 {
			// Add child to the right
			parent = treeNode{
				left:  siblings[h],
				right: currentChildHash,
			}
		} else {
			// Add child to the left
			parent = treeNode{
				left:  currentChildHash,
				right: siblings[h],
			}
		}
		currentChildHash = parent.hash()
		newNodes = append(newNodes, parent)
	}

	if err := t.storeNodes(tx, newNodes); err != nil {
		return err
	}
	t.lastRoot = currentChildHash
	return nil
}

// Reorg deletes all the data relevant from firstReorgedIndex (includded) and onwards
// and prepares the tree tfor being used as it was at firstReorgedIndex-1.
// It returns a function that must be called to rollback the changes done by this interaction
func (t *UpdatableTree) Reorg(tx kv.RwTx, firstReorgedIndex uint64) (func(), error) {
	iter, err := tx.RangeDescend(
		t.rootTable,
		dbCommon.Uint64ToBytes(math.MaxUint64),
		dbCommon.Uint64ToBytes(0),
		0,
	)
	if err != nil {
		return func() {}, err
	}
	rootBackup := t.lastRoot
	rollback := func() {
		t.lastRoot = rootBackup
	}

	for lastIndexBytes, rootBytes, err := iter.Next(); lastIndexBytes != nil; lastIndexBytes, rootBytes, err = iter.Next() { //nolint:lll
		if err != nil {
			return rollback, err
		}

		if dbCommon.BytesToUint64(lastIndexBytes) >= firstReorgedIndex {
			if err := tx.Delete(t.rootTable, lastIndexBytes); err != nil {
				return rollback, err
			}
		} else {
			t.lastRoot = common.Hash(rootBytes)
			return rollback, nil
		}
	}

	// no root found after reorg, going back to empty tree
	t.lastRoot = t.zeroHashes[DefaultHeight]
	return rollback, nil
}

// GetRootByRootIndex returns the root of the tree as it was right after adding the leaf with index
func (t *UpdatableTree) GetRootByRootIndex(tx kv.Tx, rootIndex uint64) (common.Hash, error) {
	return t.getRootByIndex(tx, rootIndex)
}
