package tree

import (
	"context"
	"fmt"
	"math"

	dbCommon "github.com/0xPolygon/cdk/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

type AppendOnlyTree struct {
	*Tree
	lastLeftCache []common.Hash
	lastIndex     int64
}

func NewAppendOnly(ctx context.Context, dbPath, dbPrefix string) (*AppendOnlyTree, error) {
	t, err := newTree(dbPath, dbPrefix)
	if err != nil {
		return nil, err
	}
	at := &AppendOnlyTree{Tree: t}
	if err := at.initLastLeftCacheAndLastDepositCount(ctx); err != nil {
		return nil, err
	}
	return at, nil
}

// AddLeaves adds a list leaves into the tree
func (t *AppendOnlyTree) AddLeaves(ctx context.Context, leaves []Leaf) error {
	// Sanity check
	if len(leaves) == 0 {
		return nil
	}
	if int64(leaves[0].Index) != t.lastIndex+1 {
		return fmt.Errorf(
			"mismatched index. Expected: %d, actual: %d",
			t.lastIndex+1, leaves[0].Index,
		)
	}
	tx, err := t.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	backupIndx := t.lastIndex
	backupCache := make([]common.Hash, len(t.lastLeftCache))
	copy(backupCache, t.lastLeftCache)

	for _, leaf := range leaves {
		if err := t.addLeaf(tx, leaf); err != nil {
			tx.Rollback()
			t.lastIndex = backupIndx
			t.lastLeftCache = backupCache
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		t.lastIndex = backupIndx
		t.lastLeftCache = backupCache
		return err
	}
	return nil
}

func (t *AppendOnlyTree) addLeaf(tx kv.RwTx, leaf Leaf) error {
	// Calculate new tree nodes
	currentChildHash := leaf.Hash
	newNodes := []treeNode{}
	for h := uint8(0); h < t.height; h++ {
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
			// TODO: review this part of the logic, skipping ?optimizaton?
			// from OG implementation
			t.lastLeftCache[h] = currentChildHash
		}
		currentChildHash = parent.hash()
		newNodes = append(newNodes, parent)
	}

	// store root
	root := currentChildHash
	if err := tx.Put(t.rootTable, dbCommon.Uint32ToBytes(leaf.Index), root[:]); err != nil {
		return err
	}

	// store nodes
	for _, node := range newNodes {
		value, err := node.MarshalBinary()
		if err != nil {
			return err
		}
		if err := tx.Put(t.rhtTable, node.hash().Bytes(), value); err != nil {
			return err
		}
	}

	t.lastIndex++
	return nil
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

// getLastIndexAndRoot return the index and the root associated to the last leaf inserted.
// If index == -1, it means no leaf added yet
func (t *AppendOnlyTree) getLastIndexAndRoot(tx kv.Tx) (int64, common.Hash, error) {
	iter, err := tx.RangeDescend(
		t.rootTable,
		dbCommon.Uint32ToBytes(math.MaxUint32),
		dbCommon.Uint32ToBytes(0),
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
	return int64(dbCommon.BytesToUint32(lastIndexBytes)), common.Hash(rootBytes), nil
}

func (t *AppendOnlyTree) initLastIndex(tx kv.Tx) (common.Hash, error) {
	ldc, root, err := t.getLastIndexAndRoot(tx)
	if err != nil {
		return common.Hash{}, err
	}
	t.lastIndex = ldc
	return root, nil
}
func (t *AppendOnlyTree) initLastLeftCache(tx kv.Tx, lastIndex int64, lastRoot common.Hash) error {
	siblings := make([]common.Hash, t.height, t.height)
	if lastIndex == -1 {
		t.lastLeftCache = siblings
		return nil
	}
	index := lastIndex

	currentNodeHash := lastRoot
	// It starts in height-1 because 0 is the level of the leafs
	for h := int(t.height - 1); h >= 0; h-- {
		currentNode, err := t.getRHTNode(tx, currentNodeHash)
		if err != nil {
			return fmt.Errorf(
				"error getting node %s from the RHT at height %d with root %s: %v",
				currentNodeHash.Hex(), h, lastRoot.Hex(), err,
			)
		}
		if currentNode == nil {
			return ErrNotFound
		}
		siblings = append(siblings, currentNode.left)
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
func (t *AppendOnlyTree) Reorg(ctx context.Context, firstReorgedIndex uint32) error {
	if t.lastIndex == -1 {
		return nil
	}
	tx, err := t.db.BeginRw(ctx)
	if err != nil {
		return err
	}
	// Clean root table
	for i := firstReorgedIndex; i <= uint32(t.lastIndex); i++ {
		if err := tx.Delete(t.rootTable, dbCommon.Uint32ToBytes(i)); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Reset
	root := common.Hash{}
	if firstReorgedIndex > 0 {
		rootBytes, err := tx.GetOne(t.rootTable, dbCommon.Uint32ToBytes(firstReorgedIndex-1))
		if err != nil {
			tx.Rollback()
			return err
		}
		if rootBytes == nil {
			tx.Rollback()
			return ErrNotFound
		}
		root = common.Hash(rootBytes)
	}
	err = t.initLastLeftCache(tx, int64(firstReorgedIndex)-1, root)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Note: not cleaning RHT, not worth it
	if err := tx.Commit(); err != nil {
		return err
	}
	t.lastIndex = int64(firstReorgedIndex) - 1
	return nil
}
