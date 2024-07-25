package l1infotreesync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
)

const (
	reorgDetectorID    = "l1infotreesync"
	downloadBufferSize = 1000
)

var (
	retryAfterErrorPeriod      = time.Second * 10
	maxRetryAttemptsAfterError = 5
)

type L1InfoTreeSync struct {
	processor *processor
	driver    *sync.EVMDriver
}

func New(
	ctx context.Context,
	dbPath string,
	globalExitRoot common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd sync.ReorgDetector,
	l1Client EthClienter,
	treeHeight uint8,
	waitForNewBlocksPeriod time.Duration,
	initialBlock uint64,
) (*L1InfoTreeSync, error) {
	processor, err := newProcessor(ctx, dbPath, treeHeight)
	if err != nil {
		return nil, err
	}
	// TODO: get the initialBlock from L1 to simplify config
	lastProcessedBlock, err := processor.GetLastProcessedBlock(ctx)
	if err != nil {
		return nil, err
	}
	if lastProcessedBlock < initialBlock {
		err = processor.ProcessBlock(sync.Block{
			Num: initialBlock,
		})
		if err != nil {
			return nil, err
		}
	}

	appender, err := buildAppender(l1Client, globalExitRoot)
	if err != nil {
		return nil, err
	}
	downloader, err := sync.NewEVMDownloader(
		l1Client,
		syncBlockChunkSize,
		blockFinalityType,
		waitForNewBlocksPeriod,
		appender,
		[]common.Address{globalExitRoot},
	)
	if err != nil {
		return nil, err
	}

	driver, err := sync.NewEVMDriver(rd, processor, downloader, reorgDetectorID, downloadBufferSize)
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

func (s *L1InfoTreeSync) ComputeMerkleProofByIndex(ctx context.Context, index uint32) ([][32]byte, common.Hash, error) {
	return s.processor.ComputeMerkleProofByIndex(ctx, index)
}

func (s *L1InfoTreeSync) ComputeMerkleProofByRoot(ctx context.Context, root common.Hash) ([][32]byte, common.Hash, error) {
	return s.processor.ComputeMerkleProofByRoot(ctx, root)
}

func (s *L1InfoTreeSync) GetInfoByRoot(ctx context.Context, root common.Hash) (*L1InfoTreeLeaf, error) {
	return s.processor.GetInfoByRoot(ctx, root)
}

func (s *L1InfoTreeSync) GetLatestInfoUntilBlock(ctx context.Context, blockNum uint64) (*L1InfoTreeLeaf, error) {
	return s.processor.GetLatestInfoUntilBlock(ctx, blockNum)
}

func (s *L1InfoTreeSync) GetInfoByIndex(ctx context.Context, index uint32) (*L1InfoTreeLeaf, error) {
	return s.processor.GetInfoByIndex(ctx, index)
}

func (s *L1InfoTreeSync) GetInfoByHash(ctx context.Context, hash []byte) (*L1InfoTreeLeaf, error) {
	return s.processor.GetInfoByHash(ctx, hash)
}
