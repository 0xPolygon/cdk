package sync

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	DefaultWaitPeriodBlockNotFound = time.Millisecond * 100
)

type EthClienter interface {
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainReader
	bind.ContractBackend
}

type EVMDownloaderInterface interface {
	WaitForNewBlocks(ctx context.Context, lastBlockSeen uint64) (newLastBlock uint64)
	GetEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) []EVMBlock
	GetLogs(ctx context.Context, fromBlock, toBlock uint64) []types.Log
	GetBlockHeader(ctx context.Context, blockNum uint64) (EVMBlockHeader, bool)
}

type LogAppenderMap map[common.Hash]func(b *EVMBlock, l types.Log) error

type EVMDownloader struct {
	syncBlockChunkSize uint64
	EVMDownloaderInterface
	log *log.Logger
}

func NewEVMDownloader(
	syncerID string,
	ethClient EthClienter,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	waitForNewBlocksPeriod time.Duration,
	appender LogAppenderMap,
	adressessToQuery []common.Address,
	rh *RetryHandler,
) (*EVMDownloader, error) {
	logger := log.WithFields("syncer", syncerID)
	finality, err := blockFinalityType.ToBlockNum()
	if err != nil {
		return nil, err
	}
	topicsToQuery := make([]common.Hash, 0, len(appender))
	for topic := range appender {
		topicsToQuery = append(topicsToQuery, topic)
	}
	return &EVMDownloader{
		syncBlockChunkSize: syncBlockChunkSize,
		log:                logger,
		EVMDownloaderInterface: &EVMDownloaderImplementation{
			ethClient:              ethClient,
			blockFinality:          finality,
			waitForNewBlocksPeriod: waitForNewBlocksPeriod,
			appender:               appender,
			topicsToQuery:          topicsToQuery,
			adressessToQuery:       adressessToQuery,
			rh:                     rh,
			log:                    logger,
		},
	}, nil
}

func (d *EVMDownloader) Download(ctx context.Context, fromBlock uint64, downloadedCh chan EVMBlock) {
	lastBlock := d.WaitForNewBlocks(ctx, 0)
	for {
		select {
		case <-ctx.Done():
			d.log.Debug("closing channel")
			close(downloadedCh)
			return
		default:
		}
		toBlock := fromBlock + d.syncBlockChunkSize
		if toBlock > lastBlock {
			toBlock = lastBlock
		}
		if fromBlock > toBlock {
			d.log.Debugf(
				"waiting for new blocks, last block processed: %d, last block seen on L1: %d",
				fromBlock-1, lastBlock,
			)
			lastBlock = d.WaitForNewBlocks(ctx, fromBlock-1)
			continue
		}
		d.log.Debugf("getting events from block %d to %d", fromBlock, toBlock)
		blocks := d.GetEventsByBlockRange(ctx, fromBlock, toBlock)
		for _, b := range blocks {
			d.log.Debugf("sending block %d to the driver (with events)", b.Num)
			downloadedCh <- b
		}
		if len(blocks) == 0 || blocks[len(blocks)-1].Num < toBlock {
			// Indicate the last downloaded block if there are not events on it
			d.log.Debugf("sending block %d to the driver (without events)", toBlock)
			header, isCanceled := d.GetBlockHeader(ctx, toBlock)
			if isCanceled {
				return
			}

			downloadedCh <- EVMBlock{
				EVMBlockHeader: header,
			}
		}
		fromBlock = toBlock + 1
	}
}

type EVMDownloaderImplementation struct {
	ethClient              EthClienter
	blockFinality          *big.Int
	waitForNewBlocksPeriod time.Duration
	appender               LogAppenderMap
	topicsToQuery          []common.Hash
	adressessToQuery       []common.Address
	rh                     *RetryHandler
	log                    *log.Logger
}

func NewEVMDownloaderImplementation(
	syncerID string,
	ethClient EthClienter,
	blockFinality *big.Int,
	waitForNewBlocksPeriod time.Duration,
	appender LogAppenderMap,
	topicsToQuery []common.Hash,
	adressessToQuery []common.Address,
	rh *RetryHandler,
) *EVMDownloaderImplementation {
	logger := log.WithFields("syncer", syncerID)
	return &EVMDownloaderImplementation{
		ethClient:              ethClient,
		blockFinality:          blockFinality,
		waitForNewBlocksPeriod: waitForNewBlocksPeriod,
		appender:               appender,
		topicsToQuery:          topicsToQuery,
		adressessToQuery:       adressessToQuery,
		rh:                     rh,
		log:                    logger,
	}
}

func (d *EVMDownloaderImplementation) WaitForNewBlocks(
	ctx context.Context, lastBlockSeen uint64,
) (newLastBlock uint64) {
	attempts := 0
	ticker := time.NewTicker(d.waitForNewBlocksPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			d.log.Info("context cancelled")
			return lastBlockSeen
		case <-ticker.C:
			header, err := d.ethClient.HeaderByNumber(ctx, d.blockFinality)
			if err != nil {
				if ctx.Err() == nil {
					attempts++
					d.log.Error("error getting last block num from eth client: ", err)
					d.rh.Handle("waitForNewBlocks", attempts)
				} else {
					d.log.Warn("context has been canceled while trying to get header by number")
				}
				continue
			}
			if header.Number.Uint64() > lastBlockSeen {
				return header.Number.Uint64()
			}
		}
	}
}

func (d *EVMDownloaderImplementation) GetEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) []EVMBlock {
	select {
	case <-ctx.Done():
		return nil
	default:
		blocks := []EVMBlock{}
		logs := d.GetLogs(ctx, fromBlock, toBlock)
		for _, l := range logs {
			if len(blocks) == 0 || blocks[len(blocks)-1].Num < l.BlockNumber {
				b, canceled := d.GetBlockHeader(ctx, l.BlockNumber)
				if canceled {
					return nil
				}

				if b.Hash != l.BlockHash {
					d.log.Infof(
						"there has been a block hash change between the event query and the block query "+
							"for block %d: %s vs %s. Retrying.",
						l.BlockNumber, b.Hash, l.BlockHash,
					)
					return d.GetEventsByBlockRange(ctx, fromBlock, toBlock)
				}
				blocks = append(blocks, EVMBlock{
					EVMBlockHeader: EVMBlockHeader{
						Num:        l.BlockNumber,
						Hash:       l.BlockHash,
						Timestamp:  b.Timestamp,
						ParentHash: b.ParentHash,
					},
					Events: []interface{}{},
				})
			}

			for {
				attempts := 0
				err := d.appender[l.Topics[0]](&blocks[len(blocks)-1], l)
				if err != nil {
					attempts++
					d.log.Error("error trying to append log: ", err)
					d.rh.Handle("getLogs", attempts)
					continue
				}
				break
			}
		}

		return blocks
	}
}

func filterQueryToString(query ethereum.FilterQuery) string {
	return fmt.Sprintf("FromBlock: %s, ToBlock: %s, Addresses: %s, Topics: %s",
		query.FromBlock.String(), query.ToBlock.String(), query.Addresses, query.Topics)
}

func (d *EVMDownloaderImplementation) GetLogs(ctx context.Context, fromBlock, toBlock uint64) []types.Log {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		Addresses: d.adressessToQuery,
		ToBlock:   new(big.Int).SetUint64(toBlock),
	}
	var (
		attempts       = 0
		unfilteredLogs []types.Log
		err            error
	)
	for {
		unfilteredLogs, err = d.ethClient.FilterLogs(ctx, query)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				// context is canceled, we don't want to fatal on max attempts in this case
				return nil
			}

			attempts++
			d.log.Errorf("error calling FilterLogs to eth client: filter: %s err:%w ",
				filterQueryToString(query),
				err,
			)
			d.rh.Handle("getLogs", attempts)
			continue
		}
		break
	}
	logs := make([]types.Log, 0, len(unfilteredLogs))
	for _, l := range unfilteredLogs {
		for _, topic := range d.topicsToQuery {
			if l.Topics[0] == topic {
				logs = append(logs, l)
				break
			}
		}
	}
	return logs
}

func (d *EVMDownloaderImplementation) GetBlockHeader(ctx context.Context, blockNum uint64) (EVMBlockHeader, bool) {
	attempts := 0
	for {
		header, err := d.ethClient.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNum))
		if err != nil {
			if errors.Is(err, context.Canceled) {
				// context is canceled, we don't want to fatal on max attempts in this case
				return EVMBlockHeader{}, true
			}
			if errors.Is(err, ethereum.NotFound) {
				// block num can temporary disappear from the execution client due to a reorg,
				// in this case, we want to wait and not panic
				log.Warnf("block %d not found on the ethereum client: %v", blockNum, err)
				if d.rh.RetryAfterErrorPeriod != 0 {
					time.Sleep(d.rh.RetryAfterErrorPeriod)
				} else {
					time.Sleep(DefaultWaitPeriodBlockNotFound)
				}
				continue
			}

			attempts++
			d.log.Errorf("error getting block header for block %d, err: %v", blockNum, err)
			d.rh.Handle("getBlockHeader", attempts)
			continue
		}
		return EVMBlockHeader{
			Num:        header.Number.Uint64(),
			Hash:       header.Hash(),
			ParentHash: header.ParentHash,
			Timestamp:  header.Time,
		}, false
	}
}
