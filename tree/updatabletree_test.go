package tree_test

import (
	"context"
	"testing"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/tree"
	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestUpdatableTreeExploratory(t *testing.T) {
	treeDB := createTreeDBForTest(t)
	sut := tree.NewUpdatableTree(treeDB, "")
	blockNum := uint64(1)
	blockPosition := uint64(1)
	leaf1 := types.Leaf{
		Index: 10,
		Hash:  common.HexToHash("0x123456"),
	}
	leaf2 := types.Leaf{
		Index: 1,
		Hash:  common.HexToHash("0x123478"),
	}
	ctx := context.TODO()

	tx, err := db.NewTx(ctx, treeDB)
	require.NoError(t, err)
	_, err = sut.UpsertLeaf(tx, blockNum, blockPosition, leaf1)
	require.NoError(t, err)

	root2, err := sut.UpsertLeaf(tx, blockNum, blockPosition, leaf2)
	require.NoError(t, err)
	leaf1get, err := sut.GetLeaf(tx, leaf1.Index, root2)
	require.NoError(t, err)
	require.Equal(t, leaf1.Hash, leaf1get)
	// If a leaf dont exist return 'not found' error
	_, err = sut.GetLeaf(tx, 99, root2)
	require.ErrorIs(t, err, db.ErrNotFound)
	leaf99 := types.Leaf{
		Index: 99,
		Hash:  common.Hash{}, // 0x00000
	}

	_, err = sut.UpsertLeaf(tx, blockNum, blockPosition, leaf99)
	require.Error(t, err, "insert 0x000 doesnt change root and return UNIQUE constraint failed: root.hash")

}
