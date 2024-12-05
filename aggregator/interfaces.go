package aggregator

import (
	"context"
	"math/big"

	ethmanTypes "github.com/0xPolygon/cdk/aggregator/ethmantypes"
	"github.com/0xPolygon/cdk/aggregator/prover"
	"github.com/0xPolygon/cdk/rpc/types"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	ethtxtypes "github.com/0xPolygon/zkevm-ethtx-manager/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/jackc/pgx/v4"
)

// Consumer interfaces required by the package.
type RPCInterface interface {
	GetBatch(batchNumber uint64) (*types.RPCBatch, error)
	GetWitness(batchNumber uint64, fullWitness bool) ([]byte, error)
}

type ProverInterface interface {
	Name() string
	ID() string
	Addr() string
	IsIdle() (bool, error)
	BatchProof(input *prover.StatelessInputProver) (*string, error)
	AggregatedProof(inputProof1, inputProof2 string) (*string, error)
	FinalProof(inputProof string, aggregatorAddr string) (*string, error)
	WaitRecursiveProof(ctx context.Context, proofID string) (string, common.Hash, common.Hash, error)
	WaitFinalProof(ctx context.Context, proofID string) (*prover.FinalProof, error)
}

// Etherman contains the methods required to interact with ethereum
type Etherman interface {
	GetRollupId() uint32
	GetLatestVerifiedBatchNum() (uint64, error)
	BuildTrustedVerifyBatchesTxData(
		lastVerifiedBatch, newVerifiedBatch uint64, inputs *ethmanTypes.FinalProofInputs, beneficiary common.Address,
	) (to *common.Address, data []byte, err error)
	GetLatestBlockHeader(ctx context.Context) (*ethtypes.Header, error)
	GetBatchAccInputHash(ctx context.Context, batchNumber uint64) (common.Hash, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*ethtypes.Header, error)
}

// aggregatorTxProfitabilityChecker interface for different profitability
// checking algorithms.
type aggregatorTxProfitabilityChecker interface {
	IsProfitable(context.Context, *big.Int) (bool, error)
}

// StateInterface gathers the methods to interact with the state.
type StateInterface interface {
	BeginStateTransaction(ctx context.Context) (pgx.Tx, error)
	CheckProofContainsCompleteSequences(ctx context.Context, proof *state.Proof, dbTx pgx.Tx) (bool, error)
	GetProofReadyToVerify(ctx context.Context, lastVerfiedBatchNumber uint64, dbTx pgx.Tx) (*state.Proof, error)
	GetProofsToAggregate(ctx context.Context, dbTx pgx.Tx) (*state.Proof, *state.Proof, error)
	AddGeneratedProof(ctx context.Context, proof *state.Proof, dbTx pgx.Tx) error
	UpdateGeneratedProof(ctx context.Context, proof *state.Proof, dbTx pgx.Tx) error
	DeleteGeneratedProofs(ctx context.Context, batchNumber uint64, batchNumberFinal uint64, dbTx pgx.Tx) error
	DeleteUngeneratedProofs(ctx context.Context, dbTx pgx.Tx) error
	CleanupGeneratedProofs(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) error
	CleanupLockedProofs(ctx context.Context, duration string, dbTx pgx.Tx) (int64, error)
	CheckProofExistsForBatch(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (bool, error)
	AddSequence(ctx context.Context, sequence state.Sequence, dbTx pgx.Tx) error
}

// EthTxManagerClient represents the eth tx manager interface
type EthTxManagerClient interface {
	Add(
		ctx context.Context,
		to *common.Address,
		value *big.Int,
		data []byte,
		gasOffset uint64,
		sidecar *ethtypes.BlobTxSidecar,
	) (common.Hash, error)
	AddWithGas(
		ctx context.Context,
		to *common.Address,
		value *big.Int,
		data []byte,
		gasOffset uint64,
		sidecar *ethtypes.BlobTxSidecar,
		gas uint64,
	) (common.Hash, error)
	EncodeBlobData(data []byte) (kzg4844.Blob, error)
	MakeBlobSidecar(blobs []kzg4844.Blob) *ethtypes.BlobTxSidecar
	ProcessPendingMonitoredTxs(ctx context.Context, resultHandler ethtxmanager.ResultHandler)
	Remove(ctx context.Context, id common.Hash) error
	RemoveAll(ctx context.Context) error
	Result(ctx context.Context, id common.Hash) (ethtxtypes.MonitoredTxResult, error)
	ResultsByStatus(ctx context.Context, statuses []ethtxtypes.MonitoredTxStatus) ([]ethtxtypes.MonitoredTxResult, error)
	Start()
	Stop()
}
