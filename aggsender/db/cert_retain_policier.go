package db

import (
	"time"

	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/db"
)

type CertificateRatainPolicier interface {
	// DeprecateCertificate is called when a new certificate is added and an old one
	// is deprecated. The old certificate is passed as a parameter
	DeprecateCertificate(tx db.Querier, certificate *types.CertificateInfo, now time.Time) error
}
