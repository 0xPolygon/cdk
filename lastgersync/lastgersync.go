package lastgersync

import (
	"context"
	"time"

	configTypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
)

const (
	reorgDetectorID = "lastGERSync"
)

type Config struct {
	// DBPath path of the DB
	DBPath string `mapstructure:"DBPath"`
	// TODO: BlockFinality doesnt work as per the jsonschema
	BlockFinality string `jsonschema:"enum=latest,enum=safe, enum=pending, enum=finalized" mapstructure:"BlockFinality"`
	// InitialBlockNum is the first block that will be queried when starting the synchronization from scratch.
	// It should be a number equal oir bellow the creation of the bridge contract
	InitialBlockNum uint64 `mapstructure:"InitialBlockNum"`
	// GlobalExitRootL2Addr is the address of the GER smart contract on L2
	GlobalExitRootL2Addr common.Address `mapstructure:"GlobalExitRootL2Addr"`
	// RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
	RetryAfterErrorPeriod configTypes.Duration `mapstructure:"RetryAfterErrorPeriod"`
	// MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
	// Any number smaller than zero will be considered as unlimited retries
	MaxRetryAttemptsAfterError int `mapstructure:"MaxRetryAttemptsAfterError"`
	// WaitForNewBlocksPeriod time that will be waited when the synchronizer has reached the latest block
	WaitForNewBlocksPeriod configTypes.Duration `mapstructure:"WaitForNewBlocksPeriod"`
	// DownloadBufferSize buffer of events to be porcessed. When reached will stop downloading events until the processing catches up
	DownloadBufferSize int `mapstructure:"WaitForNewBlocksPeriod"`
}

type LastGERSync struct {
	driver    *sync.EVMDriver
	processor *processor
}

func New(
	ctx context.Context,
	dbPath string,
	rd sync.ReorgDetector,
	l2Client EthClienter,
	globalExitRootL2 common.Address,
	l1InfoTreesync *l1infotreesync.L1InfoTreeSync,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
	blockFinality etherman.BlockNumberFinality,
	waitForNewBlocksPeriod time.Duration,
	downloadBufferSize int,
) (*LastGERSync, error) {
	processor, err := newProcessor(dbPath)
	if err != nil {
		return nil, err
	}

	rh := &sync.RetryHandler{
		RetryAfterErrorPeriod:      retryAfterErrorPeriod,
		MaxRetryAttemptsAfterError: maxRetryAttemptsAfterError,
	}
	bf, err := blockFinality.ToBlockNum()
	if err != nil {
		return nil, err
	}
	downloader, err := newDownloader(
		l2Client,
		globalExitRootL2,
		l1InfoTreesync,
		processor,
		rh,
		bf,
		waitForNewBlocksPeriod,
	)
	if err != nil {
		return nil, err
	}

	driver, err := sync.NewEVMDriver(rd, processor, downloader, reorgDetectorID, downloadBufferSize, rh)
	if err != nil {
		return nil, err
	}

	return &LastGERSync{
		driver:    driver,
		processor: processor,
	}, nil
}

func (s *LastGERSync) Start(ctx context.Context) {
	s.driver.Sync(ctx)
}

func (s *LastGERSync) GetFirstGERAfterL1InfoTreeIndex(ctx context.Context, atOrAfterL1InfoTreeIndex uint32) (injectedL1InfoTreeIndex uint32, ger common.Hash, err error) {
	return s.processor.GetFirstGERAfterL1InfoTreeIndex(ctx, atOrAfterL1InfoTreeIndex)
}

func (s *LastGERSync) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	return s.processor.GetLastProcessedBlock(ctx)
}
