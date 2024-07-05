package localbridgesync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/reorgdetector"
)

const (
	checkReorgInterval = time.Second * 10
	downloadBufferSize = 100
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
	for {
		lastProcessedBlock, err := d.processor.getLastProcessedBlock(ctx)
		if err != nil {
			// TODO: handle error
			return
		}
		cancellableCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// start downloading
		downloadCh := make(chan block, downloadBufferSize)
		go d.downloader.download(cancellableCtx, lastProcessedBlock, downloadCh)

		for {
			if shouldRestartSync := d.syncIteration(ctx, cancel, downloadCh); shouldRestartSync {
				break
			}
		}
	}
}

func (d *driver) syncIteration(
	ctx context.Context, cancel context.CancelFunc, downloadCh chan block,
) (shouldRestartSync bool) {
	shouldRestartSync = false
	select {
	case b := <-downloadCh: // new block from downloader
		err := d.reorgDetector.AddBlockToTrack(ctx, reorgDetectorID, b.Num, b.Hash)
		if err != nil {
			// TODO: handle error
			return
		}
		err = d.processor.storeBridgeEvents(b.Num, b.Events)
		if err != nil {
			// TODO: handle error
			return
		}
	case lastValidBlock := <-d.reorgSub.FirstReorgedBlock: // reorg detected
		// stop downloader
		cancel()
		// wait until downloader closes channel
		_, ok := <-downloadCh
		for ok {
			_, ok = <-downloadCh
		}
		// handle reorg
		err := d.processor.reorg(lastValidBlock)
		if err != nil {
			// TODO: handle error
			return
		}
		d.reorgSub.ReorgProcessed <- true

		// restart syncing
		shouldRestartSync = true
	}
	return
}
