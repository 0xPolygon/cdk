package l1infotreesync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
)

const (
	reorgDetectorID    = "l1infotreesync"
	downloadBufferSize = 1000
)

type Config struct {
	DBPath             string         `mapstructure:"DBPath"`
	GlobalExitRootAddr common.Address `mapstructure:"GlobalExitRootAddr"`
	RollupManagerAddr  common.Address `mapstructure:"RollupManagerAddr"`
	SyncBlockChunkSize uint64         `mapstructure:"SyncBlockChunkSize"`
	// TODO: BlockFinality doesnt work as per the jsonschema
	BlockFinality              string         `jsonschema:"enum=latest,enum=safe, enum=pending, enum=finalized" mapstructure:"BlockFinality"`
	URLRPCL1                   string         `mapstructure:"URLRPCL1"`
	WaitForNewBlocksPeriod     types.Duration `mapstructure:"WaitForNewBlocksPeriod"`
	InitialBlock               uint64         `mapstructure:"InitialBlock"`
	RetryAfterErrorPeriod      types.Duration `mapstructure:"RetryAfterErrorPeriod"`
	MaxRetryAttemptsAfterError int            `mapstructure:"MaxRetryAttemptsAfterError"`
}

type L1InfoTreeSync struct {
	processor *processor
	driver    *sync.EVMDriver
}

func New(
	ctx context.Context,
	dbPath string,
	globalExitRoot, rollupManager common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd sync.ReorgDetector,
	l1Client EthClienter,
	waitForNewBlocksPeriod time.Duration,
	initialBlock uint64,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
) (*L1InfoTreeSync, error) {
	processor, err := newProcessor(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	// TODO: get the initialBlock from L1 to simplify config
	lastProcessedBlock, err := processor.GetLastProcessedBlock(ctx)
	if err != nil {
		return nil, err
	}
	if initialBlock > 0 && lastProcessedBlock < initialBlock-1 {
		err = processor.ProcessBlock(ctx, sync.Block{
			Num: initialBlock - 1,
		})
		if err != nil {
			return nil, err
		}
	}
	rh := &sync.RetryHandler{
		RetryAfterErrorPeriod:      retryAfterErrorPeriod,
		MaxRetryAttemptsAfterError: maxRetryAttemptsAfterError,
	}

	appender, err := buildAppender(l1Client, globalExitRoot, rollupManager)
	if err != nil {
		return nil, err
	}
	downloader, err := sync.NewEVMDownloader(
		l1Client,
		syncBlockChunkSize,
		blockFinalityType,
		waitForNewBlocksPeriod,
		appender,
		[]common.Address{globalExitRoot, rollupManager},
		rh,
	)
	if err != nil {
		return nil, err
	}

	driver, err := sync.NewEVMDriver(rd, processor, downloader, reorgDetectorID, downloadBufferSize, rh)
	if err != nil {
		return nil, err
	}
	return &L1InfoTreeSync{
		processor: processor,
		driver:    driver,
	}, nil
}

func (s *L1InfoTreeSync) Start(ctx context.Context) {
	s.driver.Sync(ctx)
}

func (s *L1InfoTreeSync) GetL1InfoTreeMerkleProof(ctx context.Context, index uint32) ([]common.Hash, common.Hash, error) {
	return s.processor.GetL1InfoTreeMerkleProof(ctx, index)
}

func (s *L1InfoTreeSync) GetLatestInfoUntilBlock(ctx context.Context, blockNum uint64) (*L1InfoTreeLeaf, error) {
	return s.processor.GetLatestInfoUntilBlock(ctx, blockNum)
}

func (s *L1InfoTreeSync) GetInfoByIndex(ctx context.Context, index uint32) (*L1InfoTreeLeaf, error) {
	return s.processor.GetInfoByIndex(ctx, index)
}

func (s *L1InfoTreeSync) GetL1InfoTreeRootByIndex(ctx context.Context, index uint32) (common.Hash, error) {
	tx, err := s.processor.db.BeginRo(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	defer tx.Rollback()

	return s.processor.l1InfoTree.GetRootByIndex(tx, index)
}

func (s *L1InfoTreeSync) GetLastRollupExitRoot(ctx context.Context) (common.Hash, error) {
	return s.processor.rollupExitTree.GetLastRoot(ctx)
}

func (s *L1InfoTreeSync) GetLastL1InfoTreeRootAndIndex(ctx context.Context) (uint32, common.Hash, error) {
	return s.processor.l1InfoTree.GetLastIndexAndRoot(ctx)
}

func (s *L1InfoTreeSync) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	return s.processor.GetLastProcessedBlock(ctx)
}
