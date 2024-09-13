package pgstatestorage

import (
	"context"

	"github.com/0xPolygon/cdk/state"
	"github.com/jackc/pgx/v4"
)

// AddSequence stores the sequence information to allow the aggregator verify sequences.
func (p *PostgresStorage) AddSequence(ctx context.Context, sequence state.Sequence, dbTx pgx.Tx) error {
	const addSequenceSQL = `
	INSERT INTO aggregator.sequence (from_batch_num, to_batch_num) 
	VALUES($1, $2) 
	ON CONFLICT (from_batch_num) DO UPDATE SET to_batch_num = $2
	`

	e := p.getExecQuerier(dbTx)
	_, err := e.Exec(ctx, addSequenceSQL, sequence.FromBatchNumber, sequence.ToBatchNumber)
	return err
}
