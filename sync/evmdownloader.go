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
	"github.com/ethereum/go-ethereum/rpc"
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
	GetEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) EVMBlocks
	GetLogs(ctx context.Context, fromBlock, toBlock uint64) []types.Log
	GetBlockHeader(ctx context.Context, blockNum uint64) (EVMBlockHeader, bool)
	GetLastFinalizedBlock(ctx context.Context) (*types.Header, error)
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
			d.log.Infof(
				"waiting for new blocks, last block processed: %d, last block seen on L1: %d",
				fromBlock-1, lastBlock,
			)
			lastBlock = d.WaitForNewBlocks(ctx, fromBlock-1)
			continue
		}

		lastFinalizedBlock, err := d.GetLastFinalizedBlock(ctx)
		if err != nil {
			d.log.Error("error getting last finalized block: ", err)
			continue
		}

		lastFinalizedBlockNumber := lastFinalizedBlock.Number.Uint64()

		d.log.Infof("getting events from blocks %d to  %d. lastFinalizedBlock: %d",
			fromBlock, toBlock, lastFinalizedBlockNumber)
		blocks := d.GetEventsByBlockRange(ctx, fromBlock, toBlock)

		reportBlocksFn := func(numOfBlocksToReport int) {
			for i := 0; i < numOfBlocksToReport; i++ {
				d.log.Infof("sending block %d to the driver (with events)", blocks[i].Num)
				downloadedCh <- blocks[i]
			}
		}

		reportEmptyBlockFn := func(blockNum uint64) {
			// Indicate the last downloaded block if there are not events on it
			d.log.Debugf("sending block %d to the driver (without events)", toBlock)
			header, isCanceled := d.GetBlockHeader(ctx, blockNum)
			if isCanceled {
				return
			}

			downloadedCh <- EVMBlock{
				EVMBlockHeader: header,
			}
		}

		if blocks.Len() == 0 {
			// we have no events, keep increasing the block range until we hit a log
			d.log.Infof("no events found in blocks %d to %d", fromBlock, toBlock)
			if lastFinalizedBlockNumber > toBlock {
				// we might be behind a lot, so go until last finalized block
				toBlock = lastFinalizedBlockNumber
				lastBlock = lastFinalizedBlockNumber
				d.log.Infof("setting toBlock to last finalized block %d", toBlock)
			}

			if lastFinalizedBlockNumber > fromBlock && lastFinalizedBlockNumber-fromBlock >= d.syncBlockChunkSize {
				// if we already got a lot of finalized blocks that are empty, report an empty block
				// to the driver to indicate that we are still processing the chain
				// this is mainly needed for tests
				reportEmptyBlockFn(lastFinalizedBlockNumber)
				fromBlock = lastFinalizedBlockNumber
				d.log.Infof("setting fromBlock to last finalized block %d", fromBlock)
			}

			continue
		} else if blocks[blocks.Len()-1].Num <= lastFinalizedBlockNumber {
			// if the last block we have logs for is less than or equal to the last finalized block,
			// report all of the blocks without the need to report the last empty block, since it is finalized
			// and we do not need to track it in the reorg detector
			reportBlocksFn(blocks.Len())
			fromBlock = toBlock + 1
			d.log.Infof("got blocks that are lower than finalized block, setting fromBlock to %d", fromBlock)
		} else if blocks[blocks.Len()-1].Num < toBlock {
			// if we have logs in some of the blocks, and they are not all finalized,
			// check if we have finalized blocks in gotten range, report them and
			// set the from block from the last finalized block and keep increasing the range
			// if not keep getting that range to protect us from possible mishandling of block hashes
			lastFinalizedBlock, index, exists := blocks.LastFinalizedBlock(lastFinalizedBlockNumber)
			if exists {
				reportBlocksFn(index + 1) // num of blocks to report is index + 1 since index is zero based
				fromBlock = lastFinalizedBlock + 1
				d.log.Infof("have some finalized blocks in the range, setting fromBlock to %d", fromBlock)
				continue
			}
		} else {
			// if we have logs in the last block, just report all of them and continue
			// reorg detector will handle the reorg since the last block has events,
			// and we are not afraid to have missaligned hashes at this point
			reportBlocksFn(blocks.Len())
			fromBlock = toBlock + 1
			d.log.Infof("have logs in the last block, setting fromBlock to %d", fromBlock)
		}
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

func (d *EVMDownloaderImplementation) GetLastFinalizedBlock(ctx context.Context) (*types.Header, error) {
	return d.ethClient.HeaderByNumber(ctx, big.NewInt(int64(rpc.SafeBlockNumber)))
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

func (d *EVMDownloaderImplementation) GetEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) EVMBlocks {
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
