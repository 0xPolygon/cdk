package types

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	treeTypes "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// L1InfoTreeSyncer is an interface defining functions that an L1InfoTreeSyncer should implement
type L1InfoTreeSyncer interface {
	GetInfoByGlobalExitRoot(globalExitRoot common.Hash) (*l1infotreesync.L1InfoTreeLeaf, error)
	GetL1InfoTreeMerkleProofFromIndexToRoot(
		ctx context.Context, index uint32, root common.Hash,
	) (treeTypes.Proof, error)
	GetL1InfoTreeRootByIndex(ctx context.Context, index uint32) (treeTypes.Root, error)
}

// L2BridgeSyncer is an interface defining functions that an L2BridgeSyncer should implement
type L2BridgeSyncer interface {
	GetBlockByLER(ctx context.Context, ler common.Hash) (uint64, error)
	GetExitRootByIndex(ctx context.Context, index uint32) (treeTypes.Root, error)
	GetBridgesPublished(ctx context.Context, fromBlock, toBlock uint64) ([]bridgesync.Bridge, error)
	GetClaims(ctx context.Context, fromBlock, toBlock uint64) ([]bridgesync.Claim, error)
	OriginNetwork() uint32
	BlockFinality() etherman.BlockNumberFinality
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
}

// EthClient is an interface defining functions that an EthClient should implement
type EthClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

// Logger is an interface that defines the methods to log messages
type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
}

type CertificateInfo struct {
	Height            uint64                     `meddler:"height"`
	CertificateID     common.Hash                `meddler:"certificate_id,hash"`
	NewLocalExitRoot  common.Hash                `meddler:"new_local_exit_root,hash"`
	FromBlock         uint64                     `meddler:"from_block"`
	ToBlock           uint64                     `meddler:"to_block"`
	Status            agglayer.CertificateStatus `meddler:"status"`
	CreatedAt         int64                      `meddler:"created_at"`
	UpdatedAt         int64                      `meddler:"updated_at"`
	SignedCertificate string                     `meddler:"signed_certificate"`
}

func (c CertificateInfo) String() string {
	return fmt.Sprintf(
		"Height: %d\n"+
			"CertificateID: %s\n"+
			"FromBlock: %d\n"+
			"ToBlock: %d\n"+
			"NewLocalExitRoot: %s\n"+
			"Status: %s\n"+
			"CreatedAt: %s\n"+
			"UpdatedAt: %s\n",
		c.Height,
		c.CertificateID.String(),
		c.FromBlock,
		c.ToBlock,
		c.NewLocalExitRoot.String(),
		c.Status.String(),
		time.UnixMilli(c.CreatedAt),
		time.UnixMilli(c.UpdatedAt),
	)
}
