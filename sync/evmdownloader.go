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

type evmDownloaderInterface interface {
	waitForNewBlocks(ctx context.Context, lastBlockSeen uint64) (newLastBlock uint64)
	getEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) []EVMBlock
	getLogs(ctx context.Context, fromBlock, toBlock uint64) []types.Log
	getBlockHeader(ctx context.Context, blockNum uint64) EVMBlockHeader
}

type LogAppenderMap map[common.Hash]func(b *EVMBlock, l types.Log) error

type EVMDownloader struct {
	syncBlockChunkSize uint64
	evmDownloaderInterface
}

func NewEVMDownloader(
	ethClient EthClienter,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
	waitForNewBlocksPeriod time.Duration,
	appender LogAppenderMap,
	adressessToQuery []common.Address,
) (*EVMDownloader, error) {
	finality, err := blockFinalityType.ToBlockNum()
	if err != nil {
		return nil, err
	}
	topicsToQuery := [][]common.Hash{}
	for topic := range appender {
		topicsToQuery = append(topicsToQuery, []common.Hash{topic})
	}
	return &EVMDownloader{
		syncBlockChunkSize: syncBlockChunkSize,
		evmDownloaderInterface: &downloaderImplementation{
			ethClient:              ethClient,
			blockFinality:          finality,
			waitForNewBlocksPeriod: waitForNewBlocksPeriod,
			appender:               appender,
			topicsToQuery:          topicsToQuery,
			adressessToQuery:       adressessToQuery,
		},
	}, nil
}

func (d *EVMDownloader) download(ctx context.Context, fromBlock uint64, downloadedCh chan EVMBlock) {
	lastBlock := d.waitForNewBlocks(ctx, 0)
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
			lastBlock = d.waitForNewBlocks(ctx, toBlock)
			continue
		}
		log.Debugf("getting events from blocks %d to  %d", fromBlock, toBlock)
		blocks := d.getEventsByBlockRange(ctx, fromBlock, toBlock)
		for _, b := range blocks {
			log.Debugf("sending block %d to the driver (with events)", b.Num)
			downloadedCh <- b
		}
		if len(blocks) == 0 || blocks[len(blocks)-1].Num < toBlock {
			// Indicate the last downloaded block if there are not events on it
			log.Debugf("sending block %d to the driver (without evvents)", toBlock)
			downloadedCh <- EVMBlock{
				EVMBlockHeader: d.getBlockHeader(ctx, toBlock),
			}
		}
		fromBlock = toBlock + 1
	}
}

type downloaderImplementation struct {
	ethClient              EthClienter
	blockFinality          *big.Int
	waitForNewBlocksPeriod time.Duration
	appender               LogAppenderMap
	topicsToQuery          [][]common.Hash
	adressessToQuery       []common.Address
}

func (d *downloaderImplementation) waitForNewBlocks(ctx context.Context, lastBlockSeen uint64) (newLastBlock uint64) {
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
				RetryHandler("waitForNewBlocks", attempts)
				continue
			}
			if header.Number.Uint64() > lastBlockSeen {
				return header.Number.Uint64()
			}
		}
	}
}

func (d *downloaderImplementation) getEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) []EVMBlock {
	blocks := []EVMBlock{}
	logs := d.getLogs(ctx, fromBlock, toBlock)
	for _, l := range logs {
		if len(blocks) == 0 || blocks[len(blocks)-1].Num < l.BlockNumber {
			b := d.getBlockHeader(ctx, l.BlockNumber)
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
				RetryHandler("getLogs", attempts)
				continue
			}
			break
		}
	}

	return blocks
}

func (d *downloaderImplementation) getLogs(ctx context.Context, fromBlock, toBlock uint64) []types.Log {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		Addresses: d.adressessToQuery,
		Topics:    d.topicsToQuery,
		ToBlock:   new(big.Int).SetUint64(toBlock),
	}
	attempts := 0
	for {
		logs, err := d.ethClient.FilterLogs(ctx, query)
		if err != nil {
			attempts++
			log.Error("error calling FilterLogs to eth client: ", err)
			RetryHandler("getLogs", attempts)
			continue
		}
		return logs
	}
}

func (d *downloaderImplementation) getBlockHeader(ctx context.Context, blockNum uint64) EVMBlockHeader {
	attempts := 0
	for {
		header, err := d.ethClient.HeaderByNumber(ctx, big.NewInt(int64(blockNum)))
		if err != nil {
			attempts++
			log.Errorf("error getting block header for block %d, err: %v", blockNum, err)
			RetryHandler("getBlockHeader", attempts)
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
