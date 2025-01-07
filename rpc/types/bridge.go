package types

import (
	"github.com/agglayer/aggkit/l1infotreesync"
	tree "github.com/agglayer/aggkit/tree/types"
)

type ClaimProof struct {
	ProofLocalExitRoot  tree.Proof
	ProofRollupExitRoot tree.Proof
	L1InfoTreeLeaf      l1infotreesync.L1InfoTreeLeaf
}
