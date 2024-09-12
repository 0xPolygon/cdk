package db

import (
	"context"
	"database/sql"
)

type Tx struct {
	*sql.Tx
	rollbackCallbacks []func()
	commitCallbacks   []func()
}

func NewTx(ctx context.Context, db *sql.DB) (*Tx, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Tx{
		Tx: tx,
	}, nil
}

func (s *Tx) AddRollbackCallback(cb func()) {
	s.rollbackCallbacks = append(s.rollbackCallbacks, cb)
}
func (s *Tx) AddCommitCallback(cb func()) {
	s.commitCallbacks = append(s.commitCallbacks, cb)
}

func (s *Tx) Commit() error {
	if err := s.Tx.Commit(); err != nil {
		return err
	}
	for _, cb := range s.commitCallbacks {
		cb()
	}
	return nil
}

func (s *Tx) Rollback() error {
	if err := s.Tx.Rollback(); err != nil {
		return err
	}
	for _, cb := range s.rollbackCallbacks {
		cb()
	}
	return nil
}
