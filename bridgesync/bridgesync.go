package bridgesync

import (
	"context"
	"time"

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

var (
	retryAfterErrorPeriod      = time.Second * 10
	maxRetryAttemptsAfterError = 5
)

type LocalBridgeSync struct {
	processor *processor
	driver    *sync.EVMDriver
}

func NewL1(
	ctx context.Context,
	dbPath string,
	bridge common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd sync.ReorgDetector,
	ethClient EthClienter,
	initialBlock uint64,
) (*LocalBridgeSync, error) {
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
	)
}

func NewL2(
	ctx context.Context,
	dbPath string,
	bridge common.Address,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	rd sync.ReorgDetector,
	ethClient EthClienter,
	initialBlock uint64,
) (*LocalBridgeSync, error) {
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
) (*LocalBridgeSync, error) {
	processor, err := newProcessor(ctx, dbPath, dbPrefix)
	if err != nil {
		return nil, err
	}
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

// TODO: expose methods from the processor for consumers
