package agglayer

import (
	"fmt"
	"testing"

	"github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
)

// Helper function to create a dummy TokenInfo
func createDummyTokenInfo(t *testing.T) *TokenInfo {
	t.Helper()

	return &TokenInfo{
		OriginNetwork:      1,
		OriginTokenAddress: common.HexToAddress("0x2345"),
	}
}

// Helper function to create a dummy GlobalIndex
func createDummyGlobalIndex(t *testing.T) *GlobalIndex {
	t.Helper()

	return &GlobalIndex{
		MainnetFlag: false,
		RollupIndex: 10,
		LeafIndex:   1,
	}
}

// Helper function to create a dummy Claim
func createDummyClaim(t *testing.T) *ClaimFromMainnnet {
	t.Helper()

	return &ClaimFromMainnnet{
		ProofLeafMER: &MerkleProof{
			Root: common.HexToHash("0x1234"),
			Proof: [common.HashLength]common.Hash{
				common.HexToHash("0x1234"),
				common.HexToHash("0x5678"),
			},
		},
		ProofGERToL1Root: &MerkleProof{
			Root: common.HexToHash("0x5678"),
			Proof: [common.HashLength]common.Hash{
				common.HexToHash("0x5678"),
				common.HexToHash("0x1234"),
			},
		},
		L1Leaf: &L1InfoTreeLeaf{
			L1InfoTreeIndex: 1,
			RollupExitRoot:  common.HexToHash("0x987654321"),
			MainnetExitRoot: common.HexToHash("0x123456789"),
			Inner:           &L1InfoTreeLeafInner{},
		},
	}
}

// Helper function to create a dummy proof
func createDummyProof(t *testing.T) types.Proof {
	t.Helper()

	proof := types.Proof{}

	for i := 0; i < int(types.DefaultHeight); i++ {
		proof[i] = common.HexToHash(fmt.Sprintf("0x%x", i))
	}

	return proof
}
