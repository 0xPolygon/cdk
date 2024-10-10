package types

import (
	"github.com/0xPolygon/cdk/l1infotreesync"
	tree "github.com/0xPolygon/cdk/tree/types"
)

type ClaimProof struct {
	ProofLocalExitRoot  tree.Proof
	ProofRollupExitRoot tree.Proof
	L1InfoTreeLeaf      l1infotreesync.L1InfoTreeLeaf
}
