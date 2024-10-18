package types

import (
	"fmt"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/ethereum/go-ethereum/common"
)

type CertificateInfo struct {
	Height           uint64                     `meddler:"height"`
	CertificateID    common.Hash                `meddler:"certificate_id"`
	NewLocalExitRoot common.Hash                `meddler:"new_local_exit_root"`
	FromBlock        uint64                     `meddler:"from_block"`
	ToBlock          uint64                     `meddler:"to_block"`
	Status           agglayer.CertificateStatus `meddler:"status"`
}

func (c CertificateInfo) String() string {
	return fmt.Sprintf("Height: %d, CertificateID: %s", c.Height, c.CertificateID.String())
}
