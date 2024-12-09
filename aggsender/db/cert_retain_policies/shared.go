package certretainpolicies

import (
	"fmt"
	"time"

	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/db"
	"github.com/ethereum/go-ethereum/common"
)

type HistoryStorager interface {
	DeleteCertificate(tx db.Querier, certificateID common.Hash) error
	MoveCertificateToHistory(tx db.Querier, certificate *types.CertificateInfo) error
	DeleteCertificateHistoryOlderThan(tx db.Querier, now time.Time, olderThan time.Duration) error
}

type HistoryStorage struct {
}

// DeleteCertificate deletes a certificate from the storage using the provided db
func (h *HistoryStorage) DeleteCertificate(tx db.Querier, certificateID common.Hash) error {
	if _, err := tx.Exec(`DELETE FROM certificate_info WHERE certificate_id = $1;`, certificateID.String()); err != nil {
		return fmt.Errorf("error deleting certificate info: %w", err)
	}
	return nil
}

func (h *HistoryStorage) MoveCertificateToHistory(tx db.Querier,
	certificate *types.CertificateInfo) error {
	if _, err := tx.Exec(`INSERT INTO certificate_info_history SELECT * FROM certificate_info WHERE height = $1;`,
		certificate.Height); err != nil {
		return fmt.Errorf("error moving certificate to history: %w", err)
	}
	return nil
}

func (h *HistoryStorage) DeleteCertificateHistoryOlderThan(tx db.Querier, now time.Time, olderThan time.Duration) error {
	deleteOlderThan := now.Add(-olderThan)
	if _, err := tx.Exec(`DELETE FROM certificate_info_history WHERE updated_at < $1;`,
		deleteOlderThan); err != nil {
		return fmt.Errorf("error deleting old certificate to history: %w", err)
	}
	return nil

}
