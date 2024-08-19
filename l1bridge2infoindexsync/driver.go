package l1bridge2infoindexsync

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
)

type driver struct {
	downloader           *downloader
	processor            *processor
	rh                   *sync.RetryHandler
	waitForSyncersPeriod time.Duration
}

func newDriver(
	downloader *downloader,
	processor *processor,
	rh *sync.RetryHandler,
	waitForSyncersPeriod time.Duration,
) *driver {
	return &driver{
		downloader:           downloader,
		processor:            processor,
		rh:                   rh,
		waitForSyncersPeriod: waitForSyncersPeriod,
	}
}

func (d *driver) sync(ctx context.Context) {
	var (
		attempts                 int
		lpbProcessor             uint64
		lastProcessedL1InfoIndex uint32
		err                      error
	)
	for {
		lpbProcessor, lastProcessedL1InfoIndex, err = d.processor.GetLastProcessedBlockAndL1InfoTreeIndex(ctx)
		if err != nil {
			attempts++
			log.Errorf("error getting last processed block and index: %v", err)
			d.rh.Handle("GetLastProcessedBlockAndL1InfoTreeIndex", attempts)
			continue
		}
		break
	}
	for {
		attempts = 0
		var (
			syncUntilBlock uint64
			shouldWait     bool
		)
		for {
			syncUntilBlock, shouldWait, err = d.getTargetSynchronizationBlock(ctx, lpbProcessor)
			if err != nil {
				attempts++
				log.Errorf("error getting target sync block: %v", err)
				d.rh.Handle("getTargetSynchronizationBlock", attempts)
				continue
			}
			break
		}
		if shouldWait {
			log.Debugf("waiting for syncers to catch up")
			time.Sleep(d.waitForSyncersPeriod)
			continue
		}

		attempts = 0
		var lastL1InfoTreeIndex uint32
		found := false
		for {
			lastL1InfoTreeIndex, err = d.downloader.getLastL1InfoIndexUntilBlock(ctx, syncUntilBlock)
			if err == l1infotreesync.ErrNotFound || err == l1infotreesync.ErrBlockNotProcessed {
				log.Debugf("l1 info tree index not ready, querying until block %d: %s", syncUntilBlock, err)
				break
			}
			if err != nil {
				attempts++
				log.Errorf("error getting last l1 info tree index: %v", err)
				d.rh.Handle("getLastL1InfoIndexUntilBlock", attempts)
				continue
			}
			found = true
			break
		}
		if !found {
			time.Sleep(d.waitForSyncersPeriod)
			continue
		}

		relations := []bridge2L1InfoRelation{}
		var init uint32
		if lastProcessedL1InfoIndex > 0 {
			init = lastProcessedL1InfoIndex + 1
		}
		if init <= lastL1InfoTreeIndex {
			log.Debugf("getting relations from index %d to %d", init, lastL1InfoTreeIndex)
		}
		for i := init; i <= lastL1InfoTreeIndex; i++ {
			attempts = 0
			for {
				relation, err := d.getRelation(ctx, i)
				if err != nil {
					attempts++
					log.Errorf("error getting relation: %v", err)
					d.rh.Handle("getRelation", attempts)
					continue
				}
				relations = append(relations, relation)
				break
			}
		}

		attempts = 0
		log.Debugf("processing until block %d: %+v", syncUntilBlock, relations)
		for {
			if err := d.processor.processUntilBlock(ctx, syncUntilBlock, relations); err != nil {
				attempts++
				log.Errorf("error processing block: %v", err)
				d.rh.Handle("processUntilBlock", attempts)
				continue
			}
			break
		}

		lpbProcessor = syncUntilBlock
		if len(relations) > 0 {
			lastProcessedL1InfoIndex = relations[len(relations)-1].l1InfoTreeIndex
			log.Debugf("last processed index %d", lastProcessedL1InfoIndex)
		}
	}
}

func (d *driver) getTargetSynchronizationBlock(ctx context.Context, lpbProcessor uint64) (syncUntilBlock uint64, shouldWait bool, err error) {
	lastFinalised, err := d.downloader.getLastFinalizedL1Block(ctx) // TODO: configure finality, but then we need to deal with reorgs?
	if err != nil {
		return
	}
	if lpbProcessor >= lastFinalised {
		log.Debugf(
			"should wait because the last processed block (%d) is greater or equal than the last finalised (%d)",
			lpbProcessor, lastFinalised,
		)
		shouldWait = true
		return
	}
	lpbInfo, err := d.downloader.getLastProcessedBlockL1InfoTree(ctx)
	if err != nil {
		return
	}
	if lpbProcessor >= lpbInfo {
		log.Debugf(
			"should wait because the last processed block (%d) is greater or equal than the last block from L1 Info tree sync (%d)",
			lpbProcessor, lpbInfo,
		)
		shouldWait = true
		return
	}
	lpbBridge, err := d.downloader.getLastProcessedBlockBridge(ctx)
	if err != nil {
		return
	}
	if lpbProcessor >= lpbBridge {
		log.Debugf(
			"should wait because the last processed block (%d) is greater or equal than the last block from l1 bridge sync (%d)",
			lpbProcessor, lpbBridge,
		)
		shouldWait = true
		return
	}

	// Bridge, L1Info and L1 ahead of procesor. Pick the smallest block num as target
	if lastFinalised <= lpbInfo {
		log.Debugf("target sync block is the last finalised block (%d)", lastFinalised)
		syncUntilBlock = lastFinalised
	} else {
		log.Debugf("target sync block is the last processed block from L1 info tree (%d)", lpbInfo)
		syncUntilBlock = lpbInfo
	}
	if lpbBridge < syncUntilBlock {
		log.Debugf("target sync block is the last processed block from bridge (%d)", lpbBridge)
		syncUntilBlock = lpbBridge
	}
	return
}

func (d *driver) getRelation(ctx context.Context, l1InfoIndex uint32) (bridge2L1InfoRelation, error) {
	mer, err := d.downloader.getMainnetExitRootAtL1InfoTreeIndex(ctx, l1InfoIndex)
	if err != nil {
		return bridge2L1InfoRelation{}, err
	}

	bridgeIndex, err := d.downloader.getBridgeIndex(ctx, mer)
	if err != nil {
		return bridge2L1InfoRelation{}, err
	}

	return bridge2L1InfoRelation{
		bridgeIndex:     bridgeIndex,
		l1InfoTreeIndex: l1InfoIndex,
	}, nil
}
