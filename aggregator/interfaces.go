package aggregator

import (
	"context"
	"math/big"
	"time"

	ethmanTypes "github.com/0xPolygon/cdk/aggregator/ethmantypes"
	"github.com/0xPolygon/cdk/aggregator/prover"
	"github.com/0xPolygon/cdk/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jackc/pgx/v4"
)

// Consumer interfaces required by the package.

type proverInterface interface {
	Name() string
	ID() string
	Addr() string
	IsIdle() (bool, error)
	BatchProof(input *prover.StatelessInputProver) (*string, error)
	AggregatedProof(inputProof1, inputProof2 string) (*string, error)
	FinalProof(inputProof string, aggregatorAddr string) (*string, error)
	WaitRecursiveProof(ctx context.Context, proofID string) (string, common.Hash, error)
	WaitFinalProof(ctx context.Context, proofID string) (*prover.FinalProof, error)
}

// etherman contains the methods required to interact with ethereum
type etherman interface {
	GetRollupId() uint32
	GetLatestVerifiedBatchNum() (uint64, error)
	BuildTrustedVerifyBatchesTxData(
		lastVerifiedBatch, newVerifiedBatch uint64, inputs *ethmanTypes.FinalProofInputs, beneficiary common.Address,
	) (to *common.Address, data []byte, err error)
	GetLatestBlockHeader(ctx context.Context) (*types.Header, error)
	GetBatchAccInputHash(ctx context.Context, batchNumber uint64) (common.Hash, error)
	CheckTxWasMined(ctx context.Context, txHash common.Hash) (bool, *types.Receipt, error)
	CurrentNonce(ctx context.Context, account common.Address) (uint64, error)
	GetHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	GetLatestBlockNumber(ctx context.Context) (uint64, error)
	GetRevertMessage(ctx context.Context, tx *types.Transaction) (string, error)
	EstimateGas(ctx context.Context, from common.Address, to *common.Address, value *big.Int, data []byte) (uint64, error)
	GetSuggestGasTipCap(ctx context.Context) (*big.Int, error)
	GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)
	GetTxReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	PendingNonce(ctx context.Context, account common.Address) (uint64, error)
	SendTx(ctx context.Context, tx *types.Transaction) error
	SignTx(ctx context.Context, sender common.Address, tx *types.Transaction) (*types.Transaction, error)
	WaitTxToBeMined(ctx context.Context, tx *types.Transaction, timeout time.Duration) (bool, error)
	SuggestedGasPrice(ctx context.Context) (*big.Int, error)
	EstimateGasBlobTx(
		ctx context.Context,
		from common.Address,
		to *common.Address,
		gasFeeCap *big.Int,
		gasTipCap *big.Int,
		value *big.Int,
		data []byte,
	) (uint64, error)
}

// aggregatorTxProfitabilityChecker interface for different profitability
// checking algorithms.
type aggregatorTxProfitabilityChecker interface {
	IsProfitable(context.Context, *big.Int) (bool, error)
}

// stateInterface gathers the methods to interact with the state.
type stateInterface interface {
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
	AddBatch(ctx context.Context, dbBatch *state.DBBatch, dbTx pgx.Tx) error
	GetBatch(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (*state.DBBatch, error)
	DeleteBatchesOlderThanBatchNumber(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) error
	DeleteBatchesNewerThanBatchNumber(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) error
}
