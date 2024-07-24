package l1infotreesync

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	waitForNewBlocksPeriod    = time.Millisecond * 100
	updateL1InfoTreeSignature = crypto.Keccak256Hash([]byte("UpdateL1InfoTree(bytes32,bytes32)"))
)

type EthClienter interface {
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainReader
	bind.ContractBackend
}

type downloaderInterface interface {
	waitForNewBlocks(ctx context.Context, lastBlockSeen uint64) (newLastBlock uint64)
	getEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) []block
	getLogs(ctx context.Context, fromBlock, toBlock uint64) []types.Log
	appendLog(b *block, l types.Log)
	getBlockHeader(ctx context.Context, blockNum uint64) blockHeader
}

type L1InfoTreeUpdate struct {
	MainnetExitRoot common.Hash
	RollupExitRoot  common.Hash
}

type block struct {
	blockHeader
	Events []L1InfoTreeUpdate
}

type blockHeader struct {
	Num        uint64
	Hash       common.Hash
	ParentHash common.Hash
	Timestamp  uint64
}

type downloader struct {
	syncBlockChunkSize uint64
	downloaderInterface
}

func newDownloader(
	GERAddr common.Address,
	ethClient EthClienter,
	syncBlockChunkSize uint64,
	blockFinalityType etherman.BlockNumberFinality,
) (*downloader, error) {
	GERContract, err := polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2(GERAddr, ethClient)
	if err != nil {
		return nil, err
	}
	finality, err := blockFinalityType.ToBlockNum()
	if err != nil {
		return nil, err
	}
	return &downloader{
		syncBlockChunkSize: syncBlockChunkSize,
		downloaderInterface: &downloaderImplementation{
			GERAddr:       GERAddr,
			GERContract:   GERContract,
			ethClient:     ethClient,
			blockFinality: finality,
		},
	}, nil
}

func (d *downloader) download(ctx context.Context, fromBlock uint64, downloadedCh chan block) {
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
			downloadedCh <- block{
				blockHeader: d.getBlockHeader(ctx, toBlock),
			}
		}
		fromBlock = toBlock + 1
	}
}

type downloaderImplementation struct {
	GERAddr       common.Address
	GERContract   *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2
	ethClient     EthClienter
	blockFinality *big.Int
}

func (d *downloaderImplementation) waitForNewBlocks(ctx context.Context, lastBlockSeen uint64) (newLastBlock uint64) {
	attempts := 0
	for {
		header, err := d.ethClient.HeaderByNumber(ctx, d.blockFinality)
		if err != nil {
			attempts++
			log.Error("error geting last block num from eth client: ", err)
			retryHandler("waitForNewBlocks", attempts)
			continue
		}
		if header.Number.Uint64() > lastBlockSeen {
			return header.Number.Uint64()
		}
		time.Sleep(waitForNewBlocksPeriod)
	}
}

func (d *downloaderImplementation) getEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) []block {
	blocks := []block{}
	logs := d.getLogs(ctx, fromBlock, toBlock)
	for _, l := range logs {
		if len(blocks) == 0 || blocks[len(blocks)-1].Num < l.BlockNumber {
			b := d.getBlockHeader(ctx, l.BlockNumber)
			blocks = append(blocks, block{
				blockHeader: blockHeader{
					Num:        l.BlockNumber,
					Hash:       l.BlockHash,
					Timestamp:  b.Timestamp,
					ParentHash: b.ParentHash,
				},
				Events: []L1InfoTreeUpdate{},
			})
		}
		d.appendLog(&blocks[len(blocks)-1], l)
	}

	return blocks
}

func (d *downloaderImplementation) getLogs(ctx context.Context, fromBlock, toBlock uint64) []types.Log {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		Addresses: []common.Address{d.GERAddr},
		Topics: [][]common.Hash{
			{updateL1InfoTreeSignature},
		},
		ToBlock: new(big.Int).SetUint64(toBlock),
	}
	attempts := 0
	for {
		logs, err := d.ethClient.FilterLogs(ctx, query)
		if err != nil {
			attempts++
			log.Error("error calling FilterLogs to eth client: ", err)
			retryHandler("getLogs", attempts)
			continue
		}
		return logs
	}
}

func (d *downloaderImplementation) appendLog(b *block, l types.Log) {
	switch l.Topics[0] {
	case updateL1InfoTreeSignature:
		l1InfoTreeUpdate, err := d.GERContract.ParseUpdateL1InfoTree(l)
		if err != nil {
			log.Fatalf(
				"error parsing log %+v using d.GERContract.ParseUpdateL1InfoTree: %v",
				l, err,
			)
		}
		b.Events = append(b.Events, L1InfoTreeUpdate{
			MainnetExitRoot: l1InfoTreeUpdate.MainnetExitRoot,
			RollupExitRoot:  l1InfoTreeUpdate.RollupExitRoot,
		})
	default:
		log.Fatalf("unexpected log %+v", l)
	}
}

func (d *downloaderImplementation) getBlockHeader(ctx context.Context, blockNum uint64) blockHeader {
	attempts := 0
	for {
		header, err := d.ethClient.HeaderByNumber(ctx, big.NewInt(int64(blockNum)))
		if err != nil {
			attempts++
			log.Errorf("error getting block header for block %d, err: %v", blockNum, err)
			retryHandler("getBlockHeader", attempts)
			continue
		}
		return blockHeader{
			Num:        header.Number.Uint64(),
			Hash:       header.Hash(),
			ParentHash: header.ParentHash,
			Timestamp:  header.Time,
		}
	}
}
