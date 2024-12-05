package types

import (
	"context"
	"encoding/binary"
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
	Fatalf(format string, args ...interface{})
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
	Height        uint64      `meddler:"height"`
	RetryCount    int         `meddler:"retry_count"`
	CertificateID common.Hash `meddler:"certificate_id,hash"`
	// PreviousLocalExitRoot if it's nil means no reported
	PreviousLocalExitRoot *common.Hash               `meddler:"previous_local_exit_root,hash"`
	NewLocalExitRoot      common.Hash                `meddler:"new_local_exit_root,hash"`
	FromBlock             uint64                     `meddler:"from_block"`
	ToBlock               uint64                     `meddler:"to_block"`
	Status                agglayer.CertificateStatus `meddler:"status"`
	CreatedAt             int64                      `meddler:"created_at"`
	UpdatedAt             int64                      `meddler:"updated_at"`
	SignedCertificate     string                     `meddler:"signed_certificate"`
}

func (c *CertificateInfo) String() string {
	if c == nil {
		//nolint:all
		return "nil"
	}
	previousLocalExitRoot := "nil"
	if c.PreviousLocalExitRoot != nil {
		previousLocalExitRoot = c.PreviousLocalExitRoot.String()
	}
	return fmt.Sprintf("aggsender.CertificateInfo: "+
		"Height: %d "+
		"RetryCount: %d "+
		"CertificateID: %s "+
		"PreviousLocalExitRoot: %s "+
		"NewLocalExitRoot: %s "+
		"Status: %s "+
		"FromBlock: %d "+
		"ToBlock: %d "+
		"CreatedAt: %s "+
		"UpdatedAt: %s",
		c.Height,
		c.RetryCount,
		c.CertificateID.String(),
		previousLocalExitRoot,
		c.NewLocalExitRoot.String(),
		c.Status.String(),
		c.FromBlock,
		c.ToBlock,
		time.UnixMilli(c.CreatedAt),
		time.UnixMilli(c.UpdatedAt),
	)
}

// ID returns a string with the unique identifier of the cerificate (height+certificateID)
func (c *CertificateInfo) ID() string {
	if c == nil {
		return "nil"
	}
	return fmt.Sprintf("%d/%s (retry %d)", c.Height, c.CertificateID.String(), c.RetryCount)
}

// IsClosed returns true if the certificate is closed (settled or inError)
func (c *CertificateInfo) IsClosed() bool {
	if c == nil {
		return false
	}
	return c.Status.IsClosed()
}

// ElapsedTimeSinceCreation returns the time elapsed since the certificate was created
func (c *CertificateInfo) ElapsedTimeSinceCreation() time.Duration {
	if c == nil {
		return 0
	}
	return time.Now().UTC().Sub(time.UnixMilli(c.CreatedAt))
}

type CertificateMetadata struct {
	FromBlock uint64 `json:"fromBlock"`
	Offset    uint32 `json:"toBlock"`
	CreatedAt uint32 `json:"createdAt"`
	Version   uint8  `json:"version"`
}

// NewCertificateMetadataFromHash returns a new CertificateMetadata from the given hash
func NewCertificateMetadata(fromBlock uint64, offset uint32, createdAt uint32) *CertificateMetadata {
	return &CertificateMetadata{
		FromBlock: fromBlock,
		Offset:    offset,
		CreatedAt: createdAt,
		Version:   1,
	}
}

// NewCertificateMetadataFromHash returns a new CertificateMetadata from the given hash
func NewCertificateMetadataFromHash(hash common.Hash) *CertificateMetadata {
	b := hash.Bytes()

	if b[0] < 1 {
		hash.Big().Uint64()

		return &CertificateMetadata{
			Version:   b[0],
			FromBlock: binary.BigEndian.Uint64(b[1:9]),
			Offset:    binary.BigEndian.Uint32(b[9:13]),
			CreatedAt: binary.BigEndian.Uint32(b[13:17]),
		}
	}

	return &CertificateMetadata{
		Version:   b[0],
		FromBlock: binary.BigEndian.Uint64(b[1:9]),
		Offset:    binary.BigEndian.Uint32(b[9:13]),
		CreatedAt: binary.BigEndian.Uint32(b[13:17]),
	}
}

// ToHash returns the hash of the metadata
func (c *CertificateMetadata) ToHash() common.Hash {
	b := make([]byte, common.HashLength) // 32-byte hash

	// Encode version
	b[0] = c.Version

	// Encode fromBlock
	binary.BigEndian.PutUint64(b[1:9], c.FromBlock)

	// Encode offset
	binary.BigEndian.PutUint32(b[9:13], c.Offset)

	// Encode createdAt
	binary.BigEndian.PutUint32(b[13:17], c.CreatedAt)

	// Last 8 bytes remain as zero padding

	return common.BytesToHash(b)
}
