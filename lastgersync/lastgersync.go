package lastgersync

import (
	"context"
	"math/big"
	"time"

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
	rh *sync.RetryHandler,
	blockFinality *big.Int,
	waitForNewBlocksPeriod time.Duration,
	downloadBufferSize int,
) (*LastGERSync, error) {
	processor, err := newProcessor(dbPath)
	if err != nil {
		return nil, err
	}

	downloader, err := newDownloader(
		l2Client,
		globalExitRootL2,
		l1InfoTreesync,
		processor,
		rh,
		blockFinality,
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
