package rpc

import (
	"context"
	"math/big"

	"github.com/0xPolygon/cdk/claimsponsor"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/lastgersync"
	tree "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
)

type Bridger interface {
	GetProof(ctx context.Context, depositCount uint32, localExitRoot common.Hash) (tree.Proof, error)
	GetRootByLER(ctx context.Context, ler common.Hash) (*tree.Root, error)
}

type LastGERer interface {
	GetFirstGERAfterL1InfoTreeIndex(
		ctx context.Context, atOrAfterL1InfoTreeIndex uint32,
	) (lastgersync.Event, error)
}

type L1InfoTreer interface {
	GetInfoByIndex(ctx context.Context, index uint32) (*l1infotreesync.L1InfoTreeLeaf, error)
	GetRollupExitTreeMerkleProof(ctx context.Context, networkID uint32, root common.Hash) (tree.Proof, error)
	GetLocalExitRoot(ctx context.Context, networkID uint32, rollupExitRoot common.Hash) (common.Hash, error)
	GetLastInfo() (*l1infotreesync.L1InfoTreeLeaf, error)
	GetFirstInfo() (*l1infotreesync.L1InfoTreeLeaf, error)
	GetFirstInfoAfterBlock(blockNum uint64) (*l1infotreesync.L1InfoTreeLeaf, error)
	GetLastVerifiedBatches(rollupID uint32) (*l1infotreesync.VerifyBatches, error)
	GetFirstVerifiedBatches(rollupID uint32) (*l1infotreesync.VerifyBatches, error)
	GetFirstVerifiedBatchesAfterBlock(rollupID uint32, blockNum uint64) (*l1infotreesync.VerifyBatches, error)
	GetFirstL1InfoWithRollupExitRoot(rollupExitRoot common.Hash) (*l1infotreesync.L1InfoTreeLeaf, error)
}

type ClaimSponsorer interface {
	AddClaimToQueue(claim *claimsponsor.Claim) error
	GetClaim(globalIndex *big.Int) (*claimsponsor.Claim, error)
}
