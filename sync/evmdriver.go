package sync

import (
	"context"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/ethereum/go-ethereum/common"
)

type evmDownloaderFull interface {
	evmDownloaderInterface
	download(ctx context.Context, fromBlock uint64, downloadedCh chan EVMBlock)
}

type EVMDriver struct {
	reorgDetector      ReorgDetector
	reorgSub           *reorgdetector.Subscription
	processor          processorInterface
	downloader         evmDownloaderFull
	reorgDetectorID    string
	downloadBufferSize int
}

type processorInterface interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
	ProcessBlock(block EVMBlock) error
	Reorg(firstReorgedBlock uint64) error
}

type ReorgDetector interface {
	Subscribe(id string) (*reorgdetector.Subscription, error)
	AddBlockToTrack(ctx context.Context, id string, blockNum uint64, blockHash common.Hash) error
}

func NewEVMDriver(
	reorgDetector ReorgDetector,
	processor processorInterface,
	downloader evmDownloaderFull,
	reorgDetectorID string,
	downloadBufferSize int,
) (*EVMDriver, error) {
	reorgSub, err := reorgDetector.Subscribe(reorgDetectorID)
	if err != nil {
		return nil, err
	}
	return &EVMDriver{
		reorgDetector:      reorgDetector,
		reorgSub:           reorgSub,
		processor:          processor,
		downloader:         downloader,
		reorgDetectorID:    reorgDetectorID,
		downloadBufferSize: downloadBufferSize,
	}, nil
}

func (d *EVMDriver) Sync(ctx context.Context) {
reset:
	var (
		lastProcessedBlock uint64
		attempts           int
		err                error
	)
	for {
		lastProcessedBlock, err = d.processor.GetLastProcessedBlock(ctx)
		if err != nil {
			attempts++
			log.Error("error geting last processed block: ", err)
			RetryHandler("Sync", attempts)
			continue
		}
		break
	}
	cancellableCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// start downloading
	downloadCh := make(chan EVMBlock, d.downloadBufferSize)
	go d.downloader.download(cancellableCtx, lastProcessedBlock, downloadCh)

	for {
		select {
		case b := <-downloadCh:
			log.Debug("handleNewBlock")
			d.handleNewBlock(ctx, b)
		case firstReorgedBlock := <-d.reorgSub.FirstReorgedBlock:
			log.Debug("handleReorg")
			d.handleReorg(cancel, downloadCh, firstReorgedBlock)
			goto reset
		}
	}
}

func (d *EVMDriver) handleNewBlock(ctx context.Context, b EVMBlock) {
	attempts := 0
	for {
		err := d.reorgDetector.AddBlockToTrack(ctx, d.reorgDetectorID, b.Num, b.Hash)
		if err != nil {
			attempts++
			log.Errorf("error adding block %d to tracker: %v", b.Num, err)
			RetryHandler("handleNewBlock", attempts)
			continue
		}
		break
	}
	attempts = 0
	for {
		err := d.processor.ProcessBlock(b)
		if err != nil {
			attempts++
			log.Errorf("error processing events for blcok %d, err: ", b.Num, err)
			RetryHandler("handleNewBlock", attempts)
			continue
		}
		break
	}
}

func (d *EVMDriver) handleReorg(
	cancel context.CancelFunc, downloadCh chan EVMBlock, firstReorgedBlock uint64,
) {
	// stop downloader
	cancel()
	_, ok := <-downloadCh
	for ok {
		_, ok = <-downloadCh
	}
	// handle reorg
	attempts := 0
	for {
		err := d.processor.Reorg(firstReorgedBlock)
		if err != nil {
			attempts++
			log.Errorf(
				"error processing reorg, last valid Block %d, err: %v",
				firstReorgedBlock, err,
			)
			RetryHandler("handleReorg", attempts)
			continue
		}
		break
	}
	d.reorgSub.ReorgProcessed <- true
}
