package zkevmbridge2infoindexsync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sync"
)

const (
	reorgDetectorID = "batchsync"
)

type ZKEVMBridge2L1InfoIndexSync struct {
	driver    *sync.EVMDriver
	processor *processor
}

func New(
	ctx context.Context,
	dbPath string,
	rd sync.ReorgDetector,
	l2Client sync.EthClienter,
	zkevmClient sync.ZKEVMClientInterface,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
	waitForNewBlocksPeriod time.Duration,
	blockFinality etherman.BlockNumberFinality,
	downloadBufferSize int,
) (*ZKEVMBridge2L1InfoIndexSync, error) {
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
	downloader := newDownloader(zkevmClient, l2Client, bf, waitForNewBlocksPeriod, rh)
	driver, err := sync.NewEVMDriver(rd, processor, downloader, reorgDetectorID, downloadBufferSize, rh)
	if err != nil {
		return nil, err
	}

	return &ZKEVMBridge2L1InfoIndexSync{
		driver:    driver,
		processor: processor,
	}, nil
}

func (s *ZKEVMBridge2L1InfoIndexSync) Start(ctx context.Context) {
	s.driver.Sync(ctx)
}
