package localbridgesync

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridge"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	waitForNewBlocksPeriod = time.Millisecond * 100
)

var (
	bridgeEventSignature        = crypto.Keccak256Hash([]byte("BridgeEvent(uint8,uint32,address,uint32,address,uint256,bytes,uint32)"))
	claimEventSignature         = crypto.Keccak256Hash([]byte("ClaimEvent(uint256,uint32,address,address,uint256)"))
	claimEventSignaturePreEtrog = crypto.Keccak256Hash([]byte("ClaimEvent(uint32,uint32,address,address,uint256)"))
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

type downloader struct {
	bridgeAddr       common.Address
	bridgeContractV1 *polygonzkevmbridge.Polygonzkevmbridge
	bridgeContractV2 *polygonzkevmbridgev2.Polygonzkevmbridgev2
	ethClient        EthClienter
}

func newDownloader(
	bridgeAddr common.Address,
	ethClient EthClienter,
) (*downloader, error) {
	bridgeContractV1, err := polygonzkevmbridge.NewPolygonzkevmbridge(bridgeAddr, ethClient)
	if err != nil {
		return nil, err
	}
	bridgeContractV2, err := polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridgeAddr, ethClient)
	if err != nil {
		return nil, err
	}
	return &downloader{
		bridgeAddr:       bridgeAddr,
		bridgeContractV1: bridgeContractV1,
		bridgeContractV2: bridgeContractV2,
		ethClient:        ethClient,
	}, nil
}

func download(ctx context.Context, d downloaderInterface, fromBlock, syncBlockChunkSize uint64, downloadedCh chan block) {
	lastBlock := d.waitForNewBlocks(ctx, 0)
	for {
		log.Debug("select")
		select {
		case <-ctx.Done():
			log.Debug("closing channel")
			close(downloadedCh)
			return
		default:
		}
		log.Debug("default")
		toBlock := fromBlock + syncBlockChunkSize
		if toBlock > lastBlock {
			toBlock = lastBlock
		}
		log.Debug("1")
		if fromBlock >= toBlock {
			log.Debug("waitForNewBlocks")
			lastBlock = d.waitForNewBlocks(ctx, toBlock)
			log.Debug("out waitForNewBlocks")
			continue
		}
		log.Debug("2", fromBlock, toBlock)
		blocks := d.getEventsByBlockRange(ctx, fromBlock, toBlock)
		for _, b := range blocks {
			log.Debug("sending block with events")
			downloadedCh <- b
		}
		log.Debug("3")
		if len(blocks) == 0 || blocks[len(blocks)-1].Num < toBlock {
			// Indicate the last downloaded block if there are not events on it
			log.Debug("sending block without events")
			downloadedCh <- block{
				blockHeader: d.getBlockHeader(ctx, toBlock),
			}
		}
		log.Debug("4")
		fromBlock = toBlock + 1
	}
}

func (d *downloader) waitForNewBlocks(ctx context.Context, lastBlockSeen uint64) (newLastBlock uint64) {
	attempts := 0
	for {
		lastBlock, err := d.ethClient.BlockNumber(ctx)
		if err != nil {
			attempts++
			log.Error("error geting last block num from eth client: ", err)
			retryHandler("waitForNewBlocks", attempts)
			continue
		}
		if lastBlock > lastBlockSeen {
			return lastBlock
		}
		time.Sleep(waitForNewBlocksPeriod)
	}
}

func (d *downloader) getEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) []block {
	blocks := []block{}
	logs := d.getLogs(ctx, fromBlock, toBlock)
	for _, l := range logs {
		if len(blocks) == 0 || blocks[len(blocks)-1].Num < l.BlockNumber {
			blocks = append(blocks, block{
				blockHeader: blockHeader{
					Num:  l.BlockNumber,
					Hash: l.BlockHash,
				},
				Events: bridgeEvents{
					Claims:  []Claim{},
					Bridges: []Bridge{},
				},
			})
		}
		d.appendLog(&blocks[len(blocks)-1], l)
	}

	return blocks
}

func (d *downloader) getLogs(ctx context.Context, fromBlock, toBlock uint64) []types.Log {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		Addresses: []common.Address{d.bridgeAddr},
		Topics: [][]common.Hash{
			{bridgeEventSignature},
			{claimEventSignature},
			{claimEventSignaturePreEtrog},
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

func (d *downloader) appendLog(b *block, l types.Log) {
	switch l.Topics[0] {
	case bridgeEventSignature:
		bridge, err := d.bridgeContractV2.ParseBridgeEvent(l)
		if err != nil {
			log.Fatalf(
				"error parsing log %+v using d.bridgeContractV2.ParseBridgeEvent: %v",
				l, err,
			)
		}
		b.Events.Bridges = append(b.Events.Bridges, Bridge{
			LeafType:           bridge.LeafType,
			OriginNetwork:      bridge.OriginNetwork,
			OriginAddress:      bridge.OriginAddress,
			DestinationNetwork: bridge.DestinationNetwork,
			DestinationAddress: bridge.DestinationAddress,
			Amount:             bridge.Amount,
			Metadata:           bridge.Metadata,
			DepositCount:       bridge.DepositCount,
		})
	case claimEventSignature:
		claim, err := d.bridgeContractV2.ParseClaimEvent(l)
		if err != nil {
			log.Fatalf(
				"error parsing log %+v using d.bridgeContractV2.ParseClaimEvent: %v",
				l, err,
			)
		}
		b.Events.Claims = append(b.Events.Claims, Claim{
			GlobalIndex:        claim.GlobalIndex,
			OriginNetwork:      claim.OriginNetwork,
			OriginAddress:      claim.OriginAddress,
			DestinationAddress: claim.DestinationAddress,
			Amount:             claim.Amount,
		})
	case claimEventSignaturePreEtrog:
		claim, err := d.bridgeContractV1.ParseClaimEvent(l)
		if err != nil {
			log.Fatalf(
				"error parsing log %+v using d.bridgeContractV1.ParseClaimEvent: %v",
				l, err,
			)
		}
		b.Events.Claims = append(b.Events.Claims, Claim{
			GlobalIndex:        big.NewInt(int64(claim.Index)),
			OriginNetwork:      claim.OriginNetwork,
			OriginAddress:      claim.OriginAddress,
			DestinationAddress: claim.DestinationAddress,
			Amount:             claim.Amount,
		})
	default:
		log.Fatalf("unexpected log %+v", l)
	}
}

func (d *downloader) getBlockHeader(ctx context.Context, blockNum uint64) blockHeader {
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
			Num:  header.Number.Uint64(),
			Hash: header.Hash(),
		}
	}
}
