package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/0xPolygon/cdk/aggsender/db/migrations"
	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/russross/meddler"
)

// AggSenderStorage is the interface that defines the methods to interact with the storage
type AggSenderStorage interface {
	// GetCertificateByHeight returns a certificate by its height
	GetCertificateByHeight(ctx context.Context, height uint64) (types.CertificateInfo, error)
	// GetLastSentCertificate returns the last certificate sent to the aggLayer
	GetLastSentCertificate(ctx context.Context) (types.CertificateInfo, error)
	// SaveLastSentCertificate saves the last certificate sent to the aggLayer
	SaveLastSentCertificate(ctx context.Context, certificate types.CertificateInfo) error
	// DeleteCertificate deletes a certificate from the storage
	DeleteCertificate(ctx context.Context, certificateID common.Hash) error
}

var _ AggSenderStorage = (*AggSenderSQLStorage)(nil)

// AggSenderSQLStorage is the struct that implements the AggSenderStorage interface
type AggSenderSQLStorage struct {
	logger *log.Logger
	db     *sql.DB
}

// NewAggSenderSQLStorage creates a new AggSenderSQLStorage
func NewAggSenderSQLStorage(logger *log.Logger, dbPath string) (*AggSenderSQLStorage, error) {
	if err := migrations.RunMigrations(dbPath); err != nil {
		return nil, err
	}

	db, err := db.NewSQLiteDB(dbPath)
	if err != nil {
		return nil, err
	}

	return &AggSenderSQLStorage{
		db:     db,
		logger: logger,
	}, nil
}

// GetCertificateByHeight returns a certificate by its height
func (a *AggSenderSQLStorage) GetCertificateByHeight(ctx context.Context,
	height uint64) (types.CertificateInfo, error) {
	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return types.CertificateInfo{}, err
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			a.logger.Warnf("error rolling back tx: %w", err)
		}
	}()

	rows, err := tx.Query(`SELECT * FROM certificate_info WHERE height = $1;`, height)
	if err != nil {
		return types.CertificateInfo{}, getSelectQueryError(height, err)
	}

	var certificateInfo types.CertificateInfo
	if err = meddler.ScanRow(rows, &certificateInfo); err != nil {
		return types.CertificateInfo{}, getSelectQueryError(height, err)
	}

	return certificateInfo, nil
}

// GetLastSentCertificate returns the last certificate sent to the aggLayer
func (a *AggSenderSQLStorage) GetLastSentCertificate(ctx context.Context) (types.CertificateInfo, error) {
	tx, err := a.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return types.CertificateInfo{}, err
	}

	defer func() {
		if err := tx.Rollback(); err != nil {
			a.logger.Warnf("error rolling back tx: %w", err)
		}
	}()

	rows, err := tx.Query(`SELECT * FROM certificate_info ORDER BY height DESC LIMIT 1;`)
	if err != nil {
		return types.CertificateInfo{}, getSelectQueryError(math.MaxUint64, err) // force checking err not found
	}

	var certificateInfo types.CertificateInfo
	if err = meddler.ScanRow(rows, &certificateInfo); err != nil {
		return types.CertificateInfo{}, getSelectQueryError(math.MaxUint64, err) // force checking err not found
	}

	return certificateInfo, nil
}

// SaveLastSentCertificate saves the last certificate sent to the aggLayer
func (a *AggSenderSQLStorage) SaveLastSentCertificate(ctx context.Context, certificate types.CertificateInfo) error {
	tx, err := db.NewTx(ctx, a.db)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if errRllbck := tx.Rollback(); errRllbck != nil {
				a.logger.Errorf("error while rolling back tx %w", errRllbck)
			}
		}
	}()

	if err := meddler.Insert(tx, "certificate_info", &certificate); err != nil {
		return fmt.Errorf("error inserting certificate info: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	a.logger.Debugf("inserted certificate - Height: %d. Hash: %s", certificate.Height, certificate.CertificateID)

	return nil
}

// DeleteCertificate deletes a certificate from the storage
func (a *AggSenderSQLStorage) DeleteCertificate(ctx context.Context, certificateID common.Hash) error {
	tx, err := db.NewTx(ctx, a.db)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if errRllbck := tx.Rollback(); errRllbck != nil {
				a.logger.Errorf("error while rolling back tx %w", errRllbck)
			}
		}
	}()

	if _, err := tx.Exec(`DELETE FROM certificate_info WHERE certificate_id = $1;`, certificateID); err != nil {
		return fmt.Errorf("error deleting certificate info: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	a.logger.Debugf("deleted certificate - CertificateID: %s", certificateID)

	return nil
}

// clean deletes all the data from the storage
// NOTE: Used only in tests
func (a *AggSenderSQLStorage) clean() error {
	if _, err := a.db.Exec(`DELETE FROM certificate_info;`); err != nil {
		return err
	}

	return nil
}

func getSelectQueryError(height uint64, err error) error {
	errToReturn := err
	if errors.Is(err, sql.ErrNoRows) {
		if height == 0 {
			// height 0 is never sent to the aggLayer
			// so we don't return an error in this case
			errToReturn = nil
		} else {
			errToReturn = db.ErrNotFound
		}
	}

	return errToReturn
}
