package types

import (
	"github.com/0xPolygon/cdk/agglayer"
	"github.com/ethereum/go-ethereum/common"
)

type CertificateInfo struct {
	Height           uint64
	CertificateID    common.Hash
	NewLocalExitRoot common.Hash
	FromBlock        uint64
	ToBlock          uint64
	Status           agglayer.CertificateStatus
}
