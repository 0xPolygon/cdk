package l1infotreesync

import (
	"context"
	"testing"

	"github.com/0xPolygon/cdk/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestProcessVerifyBatchesNil(t *testing.T) {
	dbPath := "file:TestProcessVerifyBatchesNil?mode=memory&cache=shared"
	sut, err := newProcessor(dbPath)
	require.NoError(t, err)
	err = sut.processVerifyBatches(nil, 1, nil)
	require.Error(t, err)
}

func TestProcessVerifyBatchesOK(t *testing.T) {
	dbPath := "file:TestGetVerifiedBatches?mode=memory&cache=shared"
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
	dbPath := "file:TestGetVerifiedBatches?mode=memory&cache=shared"
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
