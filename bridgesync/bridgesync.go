package bridgesync

import (
	"context"
	"time"

	configTypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
)

const (
	reorgDetectorIDL1  = "bridgesyncl1"
	reorgDetectorIDL2  = "bridgesyncl2"
	dbPrefixL1         = "bridgesyncl1"
	dbPrefixL2         = "bridgesyncl2"
	downloadBufferSize = 1000
)

type Config struct {
	// DBPath path of the DB
	DBPath string `mapstructure:"DBPath"`
	// TODO: BlockFinality doesnt work as per the jsonschema
	BlockFinality string `jsonschema:"enum=latest,enum=safe, enum=pending, enum=finalized" mapstructure:"BlockFinality"`
	// InitialBlockNum is the first block that will be queried when starting the synchronization from scratch.
	// It should be a number equal oir bellow the creation of the bridge contract
	InitialBlockNum uint64 `mapstructure:"InitialBlockNum"`
	// BridgeAddr is the address of the bridge smart contract
	BridgeAddr common.Address `mapstructure:"BridgeAddr"`
	// SyncBlockChunkSize is the amount of blocks that will be queried to the client on each request
	SyncBlockChunkSize uint64 `mapstructure:"SyncBlockChunkSize"`
	// RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
	RetryAfterErrorPeriod configTypes.Duration `mapstructure:"RetryAfterErrorPeriod"`
	// MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
	// Any number smaller than zero will be considered as unlimited retries
	MaxRetryAttemptsAfterError int `mapstructure:"MaxRetryAttemptsAfterError"`
	// WaitForNewBlocksPeriod time that will be waited when the synchronizer has reached the latest block
	WaitForNewBlocksPeriod configTypes.Duration `mapstructure:"WaitForNewBlocksPeriod"`
}

type BridgeSync struct {
	processor *processor
	driver    *sync.EVMDriver
}

// NewL1 creates a bridge syncer that synchronizes the mainnet exit tree
func NewL1(
	ctx context.Context,
	dbPath string,
	bridge common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd sync.ReorgDetector,
	ethClient EthClienter,
	initialBlock uint64,
	waitForNewBlocksPeriod time.Duration,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
) (*BridgeSync, error) {
	return new(
		ctx,
		dbPath,
		bridge,
		syncBlockChunkSize,
		blockFinalityType,
		rd,
		ethClient,
		initialBlock,
		dbPrefixL1,
		reorgDetectorIDL1,
		waitForNewBlocksPeriod,
		retryAfterErrorPeriod,
		maxRetryAttemptsAfterError,
	)
}

// NewL2 creates a bridge syncer that synchronizes the local exit tree
func NewL2(
	ctx context.Context,
	dbPath string,
	bridge common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd sync.ReorgDetector,
	ethClient EthClienter,
	initialBlock uint64,
	waitForNewBlocksPeriod time.Duration,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
) (*BridgeSync, error) {
	return new(
		ctx,
		dbPath,
		bridge,
		syncBlockChunkSize,
		blockFinalityType,
		rd,
		ethClient,
		initialBlock,
		dbPrefixL1,
		reorgDetectorIDL1,
		waitForNewBlocksPeriod,
		retryAfterErrorPeriod,
		maxRetryAttemptsAfterError,
	)
}

func new(
	ctx context.Context,
	dbPath string,
	bridge common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd sync.ReorgDetector,
	ethClient EthClienter,
	initialBlock uint64,
	dbPrefix, reorgDetectorID string,
	waitForNewBlocksPeriod time.Duration,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
) (*BridgeSync, error) {
	processor, err := newProcessor(ctx, dbPath, dbPrefix)
	if err != nil {
		return nil, err
	}
	lastProcessedBlock, err := processor.GetLastProcessedBlock(ctx)
	if err != nil {
		return nil, err
	}
	if lastProcessedBlock < initialBlock {
		err = processor.ProcessBlock(ctx, sync.Block{
			Num: initialBlock,
		})
		if err != nil {
			return nil, err
		}
	}
	rh := &sync.RetryHandler{
		MaxRetryAttemptsAfterError: maxRetryAttemptsAfterError,
		RetryAfterErrorPeriod:      retryAfterErrorPeriod,
	}

	appender, err := buildAppender(ethClient, bridge)
	if err != nil {
		return nil, err
	}
	downloader, err := sync.NewEVMDownloader(
		ethClient,
		syncBlockChunkSize,
		blockFinalityType,
		waitForNewBlocksPeriod,
		appender,
		[]common.Address{bridge},
		rh,
	)
	if err != nil {
		return nil, err
	}

	driver, err := sync.NewEVMDriver(rd, processor, downloader, reorgDetectorID, downloadBufferSize, rh)
	if err != nil {
		return nil, err
	}
	return &BridgeSync{
		processor: processor,
		driver:    driver,
	}, nil
}

// Start starts the synchronization process
func (s *BridgeSync) Start(ctx context.Context) {
	s.driver.Sync(ctx)
}

func (s *BridgeSync) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	return s.processor.GetLastProcessedBlock(ctx)
}

func (s *BridgeSync) GetBridgeIndexByRoot(ctx context.Context, root common.Hash) (uint32, error) {
	return s.processor.exitTree.GetIndexByRoot(ctx, root)
}

func (s *BridgeSync) GetClaimsAndBridges(ctx context.Context, fromBlock, toBlock uint64) ([]Event, error) {
	return s.processor.GetClaimsAndBridges(ctx, fromBlock, toBlock)
}

func (s *BridgeSync) GetProof(ctx context.Context, depositCount uint32, localExitRoot common.Hash) ([32]common.Hash, error) {
	return s.processor.exitTree.GetProof(ctx, depositCount, localExitRoot)
}
