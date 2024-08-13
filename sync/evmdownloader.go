package sync

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	GetBlockHeader(ctx context.Context, blockNum uint64) EVMBlockHeader
}

type LogAppenderMap map[common.Hash]func(b *EVMBlock, l types.Log) error

type EVMDownloader struct {
	syncBlockChunkSize uint64
	EVMDownloaderInterface
}

func NewEVMDownloader(
	ethClient EthClienter,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	waitForNewBlocksPeriod time.Duration,
	appender LogAppenderMap,
	adressessToQuery []common.Address,
	rh *RetryHandler,
) (*EVMDownloader, error) {
	finality, err := blockFinalityType.ToBlockNum()
	if err != nil {
		return nil, err
	}
	topicsToQuery := []common.Hash{}
	for topic := range appender {
		topicsToQuery = append(topicsToQuery, topic)
	}
	return &EVMDownloader{
		syncBlockChunkSize: syncBlockChunkSize,
		EVMDownloaderInterface: &EVMDownloaderImplementation{
			ethClient:              ethClient,
			blockFinality:          finality,
			waitForNewBlocksPeriod: waitForNewBlocksPeriod,
			appender:               appender,
			topicsToQuery:          topicsToQuery,
			adressessToQuery:       adressessToQuery,
			rh:                     rh,
		},
	}, nil
}

func (d *EVMDownloader) Download(ctx context.Context, fromBlock uint64, downloadedCh chan EVMBlock) {
	lastBlock := d.WaitForNewBlocks(ctx, 0)
	for {
		select {
		case <-ctx.Done():
			log.Debug("closing channel")
			close(downloadedCh)
			return
		default:
		}
		toBlock := fromBlock + d.syncBlockChunkSize
		if toBlock > lastBlock {
			toBlock = lastBlock
		}
		if fromBlock > toBlock {
			log.Debug("waiting for new blocks, last block ", toBlock)
			lastBlock = d.WaitForNewBlocks(ctx, toBlock)
			continue
		}
		log.Debugf("getting events from blocks %d to  %d", fromBlock, toBlock)
		blocks := d.GetEventsByBlockRange(ctx, fromBlock, toBlock)
		for _, b := range blocks {
			log.Debugf("sending block %d to the driver (with events)", b.Num)
			downloadedCh <- b
		}
		if len(blocks) == 0 || blocks[len(blocks)-1].Num < toBlock {
			// Indicate the last downloaded block if there are not events on it
			log.Debugf("sending block %d to the driver (without events)", toBlock)
			downloadedCh <- EVMBlock{
				EVMBlockHeader: d.GetBlockHeader(ctx, toBlock),
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
}

func NewEVMDownloaderImplementation(
	ethClient EthClienter,
	blockFinality *big.Int,
	waitForNewBlocksPeriod time.Duration,
	appender LogAppenderMap,
	topicsToQuery []common.Hash,
	adressessToQuery []common.Address,
	rh *RetryHandler,
) *EVMDownloaderImplementation {
	return &EVMDownloaderImplementation{
		ethClient:              ethClient,
		blockFinality:          blockFinality,
		waitForNewBlocksPeriod: waitForNewBlocksPeriod,
		appender:               appender,
		topicsToQuery:          topicsToQuery,
		adressessToQuery:       adressessToQuery,
		rh:                     rh,
	}
}

func (d *EVMDownloaderImplementation) WaitForNewBlocks(ctx context.Context, lastBlockSeen uint64) (newLastBlock uint64) {
	attempts := 0
	ticker := time.NewTicker(d.waitForNewBlocksPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info("context cancelled")
			return lastBlockSeen
		case <-ticker.C:
			header, err := d.ethClient.HeaderByNumber(ctx, d.blockFinality)
			if err != nil {
				attempts++
				log.Error("error getting last block num from eth client: ", err)
				d.rh.Handle("waitForNewBlocks", attempts)
				continue
			}
			if header.Number.Uint64() > lastBlockSeen {
				return header.Number.Uint64()
			}
		}
	}
}

func (d *EVMDownloaderImplementation) GetEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) []EVMBlock {
	blocks := []EVMBlock{}
	logs := d.GetLogs(ctx, fromBlock, toBlock)
	for _, l := range logs {
		if len(blocks) == 0 || blocks[len(blocks)-1].Num < l.BlockNumber {
			b := d.GetBlockHeader(ctx, l.BlockNumber)
			if b.Hash != l.BlockHash {
				log.Infof(
					"there has been a block hash change between the event query and the block query for block %d: %s vs %s. Retrtying.",
					l.BlockNumber, b.Hash, l.BlockHash)
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
				log.Error("error trying to append log: ", err)
				d.rh.Handle("getLogs", attempts)
				continue
			}
			break
		}
	}

	return blocks
}

func (d *EVMDownloaderImplementation) GetLogs(ctx context.Context, fromBlock, toBlock uint64) []types.Log {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		Addresses: d.adressessToQuery,
		ToBlock:   new(big.Int).SetUint64(toBlock),
	}
	attempts := 0
	var (
		unfilteredLogs []types.Log
		err            error
	)
	for {
		unfilteredLogs, err = d.ethClient.FilterLogs(ctx, query)
		if err != nil {
			attempts++
			log.Error("error calling FilterLogs to eth client: ", err)
			d.rh.Handle("getLogs", attempts)
			continue
		}
		break
	}
	logs := []types.Log{}
	for _, l := range unfilteredLogs {
		found := false
		for _, topic := range d.topicsToQuery {
			if l.Topics[0] == topic {
				logs = append(logs, l)
				found = true
				break
			}
		}
		if !found {
			log.Debugf("ignoring log %+v because it's not under the list of topics to query", l)
		}
	}
	return logs
}

func (d *EVMDownloaderImplementation) GetBlockHeader(ctx context.Context, blockNum uint64) EVMBlockHeader {
	attempts := 0
	for {
		header, err := d.ethClient.HeaderByNumber(ctx, big.NewInt(int64(blockNum)))
		if err != nil {
			attempts++
			log.Errorf("error getting block header for block %d, err: %v", blockNum, err)
			d.rh.Handle("getBlockHeader", attempts)
			continue
		}
		return EVMBlockHeader{
			Num:        header.Number.Uint64(),
			Hash:       header.Hash(),
			ParentHash: header.ParentHash,
			Timestamp:  header.Time,
		}
	}
}
