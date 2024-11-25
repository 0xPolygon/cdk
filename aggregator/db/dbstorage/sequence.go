package dbstorage

import (
	"context"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/state"
)

// AddSequence stores the sequence information to allow the aggregator verify sequences.
func (d *DBStorage) AddSequence(ctx context.Context, sequence state.Sequence, dbTx db.Txer) error {
	const addSequenceSQL = `
	INSERT INTO sequence (from_batch_num, to_batch_num) 
	VALUES($1, $2) 
	ON CONFLICT (from_batch_num) DO UPDATE SET to_batch_num = $2
	`

	e := d.getExecQuerier(dbTx)
	_, err := e.Exec(addSequenceSQL, sequence.FromBatchNumber, sequence.ToBatchNumber)
	return err
}
