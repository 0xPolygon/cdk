package tree

import (
	"context"
	"fmt"

	dbCommon "github.com/0xPolygon/cdk/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

// AppendOnlyTree is a tree where leaves are added sequentially (by index)
type AppendOnlyTree struct {
	*Tree
	lastLeftCache [DefaultHeight]common.Hash
	lastIndex     int64
}

// NewAppendOnlyTree creates a AppendOnlyTree
func NewAppendOnlyTree(ctx context.Context, db kv.RwDB, dbPrefix string) (*AppendOnlyTree, error) {
	t := newTree(db, dbPrefix)
	at := &AppendOnlyTree{Tree: t}
	if err := at.initLastLeftCacheAndLastDepositCount(ctx); err != nil {
		return nil, err
	}
	return at, nil
}

// AddLeaves adds a list leaves into the tree. The indexes of the leaves must be consecutive,
// starting by the index of the last leaf added +1
// It returns a function that must be called to rollback the changes done by this interaction
func (t *AppendOnlyTree) AddLeaves(tx kv.RwTx, leaves []Leaf) (func(), error) {
	// Sanity check
	if len(leaves) == 0 {
		return func() {}, nil
	}

	backupIndx := t.lastIndex
	backupCache := [DefaultHeight]common.Hash{}
	copy(backupCache[:], t.lastLeftCache[:])
	rollback := func() {
		t.lastIndex = backupIndx
		t.lastLeftCache = backupCache
	}

	for _, leaf := range leaves {
		if err := t.addLeaf(tx, leaf); err != nil {
			return rollback, err
		}
	}

	return rollback, nil
}

func (t *AppendOnlyTree) addLeaf(tx kv.RwTx, leaf Leaf) error {
	if int64(leaf.Index) != t.lastIndex+1 {
		return fmt.Errorf(
			"mismatched index. Expected: %d, actual: %d",
			t.lastIndex+1, leaf.Index,
		)
	}
	// Calculate new tree nodes
	currentChildHash := leaf.Hash
	newNodes := []treeNode{}
	for h := uint8(0); h < DefaultHeight; h++ {
		var parent treeNode
		if leaf.Index&(1<<h) > 0 {
			// Add child to the right
			parent = treeNode{
				left:  t.lastLeftCache[h],
				right: currentChildHash,
			}
		} else {
			// Add child to the left
			parent = treeNode{
				left:  currentChildHash,
				right: t.zeroHashes[h],
			}
			// Update cache
			// TODO: review this part of the logic, skipping? optimisation?
			// from OG implementation
			t.lastLeftCache[h] = currentChildHash
		}
		currentChildHash = parent.hash()
		newNodes = append(newNodes, parent)
	}

	// store root
	if err := t.storeRoot(tx, uint64(leaf.Index), currentChildHash); err != nil {
		return fmt.Errorf("failed to store root: %w", err)
	}
	root := currentChildHash
	if err := tx.Put(t.rootTable, dbCommon.Uint64ToBytes(uint64(leaf.Index)), root[:]); err != nil {
		return err
	}
	// store nodes
	if err := t.storeNodes(tx, newNodes); err != nil {
		return err
	}
	t.lastIndex++
	return nil
}

// GetRootByIndex returns the root of the tree as it was right after adding the leaf with index
func (t *AppendOnlyTree) GetRootByIndex(tx kv.Tx, index uint32) (common.Hash, error) {
	return t.getRootByIndex(tx, uint64(index))
}

func (t *AppendOnlyTree) GetIndexByRoot(ctx context.Context, root common.Hash) (uint32, error) {
	tx, err := t.db.BeginRo(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	index, err := t.getIndexByRoot(tx, root)
	return uint32(index), err
}

// GetLastIndexAndRoot returns the last index and root added to the tree
func (t *AppendOnlyTree) GetLastIndexAndRoot(ctx context.Context) (uint32, common.Hash, error) {
	tx, err := t.db.BeginRo(ctx)
	if err != nil {
		return 0, common.Hash{}, err
	}
	defer tx.Rollback()
	i, root, err := t.getLastIndexAndRootWithTx(tx)
	if err != nil {
		return 0, common.Hash{}, err
	}
	if i == -1 {
		return 0, common.Hash{}, ErrNotFound
	}
	return uint32(i), root, nil
}

func (t *AppendOnlyTree) initLastLeftCacheAndLastDepositCount(ctx context.Context) error {
	tx, err := t.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	root, err := t.initLastIndex(tx)
	if err != nil {
		return err
	}
	return t.initLastLeftCache(tx, t.lastIndex, root)
}

func (t *AppendOnlyTree) initLastIndex(tx kv.Tx) (common.Hash, error) {
	lastIndex, root, err := t.getLastIndexAndRootWithTx(tx)
	if err != nil {
		return common.Hash{}, err
	}
	t.lastIndex = lastIndex
	return root, nil
}

func (t *AppendOnlyTree) initLastLeftCache(tx kv.Tx, lastIndex int64, lastRoot common.Hash) error {
	siblings := [DefaultHeight]common.Hash{}
	if lastIndex == -1 {
		t.lastLeftCache = siblings
		return nil
	}
	index := lastIndex

	currentNodeHash := lastRoot
	// It starts in height-1 because 0 is the level of the leafs
	for h := int(DefaultHeight - 1); h >= 0; h-- {
		currentNode, err := t.getRHTNode(tx, currentNodeHash)
		if err != nil {
			return fmt.Errorf(
				"error getting node %s from the RHT at height %d with root %s: %w",
				currentNodeHash.Hex(), h, lastRoot.Hex(), err,
			)
		}
		if currentNode == nil {
			return ErrNotFound
		}
		siblings[h] = currentNode.left
		if index&(1<<h) > 0 {
			currentNodeHash = currentNode.right
		} else {
			currentNodeHash = currentNode.left
		}
	}

	// Reverse the siblings to go from leafs to root
	for i, j := 0, len(siblings)-1; i < j; i, j = i+1, j-1 {
		siblings[i], siblings[j] = siblings[j], siblings[i]
	}

	t.lastLeftCache = siblings
	return nil
}

// Reorg deletes all the data relevant from firstReorgedIndex (includded) and onwards
// and prepares the tree tfor being used as it was at firstReorgedIndex-1
// It returns a function that must be called to rollback the changes done by this interaction
func (t *AppendOnlyTree) Reorg(tx kv.RwTx, firstReorgedIndex uint32) (func(), error) {
	if t.lastIndex == -1 {
		return func() {}, nil
	}
	// Clean root table
	for i := firstReorgedIndex; i <= uint32(t.lastIndex); i++ {
		if err := tx.Delete(t.rootTable, dbCommon.Uint64ToBytes(uint64(i))); err != nil {
			return func() {}, err
		}
	}

	// Reset
	root := common.Hash{}
	if firstReorgedIndex > 0 {
		rootBytes, err := tx.GetOne(t.rootTable, dbCommon.Uint64ToBytes(uint64(firstReorgedIndex)-1))
		if err != nil {
			return func() {}, err
		}
		if rootBytes == nil {
			return func() {}, ErrNotFound
		}
		root = common.Hash(rootBytes)
	}
	err := t.initLastLeftCache(tx, int64(firstReorgedIndex)-1, root)
	if err != nil {
		return func() {}, err
	}

	// Note: not cleaning RHT, not worth it
	backupLastIndex := t.lastIndex
	t.lastIndex = int64(firstReorgedIndex) - 1
	return func() {
		t.lastIndex = backupLastIndex
	}, nil
}
