package sync

import (
	"context"
	"errors"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/ethereum/go-ethereum/common"
)

type evmDownloaderFull interface {
	EVMDownloaderInterface
	downloader
}

type downloader interface {
	Download(ctx context.Context, fromBlock uint64, downloadedCh chan EVMBlock)
}

type EVMDriver struct {
	reorgDetector      ReorgDetector
	reorgSub           *reorgdetector.Subscription
	processor          processorInterface
	downloader         downloader
	reorgDetectorID    string
	downloadBufferSize int
	rh                 *RetryHandler
	log                *log.Logger
}

type processorInterface interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
	ProcessBlock(ctx context.Context, block Block) error
	Reorg(ctx context.Context, firstReorgedBlock uint64) error
}

type ReorgDetector interface {
	Subscribe(id string) (*reorgdetector.Subscription, error)
	AddBlockToTrack(ctx context.Context, id string, blockNum uint64, blockHash common.Hash) error
}

func NewEVMDriver(
	reorgDetector ReorgDetector,
	processor processorInterface,
	downloader downloader,
	reorgDetectorID string,
	downloadBufferSize int,
	rh *RetryHandler,
) (*EVMDriver, error) {
	logger := log.WithFields("syncer", reorgDetectorID)
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
		rh:                 rh,
		log:                logger,
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
			d.log.Error("error getting last processed block: ", err)
			d.rh.Handle("Sync", attempts)
			continue
		}
		break
	}
	cancellableCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Info("Starting sync...", " lastProcessedBlock", lastProcessedBlock)
	// start downloading
	downloadCh := make(chan EVMBlock, d.downloadBufferSize)
	go d.downloader.Download(cancellableCtx, lastProcessedBlock+1, downloadCh)

	for {
		select {
		case <-ctx.Done():
			d.log.Info("sync stopped due to context done")
			cancel()
			return
		case b := <-downloadCh:
			d.log.Debugf("handleNewBlock, blockNum: %d, blockHash: %s", b.Num, b.Hash)
			d.handleNewBlock(ctx, cancel, b)
		case firstReorgedBlock := <-d.reorgSub.ReorgedBlock:
			d.log.Debug("handleReorg from block: ", firstReorgedBlock)
			d.handleReorg(ctx, cancel, firstReorgedBlock)
			goto reset
		}
	}
}

func (d *EVMDriver) handleNewBlock(ctx context.Context, cancel context.CancelFunc, b EVMBlock) {
	attempts := 0
	succeed := false
	for {
		select {
		case <-ctx.Done():
			// If the context is canceled, exit the function
			d.log.Warnf("context canceled while adding block %d to tracker", b.Num)
			return
		default:
			err := d.reorgDetector.AddBlockToTrack(ctx, d.reorgDetectorID, b.Num, b.Hash)
			if err != nil {
				attempts++
				d.log.Errorf("error adding block %d to tracker: %v", b.Num, err)
				d.rh.Handle("handleNewBlock", attempts)
			} else {
				succeed = true
			}
		}
		if succeed {
			break
		}
	}
	attempts = 0
	succeed = false
	for {
		select {
		case <-ctx.Done():
			// If the context is canceled, exit the function
			d.log.Warnf("context canceled while processing block %d", b.Num)
			return
		default:
			blockToProcess := Block{
				Num:    b.Num,
				Events: b.Events,
			}
			err := d.processor.ProcessBlock(ctx, blockToProcess)
			if err != nil {
				if errors.Is(err, ErrInconsistentState) {
					d.log.Warn("state got inconsistent after processing this block. Stopping downloader until there is a reorg")
					cancel()
					return
				}
				attempts++
				d.log.Errorf("error processing events for block %d, err: ", b.Num, err)
				d.rh.Handle("handleNewBlock", attempts)
			} else {
				succeed = true
			}
		}
		if succeed {
			break
		}
	}
}

func (d *EVMDriver) handleReorg(ctx context.Context, cancel context.CancelFunc, firstReorgedBlock uint64) {
	// stop downloader
	cancel()

	// handle reorg
	attempts := 0
	for {
		err := d.processor.Reorg(ctx, firstReorgedBlock)
		if err != nil {
			attempts++
			d.log.Errorf(
				"error processing reorg, last valid Block %d, err: %v",
				firstReorgedBlock, err,
			)
			d.rh.Handle("handleReorg", attempts)
			continue
		}
		break
	}
	d.reorgSub.ReorgProcessed <- true
}
