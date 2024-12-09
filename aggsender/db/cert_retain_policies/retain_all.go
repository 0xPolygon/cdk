package certretainpolicies

import (
	"fmt"

	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/db"
)

type CertificateRetainAll struct {
	storage HistoryStorager
}

func (c *CertificateRetainAll) DeprecateCertificate(tx db.Querier, certificate *types.CertificateInfo) error {
	if err := c.storage.MoveCertificateToHistory(tx, certificate); err != nil {
		return fmt.Errorf("certificateRetainAll. Error moving certificate to history: %w", err)
	}
	if err := c.storage.DeleteCertificate(tx, certificate.CertificateID); err != nil {
		return fmt.Errorf("certificateRetainAll. DeleteCertificate %s . Error: %w", certificate.ID(), err)
	}
}
