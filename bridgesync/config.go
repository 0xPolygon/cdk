package bridgesync

import (
	"github.com/0xPolygon/cdk/config/types"
	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	// DBPath path of the DB
	DBPath string `mapstructure:"DBPath"`
	// BlockFinality indicates the status of the blocks that will be queried in order to sync
	BlockFinality string `jsonschema:"enum=LatestBlock, enum=SafeBlock, enum=PendingBlock, enum=FinalizedBlock, enum=EarliestBlock" mapstructure:"BlockFinality"` //nolint:lll
	// InitialBlockNum is the first block that will be queried when starting the synchronization from scratch.
	// It should be a number equal or bellow the creation of the bridge contract
	InitialBlockNum uint64 `mapstructure:"InitialBlockNum"`
	// BridgeAddr is the address of the bridge smart contract
	BridgeAddr common.Address `mapstructure:"BridgeAddr"`
	// SyncBlockChunkSize is the amount of blocks that will be queried to the client on each request
	SyncBlockChunkSize uint64 `mapstructure:"SyncBlockChunkSize"`
	// RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
	RetryAfterErrorPeriod types.Duration `mapstructure:"RetryAfterErrorPeriod"`
	// MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
	// Any number smaller than zero will be considered as unlimited retries
	MaxRetryAttemptsAfterError int `mapstructure:"MaxRetryAttemptsAfterError"`
	// WaitForNewBlocksPeriod time that will be waited when the synchronizer has reached the latest block
	WaitForNewBlocksPeriod types.Duration `mapstructure:"WaitForNewBlocksPeriod"`
	// OriginNetwork is the id of the network where the bridge is deployed
	OriginNetwork uint32 `mapstructure:"OriginNetwork"`
}
