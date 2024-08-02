package l1bridge2infoindexsync

import (
	"context"

	"github.com/0xPolygon/cdk/log"
)

type driver struct {
	downloader *downloader
	processor  *processor
}

func newDriver(
	downloader *downloader,
	processor *processor) *driver {
	return &driver{
		downloader: downloader,
		processor:  processor,
	}
}

func (d *driver) sync(ctx context.Context) {
	lpbProcessor, err := d.processor.GetLastProcessedBlock(ctx)
	if err != nil {
		log.Fatal("TODO")
	}
	lastProcessedL1InfoIndex, err := d.processor.getLastL1InfoTreeIndexProcessed(ctx)
	if err != nil {
		log.Fatal("TODO")
	}
	for {
		syncUntilBlock, shouldWait, err := d.getTargetSynchronizationBlock(ctx, lpbProcessor)
		if err != nil {
			log.Fatal("TODO")
		}
		if shouldWait {
			// TODO: wait using ticker
			continue
		}
		lastL1InfoTreeIndex, err := d.downloader.getLastL1InfoIndexUntilBlock(ctx, syncUntilBlock)
		if err != nil {
			log.Fatal("TODO")
		}
		relations := []bridge2L1InfoRelation{}
		for i := lastProcessedL1InfoIndex + 1; i <= lastL1InfoTreeIndex; i++ {
			relation, err := d.getRelation(ctx, i)
			if err != nil {
				log.Fatal("TODO")
			}
			relations = append(relations, relation)
		}
		if err := d.processor.addBridge2L1InfoRelations(ctx, syncUntilBlock, relations); err != nil {
			log.Fatal("TODO")
		}
		lpbProcessor = syncUntilBlock
		lastProcessedL1InfoIndex = 0 // TODO
	}
}

func (d *driver) getTargetSynchronizationBlock(ctx context.Context, lpbProcessor uint64) (syncUntilBlock uint64, shouldWait bool, err error) {
	lastFinalised, err := d.downloader.getLastFinalisedL1Block(ctx) // TODO: configure finality, but then we need to deal with reorgs?
	if err != nil {
		return
	}
	if lpbProcessor >= lastFinalised {
		shouldWait = true
		return
	}
	lpbInfo, err := d.downloader.getLastProcessedBlockL1InfoTree(ctx)
	if err != nil {
		return
	}
	if lpbProcessor >= lpbInfo {
		shouldWait = true
		return
	}
	lpbBridge, err := d.downloader.getLastProcessedBlockBridge(ctx)
	if err != nil {
		return
	}
	if lpbProcessor >= lpbBridge {
		shouldWait = true
		return
	}

	// Bridge, L1Info and L1 ahead of procesor. Pick the smallest block num as target
	if lastFinalised <= lpbInfo {
		syncUntilBlock = lastFinalised
	} else {
		syncUntilBlock = lpbInfo
	}
	if lpbBridge < syncUntilBlock {
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
