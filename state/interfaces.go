package state

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type storage interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	AddSequence(ctx context.Context, sequence Sequence, dbTx pgx.Tx) error
	CheckProofContainsCompleteSequences(ctx context.Context, proof *Proof, dbTx pgx.Tx) (bool, error)
	GetProofReadyToVerify(ctx context.Context, lastVerfiedBatchNumber uint64, dbTx pgx.Tx) (*Proof, error)
	GetProofsToAggregate(ctx context.Context, dbTx pgx.Tx) (*Proof, *Proof, error)
	AddGeneratedProof(ctx context.Context, proof *Proof, dbTx pgx.Tx) error
	UpdateGeneratedProof(ctx context.Context, proof *Proof, dbTx pgx.Tx) error
	DeleteGeneratedProofs(ctx context.Context, batchNumber uint64, batchNumberFinal uint64, dbTx pgx.Tx) error
	DeleteUngeneratedProofs(ctx context.Context, dbTx pgx.Tx) error
	CleanupGeneratedProofs(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) error
	CleanupLockedProofs(ctx context.Context, duration string, dbTx pgx.Tx) (int64, error)
	CheckProofExistsForBatch(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (bool, error)
}
