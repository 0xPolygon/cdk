package pgstatestorage

import (
	"context"

	"github.com/0xPolygon/cdk/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
)

// AddBatch stores a batch
func (p *PostgresStorage) AddBatch(ctx context.Context, batch *state.Batch, datastream []byte, dbTx pgx.Tx) error {
	const addInputHashSQL = "INSERT INTO aggregator.batch (batch_num, batch, datastream) VALUES ($1, $2, $3) ON CONFLICT (batch_num) DO UPDATE SET batch = $2, datastream = $3"
	e := p.getExecQuerier(dbTx)
	_, err := e.Exec(ctx, addInputHashSQL, batch.BatchNumber, &batch, common.Bytes2Hex(datastream))
	return err
}

// GetBatch gets a batch by a given batch number
func (p *PostgresStorage) GetBatch(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (*state.Batch, []byte, error) {
	const getInputHashSQL = "SELECT batch, datastream FROM aggregator.batch WHERE batch_num = $1"
	e := p.getExecQuerier(dbTx)
	var batch *state.Batch
	var streamStr string
	err := e.QueryRow(ctx, getInputHashSQL, batchNumber).Scan(&batch, &streamStr)
	if err != nil {
		return nil, nil, err
	}
	return batch, common.Hex2Bytes(streamStr), nil
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
