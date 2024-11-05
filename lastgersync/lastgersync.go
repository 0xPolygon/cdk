package lastgersync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
)

const (
	reorgDetectorID = "lastGERSync"
)

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
	processor, err := newProcessor(dbPath, "lastGERSync")
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

func (s *LastGERSync) GetFirstGERAfterL1InfoTreeIndex(
	ctx context.Context, atOrAfterL1InfoTreeIndex uint32,
) (Event, error) {
	return s.processor.GetFirstGERAfterL1InfoTreeIndex(ctx, atOrAfterL1InfoTreeIndex)
}

func (s *LastGERSync) GetLastProcessedBlock(ctx context.Context) (uint64, error) {
	return s.processor.GetLastProcessedBlock(ctx)
}
