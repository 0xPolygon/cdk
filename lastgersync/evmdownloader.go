package lastgersync

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/l2-sovereign-chain-paris/globalexitrootmanagerl2sovereignchain"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type EthClienter interface {
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainReader
	bind.ContractBackend
}

type downloader struct {
	*sync.EVMDownloaderImplementation
	l2GERManager   *globalexitrootmanagerl2sovereignchain.Globalexitrootmanagerl2sovereignchain
	l1InfoTreesync *l1infotreesync.L1InfoTreeSync
	processor      *processor
	rh             *sync.RetryHandler
}

func newDownloader(
	l2Client EthClienter,
	l2GERAddr common.Address,
	l1InfoTreeSync *l1infotreesync.L1InfoTreeSync,
	processor *processor,
	rh *sync.RetryHandler,
	blockFinality *big.Int,
	waitForNewBlocksPeriod time.Duration,
) (*downloader, error) {
	gerContract, err := globalexitrootmanagerl2sovereignchain.NewGlobalexitrootmanagerl2sovereignchain(
		l2GERAddr, l2Client)
	if err != nil {
		return nil, err
	}

	return &downloader{
		EVMDownloaderImplementation: sync.NewEVMDownloaderImplementation(
			"lastgersync", l2Client, blockFinality, waitForNewBlocksPeriod, nil, nil, nil, rh,
		),
		l2GERManager:   gerContract,
		l1InfoTreesync: l1InfoTreeSync,
		processor:      processor,
		rh:             rh,
	}, nil
}

func (d *downloader) Download(ctx context.Context, fromBlock uint64, downloadedCh chan sync.EVMBlock) {
	var (
		attempts  int
		nextIndex uint32
		err       error
	)
	for {
		lastIndex, err := d.processor.getLastIndex()
		if errors.Is(err, db.ErrNotFound) {
			nextIndex = 0
		} else if err != nil {
			log.Errorf("error getting last indes: %v", err)
			attempts++
			d.rh.Handle("getLastIndex", attempts)

			continue
		}
		if lastIndex > 0 {
			nextIndex = lastIndex + 1
		}
		break
	}
	for {
		select {
		case <-ctx.Done():
			log.Debug("closing channel")
			close(downloadedCh)

			return
		default:
		}
		fromBlock = d.WaitForNewBlocks(ctx, fromBlock)

		attempts = 0
		var gers []Event
		for {
			gers, err = d.getGERsFromIndex(ctx, nextIndex)
			if err != nil {
				log.Errorf("error getting GERs: %v", err)
				attempts++
				d.rh.Handle("getGERsFromIndex", attempts)

				continue
			}

			break
		}

		blockHeader, isCanceled := d.GetBlockHeader(ctx, fromBlock)
		if isCanceled {
			return
		}

		block := &sync.EVMBlock{
			EVMBlockHeader: sync.EVMBlockHeader{
				Num:        blockHeader.Num,
				Hash:       blockHeader.Hash,
				ParentHash: blockHeader.ParentHash,
				Timestamp:  blockHeader.Timestamp,
			},
		}
		d.setGreatestGERInjectedFromList(block, gers)

		downloadedCh <- *block
		if len(block.Events) > 0 {
			event, ok := block.Events[0].(Event)
			if !ok {
				log.Errorf("unexpected type %T in events", block.Events[0])
			}
			nextIndex = event.L1InfoTreeIndex + 1
		}
	}
}

func (d *downloader) getGERsFromIndex(ctx context.Context, fromL1InfoTreeIndex uint32) ([]Event, error) {
	lastRoot, err := d.l1InfoTreesync.GetLastL1InfoTreeRoot(ctx)
	if errors.Is(err, db.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error calling GetLastL1InfoTreeRoot: %w", err)
	}

	gers := []Event{}
	for i := fromL1InfoTreeIndex; i <= lastRoot.Index; i++ {
		info, err := d.l1InfoTreesync.GetInfoByIndex(ctx, i)
		if err != nil {
			return nil, fmt.Errorf("error calling GetInfoByIndex: %w", err)
		}
		gers = append(gers, Event{
			L1InfoTreeIndex: i,
			GlobalExitRoot:  info.GlobalExitRoot,
		})
	}

	return gers, nil
}

func (d *downloader) setGreatestGERInjectedFromList(b *sync.EVMBlock, list []Event) {
	for _, event := range list {
		var attempts int
		for {
			blockHashBigInt, err := d.l2GERManager.GlobalExitRootMap(&bind.CallOpts{Pending: false}, event.GlobalExitRoot)
			if err != nil {
				attempts++
				log.Errorf("failed to check if global exit root %s is injected on L2: %s", event.GlobalExitRoot.Hex(), err)
				d.rh.Handle("GlobalExitRootMap", attempts)

				continue
			}

			if common.BigToHash(blockHashBigInt) != cdkcommon.ZeroHash {
				b.Events = []interface{}{event}
			}

			break
		}
	}
}
