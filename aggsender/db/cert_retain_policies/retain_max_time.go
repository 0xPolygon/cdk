package certretainpolicies

import (
	"fmt"
	"time"

	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/db"
)

type CertificateRaetainMaxTime struct {
	storage HistoryStorager
	MaxTime time.Duration
}

func (c *CertificateRaetainMaxTime) DeprecateCertificate(tx db.Querier, certificate *types.CertificateInfo, now time.Time) error {
	if err := c.storage.MoveCertificateToHistory(tx, certificate); err != nil {
		return fmt.Errorf("CertificateRaetainMaxTime. Error moving certificate to history: %w", err)
	}
	if err := c.storage.DeleteCertificateHistoryOlderThan(tx, now, c.MaxTime); err != nil {
		return fmt.Errorf("CertificateRaetainMaxTime. Error cleaning old data. Error: %w", err)
	}
}
