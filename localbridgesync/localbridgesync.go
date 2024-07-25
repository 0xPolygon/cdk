package localbridgesync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
)

const (
	reorgDetectorID    = "localbridgesync"
	downloadBufferSize = 1000
)

var (
	retryAfterErrorPeriod      = time.Second * 10
	maxRetryAttemptsAfterError = 5
)

type LocalBridgeSync struct {
	processor *processor
	driver    *sync.EVMDriver
}

func New(
	dbPath string,
	bridge common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd sync.ReorgDetector,
	l2Client EthClienter,
) (*LocalBridgeSync, error) {
	processor, err := newProcessor(dbPath)
	if err != nil {
		return nil, err
	}

	appender, err := buildAppender(l2Client, bridge)
	if err != nil {
		return nil, err
	}
	downloader, err := sync.NewEVMDownloader(
		l2Client,
		syncBlockChunkSize,
		blockFinalityType,
		waitForNewBlocksPeriod,
		appender,
		[]common.Address{bridge},
	)
	if err != nil {
		return nil, err
	}

	driver, err := sync.NewEVMDriver(rd, processor, downloader, reorgDetectorID, downloadBufferSize)
	if err != nil {
		return nil, err
	}
	return &LocalBridgeSync{
		processor: processor,
		driver:    driver,
	}, nil
}

func (s *LocalBridgeSync) Start(ctx context.Context) {
	s.driver.Sync(ctx)
}
