package certretainpolicies

import (
	"fmt"

	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/db"
)

type CertificateRetainNone struct {
	storage HistoryStorager
}

func (c *CertificateRetainNone) DeprecateCertificate(tx db.Querier, certificate *types.CertificateInfo) error {
	if err := c.storage.DeleteCertificate(tx, certificate.CertificateID); err != nil {
		return fmt.Errorf("certificateRetainNone. DleteCertificate %s . Error: %w", certificate.ID(), err)
	}
}
