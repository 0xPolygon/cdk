package types

import (
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/ethereum/go-ethereum/common"
)

type ClaimProof struct {
	ProofLocalExitRoot  [32]common.Hash
	ProofRollupExitRoot [32]common.Hash
	L1InfoTreeLeaf      l1infotreesync.L1InfoTreeLeaf
}
