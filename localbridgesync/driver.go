package localbridgesync

import (
	"context"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
)

const (
	downloadBufferSize = 1000
	reorgDetectorID    = "localbridgesync"
)

type driver struct {
	reorgDetector *reorgdetector.ReorgDetector
	reorgSub      *reorgdetector.Subscription
	processor     *processor
	downloader    *downloader
}

func newDriver(
	reorgDetector *reorgdetector.ReorgDetector,
	processor *processor,
	downloader *downloader,
) (*driver, error) {
	reorgSub := reorgDetector.Subscribe(reorgDetectorID)
	return &driver{
		reorgDetector: reorgDetector,
		reorgSub:      reorgSub,
		processor:     processor,
		downloader:    downloader,
	}, nil
}

func (d *driver) Sync(ctx context.Context) {
	attempts := 0
	for {
		lastProcessedBlock, err := d.processor.getLastProcessedBlock(ctx)
		if err != nil {
			attempts++
			log.Error("error geting last processed block: ", err)
			retryHandler("Sync", attempts)
			continue
		}
		attempts = 0
		cancellableCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// start downloading
		downloadCh := make(chan block, downloadBufferSize)
		go download(cancellableCtx, d.downloader, lastProcessedBlock, 100, downloadCh) // TODO: 100 should come from config

		for {
			select {
			case b := <-downloadCh:
				d.handleNewBlock(ctx, b)
			case firstReorgedBlock := <-d.reorgSub.FirstReorgedBlock:
				d.handleReorg(cancel, downloadCh, firstReorgedBlock)
				break
			}
		}
	}
}

func (d *driver) handleNewBlock(ctx context.Context, b block) {
	attempts := 0
	for {
		err := d.reorgDetector.AddBlockToTrack(ctx, reorgDetectorID, b.Num, b.Hash)
		if err != nil {
			attempts++
			log.Errorf("error adding block %d to tracker: %v", b.Num, err)
			retryHandler("handleNewBlock", attempts)
			continue
		}
		break
	}
	attempts = 0
	for {
		err := d.processor.storeBridgeEvents(b.Num, b.Events)
		if err != nil {
			attempts++
			log.Errorf("error processing events for blcok %d, err: ", b.Num, err)
			retryHandler("handleNewBlock", attempts)
			continue
		}
		break
	}
}

func (d *driver) handleReorg(
	cancel context.CancelFunc, downloadCh chan block, firstReorgedBlock uint64,
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
		err := d.processor.reorg(firstReorgedBlock)
		if err != nil {
			attempts++
			log.Errorf(
				"error processing reorg, last valid block %d, err: %v",
				firstReorgedBlock, err,
			)
			retryHandler("handleReorg", attempts)
			continue
		}
		break
	}
	d.reorgSub.ReorgProcessed <- true
}
