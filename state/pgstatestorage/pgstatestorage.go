package pgstatestorage

import (
	"github.com/0xPolygon/cdk/state"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// PostgresStorage implements the Storage interface
type PostgresStorage struct {
	cfg state.Config
	*pgxpool.Pool
}

// NewPostgresStorage creates a new StateDB
func NewPostgresStorage(cfg state.Config, db *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{
		cfg,
		db,
	}
}

// getExecQuerier determines which execQuerier to use, dbTx or the main pgxpool
func (p *PostgresStorage) getExecQuerier(dbTx pgx.Tx) ExecQuerier {
	if dbTx != nil {
		return dbTx
	}
	return p
}
