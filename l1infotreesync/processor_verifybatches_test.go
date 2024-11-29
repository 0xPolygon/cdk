package l1infotreesync

import (
	"context"
	"path"
	"testing"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestProcessVerifyBatchesNil(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "l1infotreesyncTestProcessVerifyBatchesNil.sqlite")
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	err = sut.processVerifyBatches(nil, 1, nil)
	require.Error(t, err)
}

func TestProcessVerifyBatchesOK(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "l1infotreesyncTestProcessVerifyBatchesOK.sqlite")
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	event := VerifyBatches{
		BlockPosition:  1,
		RollupID:       1,
		NumBatch:       1,
		StateRoot:      common.HexToHash("5ca1e"),
		ExitRoot:       common.HexToHash("b455"),
		Aggregator:     common.HexToAddress("beef"),
		RollupExitRoot: common.HexToHash("b455"),
	}
	ctx := context.TODO()
	tx, err := db.NewTx(ctx, sut.db)
	require.NoError(t, err)
	_, err = tx.Exec(`INSERT INTO block (num) VALUES ($1)`, 1)
	require.NoError(t, err)
	err = sut.processVerifyBatches(tx, 1, &event)
	require.NoError(t, err)
}

func TestProcessVerifyBatchesSkip0000(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "l1infotreesyncTestProcessVerifyBatchesSkip0000.sqlite")
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	event := VerifyBatches{
		BlockPosition:  1,
		RollupID:       1,
		NumBatch:       1,
		StateRoot:      common.HexToHash("5ca1e"),
		ExitRoot:       common.Hash{},
		Aggregator:     common.HexToAddress("beef"),
		RollupExitRoot: common.HexToHash("b455"),
	}
	ctx := context.TODO()
	tx, err := db.NewTx(ctx, sut.db)
	require.NoError(t, err)
	err = sut.processVerifyBatches(tx, 1, &event)
	require.NoError(t, err)
}

func TestGetVerifiedBatches(t *testing.T) {
	dbPath := path.Join(t.TempDir(), "l1infotreesyncTestGetVerifiedBatches.sqlite")
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
