package l1infotreesync

import (
	"testing"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestGetInfo(t *testing.T) {
	dbPath := "file:TestGetInfo?mode=memory&cache=shared"
	p, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.Background()

	// Test ErrNotFound returned correctly on all methods
	_, err = p.GetFirstL1InfoWithRollupExitRoot(common.Hash{})
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetLastInfo()
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetFirstInfo()
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetFirstInfoAfterBlock(0)
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetInfoByGlobalExitRoot(common.Hash{})
	require.Equal(t, db.ErrNotFound, err)

	// First insert
	info1 := &UpdateL1InfoTree{
		MainnetExitRoot: common.HexToHash("beef"),
		RollupExitRoot:  common.HexToHash("5ca1e"),
		ParentHash:      common.HexToHash("1010101"),
		Timestamp:       420,
	}
	expected1 := L1InfoTreeLeaf{
		BlockNumber:       1,
		L1InfoTreeIndex:   0,
		PreviousBlockHash: info1.ParentHash,
		Timestamp:         info1.Timestamp,
		MainnetExitRoot:   info1.MainnetExitRoot,
		RollupExitRoot:    info1.RollupExitRoot,
	}
	expected1.GlobalExitRoot = expected1.globalExitRoot()
	expected1.Hash = expected1.hash()
	err = p.ProcessBlock(ctx, sync.Block{
		Num: 1,
		Events: []interface{}{
			Event{UpdateL1InfoTree: info1},
		},
	})
	require.NoError(t, err)
	actual, err := p.GetFirstL1InfoWithRollupExitRoot(info1.RollupExitRoot)
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetLastInfo()
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetFirstInfo()
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetFirstInfoAfterBlock(0)
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetInfoByGlobalExitRoot(expected1.GlobalExitRoot)
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)

	// Second insert
	info2 := &UpdateL1InfoTree{
		MainnetExitRoot: common.HexToHash("b055"),
		RollupExitRoot:  common.HexToHash("5ca1e"),
		ParentHash:      common.HexToHash("1010101"),
		Timestamp:       420,
	}
	expected2 := L1InfoTreeLeaf{
		BlockNumber:       2,
		L1InfoTreeIndex:   1,
		PreviousBlockHash: info2.ParentHash,
		Timestamp:         info2.Timestamp,
		MainnetExitRoot:   info2.MainnetExitRoot,
		RollupExitRoot:    info2.RollupExitRoot,
	}
	expected2.GlobalExitRoot = expected2.globalExitRoot()
	expected2.Hash = expected2.hash()
	err = p.ProcessBlock(ctx, sync.Block{
		Num: 2,
		Events: []interface{}{
			Event{UpdateL1InfoTree: info2},
		},
	})
	require.NoError(t, err)
	actual, err = p.GetFirstL1InfoWithRollupExitRoot(info2.RollupExitRoot)
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetLastInfo()
	require.NoError(t, err)
	require.Equal(t, expected2, *actual)
	actual, err = p.GetFirstInfo()
	require.NoError(t, err)
	require.Equal(t, expected1, *actual)
	actual, err = p.GetFirstInfoAfterBlock(2)
	require.NoError(t, err)
	require.Equal(t, expected2, *actual)
	actual, err = p.GetInfoByGlobalExitRoot(expected2.GlobalExitRoot)
	require.NoError(t, err)
	require.Equal(t, expected2, *actual)
}

func TestGetLatestInfoUntilBlockIfNotFoundReturnsErrNotFound(t *testing.T) {
	dbPath := "file:TestGetLatestInfoUntilBlock?mode=memory&cache=shared"
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.Background()
	// Fake block 1
	_, err = sut.db.Exec(`INSERT INTO block (num) VALUES ($1)`, 1)
	require.NoError(t, err)

	_, err = sut.GetLatestInfoUntilBlock(ctx, 1)
	require.Equal(t, db.ErrNotFound, err)
}
