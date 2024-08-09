package l1infotreesync

import (
	"context"
	"errors"
	"time"

	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree"
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

// New creates a L1 Info tree syncer that syncs the L1 info tree
// and the rollup exit tree
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

// Start starts the synchronization process
func (s *L1InfoTreeSync) Start(ctx context.Context) {
	s.driver.Sync(ctx)
}

// GetL1InfoTreeMerkleProof creates a merkle proof for the L1 Info tree
func (s *L1InfoTreeSync) GetL1InfoTreeMerkleProof(ctx context.Context, index uint32) ([32]common.Hash, common.Hash, error) {
	return s.processor.GetL1InfoTreeMerkleProof(ctx, index)
}

// GetRollupExitTreeMerkleProof creates a merkle proof for the rollup exit tree
func (s *L1InfoTreeSync) GetRollupExitTreeMerkleProof(ctx context.Context, networkID uint32, root common.Hash) ([32]common.Hash, error) {
	if networkID == 0 {
		return tree.EmptyProof, nil
	}
	return s.processor.rollupExitTree.GetProof(ctx, networkID-1, root)
}

// GetLatestInfoUntilBlock returns the most recent L1InfoTreeLeaf that occurred before or at blockNum.
// If the blockNum has not been processed yet the error ErrBlockNotProcessed will be returned
func (s *L1InfoTreeSync) GetLatestInfoUntilBlock(ctx context.Context, blockNum uint64) (*L1InfoTreeLeaf, error) {
	return s.processor.GetLatestInfoUntilBlock(ctx, blockNum)
}

// GetInfoByIndex returns the value of a leave (not the hash) of the L1 info tree
func (s *L1InfoTreeSync) GetInfoByIndex(ctx context.Context, index uint32) (*L1InfoTreeLeaf, error) {
	return s.processor.GetInfoByIndex(ctx, index)
}

// GetL1InfoTreeRootByIndex returns the root of the L1 info tree at the moment the leaf with the given index was added
func (s *L1InfoTreeSync) GetL1InfoTreeRootByIndex(ctx context.Context, index uint32) (common.Hash, error) {
	tx, err := s.processor.db.BeginRo(ctx)
	if err != nil {
		return common.Hash{}, err
	}
	defer tx.Rollback()

	return s.processor.l1InfoTree.GetRootByIndex(tx, index)
}

// GetLastRollupExitRoot return the last rollup exit root processed
func (s *L1InfoTreeSync) GetLastRollupExitRoot(ctx context.Context) (common.Hash, error) {
	return s.processor.rollupExitTree.GetLastRoot(ctx)
}

// GetLastL1InfoTreeRootAndIndex return the last root and index processed from the L1 Info tree
func (s *L1InfoTreeSync) GetLastL1InfoTreeRootAndIndex(ctx context.Context) (uint32, common.Hash, error) {
	return s.processor.l1InfoTree.GetLastIndexAndRoot(ctx)
}

// GetLastProcessedBlock return the last processed block
func (s *L1InfoTreeSync) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	return s.processor.GetLastProcessedBlock(ctx)
}

func (s *L1InfoTreeSync) GetLocalExitRoot(ctx context.Context, networkID uint32, rollupExitRoot common.Hash) (common.Hash, error) {
	if networkID == 0 {
		return common.Hash{}, errors.New("network 0 is not a rollup, and it's not part of the rollup exit tree")
	}
	return s.processor.rollupExitTree.GetLeaf(ctx, networkID-1, rollupExitRoot)
}
