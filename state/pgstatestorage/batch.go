package pgstatestorage

import (
	"context"
	"errors"

	"github.com/0xPolygon/cdk/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
)

// AddBatch stores a batch
func (p *PostgresStorage) AddBatch(ctx context.Context, dbBatch *state.DBBatch, dbTx pgx.Tx) error {
	const addInputHashSQL = "INSERT INTO aggregator.batch (batch_num, batch, datastream, witness) VALUES ($1, $2, $3, $4) ON CONFLICT (batch_num) DO UPDATE SET batch = $2, datastream = $3, witness = $4"
	e := p.getExecQuerier(dbTx)
	_, err := e.Exec(ctx, addInputHashSQL, dbBatch.Batch.BatchNumber, &dbBatch.Batch, common.Bytes2Hex(dbBatch.Datastream), common.Bytes2Hex(dbBatch.Witness))
	return err
}

// GetBatch gets a batch by a given batch number
func (p *PostgresStorage) GetBatch(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (*state.DBBatch, error) {
	const getInputHashSQL = "SELECT batch, datastream, witness FROM aggregator.batch WHERE batch_num = $1"
	e := p.getExecQuerier(dbTx)
	var batch state.Batch
	var streamStr string
	var witnessStr string
	err := e.QueryRow(ctx, getInputHashSQL, batchNumber).Scan(&batch, &streamStr, &witnessStr)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, state.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return &state.DBBatch{
		Batch:      batch,
		Datastream: common.Hex2Bytes(streamStr),
		Witness:    common.Hex2Bytes(witnessStr),
	}, nil
}

// DeleteBatchesOlderThanBatchNumber deletes batches previous to the given batch number
func (p *PostgresStorage) DeleteBatchesOlderThanBatchNumber(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) error {
	const deleteBatchesSQL = "DELETE FROM aggregator.batch WHERE batch_num < $1"
	e := p.getExecQuerier(dbTx)
	_, err := e.Exec(ctx, deleteBatchesSQL, batchNumber)
	return err
}

// DeleteBatchesNewerThanBatchNumber deletes batches previous to the given batch number
func (p *PostgresStorage) DeleteBatchesNewerThanBatchNumber(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) error {
	const deleteBatchesSQL = "DELETE FROM aggregator.batch WHERE batch_num > $1"
	e := p.getExecQuerier(dbTx)
	_, err := e.Exec(ctx, deleteBatchesSQL, batchNumber)
	return err
}
