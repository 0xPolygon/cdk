package certretainpolicies

import (
	"time"

	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/db"
)

type CertificateRaetainMaxTime struct {
	MaxTime time.Duration
}

func (c *CertificateRaetainMaxTime) DeprecateCertificate(tx db.Querier, certificate *types.CertificateInfo) error {
	return nil
}
