package l1infotreesync

import (
	"testing"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestGetVerifiedBatches(t *testing.T) {
	dbPath := "file:TestGetVerifiedBatches?mode=memory&cache=shared"
	p, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.Background()

	// Test ErrNotFound returned correctly on all methods
	_, err = p.GetLastVerifiedBatches(0)
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetFirstVerifiedBatches(0)
	require.Equal(t, db.ErrNotFound, err)
	_, err = p.GetFirstVerifiedBatchesAfterBlock(0, 0)
	require.Equal(t, db.ErrNotFound, err)

	// First insert
	expected1 := &VerifyBatches{
		RollupID:   420,
		NumBatch:   69,
		StateRoot:  common.HexToHash("5ca1e"),
		ExitRoot:   common.HexToHash("b455"),
		Aggregator: common.HexToAddress("beef"),
	}
	err = p.ProcessBlock(ctx, sync.Block{
		Num: 1,
		Events: []interface{}{
			Event{VerifyBatches: expected1},
		},
	})
	require.NoError(t, err)
	_, err = p.GetLastVerifiedBatches(0)
	require.Equal(t, db.ErrNotFound, err)
	actual, err := p.GetLastVerifiedBatches(420)
	require.NoError(t, err)
	require.Equal(t, expected1, actual)
	actual, err = p.GetFirstVerifiedBatches(420)
	require.NoError(t, err)
	require.Equal(t, expected1, actual)

	// Second insert
	expected2 := &VerifyBatches{
		RollupID:   420,
		NumBatch:   690,
		StateRoot:  common.HexToHash("5ca1e3"),
		ExitRoot:   common.HexToHash("ba55"),
		Aggregator: common.HexToAddress("beef3"),
	}
	err = p.ProcessBlock(ctx, sync.Block{
		Num: 2,
		Events: []interface{}{
			Event{VerifyBatches: expected2},
		},
	})
	require.NoError(t, err)
	_, err = p.GetLastVerifiedBatches(0)
	require.Equal(t, db.ErrNotFound, err)
	actual, err = p.GetLastVerifiedBatches(420)
	require.NoError(t, err)
	require.Equal(t, expected2, actual)
	actual, err = p.GetFirstVerifiedBatches(420)
	require.NoError(t, err)
	require.Equal(t, expected1, actual)
	actual, err = p.GetFirstVerifiedBatchesAfterBlock(420, 2)
	require.NoError(t, err)
	require.Equal(t, expected2, actual)
}

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

func TestInitL1InfoRootMap(t *testing.T) {
	dbPath := "file:TestGetVerifiedBatches?mode=memory&cache=shared"
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.TODO()
	event := InitL1InfoRootMap{
		LeafCount:         1,
		CurrentL1InfoRoot: common.HexToHash("beef"),
	}
	block := sync.Block{
		Num: 1,
		Events: []interface{}{
			Event{InitL1InfoRootMap: &event},
		},
	}

	err = sut.ProcessBlock(ctx, block)
	require.NoError(t, err)

	info, err := sut.GetInitL1InfoRootMap(nil)
	require.NoError(t, err)
	require.NotNil(t, info)
	require.Equal(t, event.LeafCount, info.LeafCount)
	require.Equal(t, event.CurrentL1InfoRoot, info.L1InfoRoot)
	require.Equal(t, block.Num, info.BlockNumber)

}

func TestInitL1InfoRootMapDontAllow2Rows(t *testing.T) {
	dbPath := "file:test?mode=memory&cache=shared"
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	ctx := context.TODO()
	block := sync.Block{
		Num: 1,
		Events: []interface{}{
			Event{InitL1InfoRootMap: &InitL1InfoRootMap{
				LeafCount:         1,
				CurrentL1InfoRoot: common.HexToHash("beef"),
			}},
		},
	}
	err = sut.ProcessBlock(ctx, block)
	require.NoError(t, err)
	block.Num = 2
	err = sut.ProcessBlock(ctx, block)
	require.Error(t, err, "should not allow to insert a second row")
}

func TestGetInitL1InfoRootMap(t *testing.T) {
	dbPath := "file:test?mode=memory&cache=shared"
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	info, err := sut.GetInitL1InfoRootMap(nil)
	require.NoError(t, err, "should return no error if no row is present, because it returns data=nil")
	require.Nil(t, info, "should return nil if no row is present")
}
