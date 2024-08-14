package zkevmbridge2infoindexsync

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
)

type downloader struct {
	*sync.EVMDownloaderImplementation
	zkevmClient            sync.ZKEVMClientInterface
	l2Client               sync.EthClienter
	rh                     *sync.RetryHandler
	l2BridgeSync           *bridgesync.BridgeSync
	waitForNewBlocksPeriod time.Duration
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
		var (
			attempts int
			batchNum uint64
			err      error
		)
		// wait for block to be verified
		for {
			isVerified, err := d.zkevmClient.
		}
		attempts = 0
		// get batch num associated to the block num
		for {
			batchNum, err = d.zkevmClient.BatchNumberByBlockNumber(ctx, fromBlock)
			if err != nil {
				attempts++
				log.Error(err)
				d.rh.Handle(fmt.Sprintf("d.zkevmClient.BatchNumberByBlockNumber(ctx, &d)", fromBlock), attempts)
				continue
			}
			break
		}
		// get bridges in block
		// get batch -> index
		attempts = 0
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
