package sequencesender

import (
	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	"github.com/ethereum/go-ethereum/common"
)

// Config represents the configuration of a sequence sender
type Config struct {
	// WaitPeriodSendSequence is the time the sequencer waits until
	// trying to send a sequence to L1
	WaitPeriodSendSequence types.Duration `mapstructure:"WaitPeriodSendSequence"`
	// LastBatchVirtualizationTimeMaxWaitPeriod is time since sequences should be sent
	LastBatchVirtualizationTimeMaxWaitPeriod types.Duration `mapstructure:"LastBatchVirtualizationTimeMaxWaitPeriod"`
	// L1BlockTimestampMargin is the time difference (margin) that must exists between last L1 block
	// and last L2 block in the sequence before sending the sequence to L1. If the difference is
	// lower than this value, then sequencesender will wait until the difference is equal or greater
	L1BlockTimestampMargin types.Duration `mapstructure:"L1BlockTimestampMargin"`
	// MaxTxSizeForL1 is the maximum size a single transaction can have. This field has
	// non-trivial consequences: larger transactions than 128KB are significantly harder and
	// more expensive to propagate; larger transactions also take more resources
	// to validate whether they fit into the pool or not.
	MaxTxSizeForL1 uint64 `mapstructure:"MaxTxSizeForL1"`
	// SenderAddress defines which private key the eth tx manager needs to use
	// to sign the L1 txs
	SenderAddress common.Address
	// L2Coinbase defines which address is going to receive the fees
	L2Coinbase common.Address `mapstructure:"L2Coinbase"`
	// PrivateKey defines all the key store files that are going
	// to be read in order to provide the private keys to sign the L1 txs
	PrivateKey types.KeystoreFileConfig `mapstructure:"PrivateKey"`
	// Batch number where there is a forkid change (fork upgrade)
	ForkUpgradeBatchNumber uint64
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

	// SequencesTxFileName is the file name to store sequences sent to L1
	SequencesTxFileName string

	// WaitPeriodPurgeTxFile is the time to wait before purging from file the finished sent L1 tx
	WaitPeriodPurgeTxFile types.Duration `mapstructure:"WaitPeriodPurgeTxFile"`

	// MaxPendingTx is the maximum number of pending transactions (those that are not in a final state)
	MaxPendingTx uint64

	// EthTxManager is the config for the ethtxmanager
	EthTxManager ethtxmanager.Config `mapstructure:"EthTxManager"`

	// Log is the log configuration
	Log log.Config `mapstructure:"Log"`

	// MaxBatchesForL1 is the maximum amount of batches to be sequenced in a single L1 tx
	MaxBatchesForL1 uint64 `mapstructure:"MaxBatchesForL1"`
	// BlockFinality indicates the status of the blocks that will be queried in order to sync
	BlockFinality string `jsonschema:"enum=LatestBlock, enum=SafeBlock, enum=PendingBlock, enum=FinalizedBlock, enum=EarliestBlock" mapstructure:"BlockFinality"` //nolint:lll

	// RPCURL is the URL of the RPC server
	RPCURL string `mapstructure:"RPCURL"`

	// GetBatchWaitInterval is the time to wait to query for a new batch when there are no more batches available
	GetBatchWaitInterval types.Duration `mapstructure:"GetBatchWaitInterval"`
}
