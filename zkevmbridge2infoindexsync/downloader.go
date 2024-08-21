package zkevmbridge2infoindexsync

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonrollupmanager"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
)

type downloader struct {
	*sync.EVMDownloaderImplementation
	zkevmClient                    sync.ZKEVMClientInterface
	l2Client                       sync.EthClienter
	rh                             *sync.RetryHandler
	l2BridgeSync                   *bridgesync.BridgeSync
	waitForNewBlocksPeriod         time.Duration
	waitForBatchToBeVerifiedPeriod time.Duration
	rollupManager                  *polygonrollupmanager.Polygonrollupmanager
	rollupID                       uint32
}

func newDownloader(
	zkevmClient sync.ZKEVMClientInterface,
	l2Client sync.EthClienter,
	blockFinality *big.Int,
	waitForNewBlocksPeriod time.Duration,
	rh *sync.RetryHandler,
	l2BridgeSync *bridgesync.BridgeSync,
) *downloader {
	return &downloader{
		EVMDownloaderImplementation: sync.NewEVMDownloaderImplementation(
			"lastgersync", l2Client, blockFinality, waitForNewBlocksPeriod, nil, nil, nil, rh,
		),
		rh:                     rh,
		zkevmClient:            zkevmClient,
		l2Client:               l2Client,
		l2BridgeSync:           l2BridgeSync,
		waitForNewBlocksPeriod: waitForNewBlocksPeriod,
	}
}

func (d *downloader) Download(ctx context.Context, fromBlock uint64, downloadedCh chan sync.EVMBlock) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("closing channel")
			close(downloadedCh)
			return
		default:
		}
		if fromBlock != 0 {
			d.WaitForNewBlocks(ctx, fromBlock-1)
		}
		// get depositCounts in block
		depositCounts := d.getDepositCountsInBlock(ctx, fromBlock)
		if len(depositCounts) > 0 {
			// get batch num associated to the block num
			batchNum := d.blockNumToBatchNum(ctx, fromBlock)
			// wait for block to be verified
			d.waitForBatchToBeVerified(ctx, batchNum)
			// get batch -> index
		}

		blockHeader := d.GetBlockHeader(ctx, fromBlock)
		downloadedCh <- sync.EVMBlock{
			EVMBlockHeader: sync.EVMBlockHeader{
				Num:        blockHeader.Num,
				Hash:       blockHeader.Hash,
				ParentHash: blockHeader.ParentHash,
				Timestamp:  blockHeader.Timestamp,
			},
			Events: []interface{}{batchNum},
		}
		fromBlock++
	}
}

func (d *downloader) getDepositCountsInBlock(ctx context.Context, blockNum uint64) []uint32 {
	var (
		attempts int
		blocks   []bridgesync.EventsWithBlock
		err      error
	)
	for {
		blocks, err = d.l2BridgeSync.GetClaimsAndBridges(ctx, blockNum, blockNum+1)
		if err != nil {
			attempts++
			log.Error(err)
			d.rh.Handle(fmt.Sprintf("d.l2BridgeSync.GetClaimsAndBridges(ctx, %d, %d)", blockNum, blockNum), attempts)
			continue
		}
		break
	}
	if len(blocks) == 0 {
		// block doesnt have events
		return nil
	}
	if len(blocks) > 1 {
		log.Fatalf("unexpected amount of returned blocks. Expected 1 or 0, actual %d", len(blocks))
	}
	depositCounts := []uint32{}
	for _, event := range blocks[0].Events {
		if event.Bridge != nil {
			depositCounts = append(depositCounts, event.Bridge.DepositCount)
		}
	}
	return depositCounts
}

func (d *downloader) blockNumToBatchNum(ctx context.Context, blockNum uint64) uint64 {
	attempts := 0
	for {
		batchNum, err := d.zkevmClient.BatchNumberByBlockNumber(ctx, blockNum)
		if err != nil {
			attempts++
			log.Error(err)
			d.rh.Handle(fmt.Sprintf("d.zkevmClient.BatchNumberByBlockNumber(ctx, %d)", blockNum), attempts)
			continue
		}
		return batchNum
	}
}

func (d *downloader) waitForBatchToBeVerified(ctx context.Context, batchNum uint64) {
	attempts := 0
	for {
		lastVerifiedBatchNum, err := d.rollupManager.GetLastVerifiedBatch(nil, d.rollupID)
		if err != nil {
			attempts++
			log.Error(err)
			d.rh.Handle(fmt.Sprintf("d.rollupManager.GetLastVerifiedBatch(nil, d.rollupID)"), attempts)
			continue
		}
		attempts = 0
		if lastVerifiedBatchNum > batchNum {
			log.Infof("batch %d has not been verified yet, waiting...", batchNum)
			time.Sleep(d.waitForBatchToBeVerifiedPeriod)
			continue
		}
		return
	}
}
