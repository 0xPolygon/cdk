package aggregator

import (
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk/aggregator/db"
	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	syncronizerConfig "github.com/0xPolygonHermez/zkevm-synchronizer-l1/config"
)

// SettlementBackend is the type of the settlement backend
type SettlementBackend string

const (
	// AggLayer settlement backend
	AggLayer SettlementBackend = "agglayer"

	// L1 settlement backend
	L1 SettlementBackend = "l1"

	// TenToThePowerOf18 represents 1000000000000000000
	TenToThePowerOf18 = 1000000000000000000
)

// TokenAmountWithDecimals is a wrapper type that parses token amount with decimals to big int
type TokenAmountWithDecimals struct {
	*big.Int `validate:"required"`
}

// UnmarshalText unmarshal token amount from float string to big int
func (t *TokenAmountWithDecimals) UnmarshalText(data []byte) error {
	amount, ok := new(big.Float).SetString(string(data))
	if !ok {
		return fmt.Errorf("failed to unmarshal string to float")
	}
	coin := new(big.Float).SetInt(big.NewInt(TenToThePowerOf18))
	bigval := new(big.Float).Mul(amount, coin)
	result := new(big.Int)
	bigval.Int(result)
	t.Int = result

	return nil
}

// Config represents the configuration of the aggregator
type Config struct {
	// Host for the grpc server
	Host string `mapstructure:"Host"`
	// Port for the grpc server
	Port int `mapstructure:"Port"`

	// RetryTime is the time the aggregator main loop sleeps if there are no proofs to aggregate
	// or batches to generate proofs. It is also used in the isSynced loop
	RetryTime types.Duration `mapstructure:"RetryTime"`

	// VerifyProofInterval is the interval of time to verify/send an proof in L1
	VerifyProofInterval types.Duration `mapstructure:"VerifyProofInterval"`

	// ProofStatePollingInterval is the interval time to polling the prover about the generation state of a proof
	ProofStatePollingInterval types.Duration `mapstructure:"ProofStatePollingInterval"`

	// TxProfitabilityCheckerType type for checking is it profitable for aggregator to validate batch
	// possible values: base/acceptall
	TxProfitabilityCheckerType TxProfitabilityCheckerType `mapstructure:"TxProfitabilityCheckerType"`

	// TxProfitabilityMinReward min reward for base tx profitability checker when aggregator will validate batch
	// this parameter is used for the base tx profitability checker
	TxProfitabilityMinReward TokenAmountWithDecimals `mapstructure:"TxProfitabilityMinReward"`

	// IntervalAfterWhichBatchConsolidateAnyway is the interval duration for the main sequencer to check
	// if there are no transactions. If there are no transactions in this interval, the sequencer will
	// consolidate the batch anyway.
	IntervalAfterWhichBatchConsolidateAnyway types.Duration `mapstructure:"IntervalAfterWhichBatchConsolidateAnyway"`

	// BatchProofSanityCheckEnabled is a flag to enable the sanity check of the batch proof
	BatchProofSanityCheckEnabled bool `mapstructure:"BatchProofSanityCheckEnabled"`

	// ChainID is the L2 ChainID provided by the Network Config
	ChainID uint64

	// ForkID is the L2 ForkID provided by the Network Config
	ForkId uint64 `mapstructure:"ForkId"` //nolint:stylecheck

	// SenderAddress defines which private key the eth tx manager needs to use
	// to sign the L1 txs
	SenderAddress string `mapstructure:"SenderAddress"`

	// CleanupLockedProofsInterval is the interval of time to clean up locked proofs.
	CleanupLockedProofsInterval types.Duration `mapstructure:"CleanupLockedProofsInterval"`

	// GeneratingProofCleanupThreshold represents the time interval after
	// which a proof in generating state is considered to be stuck and
	// allowed to be cleared.
	GeneratingProofCleanupThreshold string `mapstructure:"GeneratingProofCleanupThreshold"`

	// GasOffset is the amount of gas to be added to the gas estimation in order
	// to provide an amount that is higher than the estimated one. This is used
	// to avoid the TX getting reverted in case something has changed in the network
	// state after the estimation which can cause the TX to require more gas to be
	// executed.
	//
	// ex:
	// gas estimation: 1000
	// gas offset: 100
	// final gas: 1100
	GasOffset uint64 `mapstructure:"GasOffset"`

	// RPCURL is the URL of the RPC server
	RPCURL string `mapstructure:"RPCURL"`

	// WitnessURL is the URL of the witness server
	WitnessURL string `mapstructure:"WitnessURL"`

	// UseFullWitness is a flag to enable the use of full witness in the aggregator
	UseFullWitness bool `mapstructure:"UseFullWitness"`

	// DB is the database configuration
	DB db.Config `mapstructure:"DB"`

	// EthTxManager is the config for the ethtxmanager
	EthTxManager ethtxmanager.Config `mapstructure:"EthTxManager"`

	// Log is the log configuration
	Log log.Config `mapstructure:"Log"`

	// Synchornizer config
	Synchronizer syncronizerConfig.Config `mapstructure:"Synchronizer"`

	// SettlementBackend configuration defines how a final ZKP should be settled.
	// It can be settled directly to L1 or over Agglayer.
	SettlementBackend SettlementBackend `mapstructure:"SettlementBackend" jsonschema:"enum=agglayer,enum=l1"`

	// SequencerPrivateKey Private key of the trusted sequencer
	SequencerPrivateKey types.KeystoreFileConfig `mapstructure:"SequencerPrivateKey"`

	// AggLayerTxTimeout is the interval time to wait for a tx to be mined from the agglayer
	AggLayerTxTimeout types.Duration `mapstructure:"AggLayerTxTimeout"`

	// AggLayerURL url of the agglayer service
	AggLayerURL string `mapstructure:"AggLayerURL"`

	// SyncModeOnlyEnabled is a flag that activates sync mode exclusively.
	// When enabled, the aggregator will sync data only from L1 and will not generate or read the data stream.
	SyncModeOnlyEnabled bool `mapstructure:"SyncModeOnlyEnabled"`
}
