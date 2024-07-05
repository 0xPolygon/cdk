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

type downloader struct {
	bridgeAddr         common.Address
	bridgeContractV1   *polygonzkevmbridge.Polygonzkevmbridge
	bridgeContractV2   *polygonzkevmbridgev2.Polygonzkevmbridgev2
	ethClient          EthClienter
	syncBlockChunkSize uint64
}

func newDownloader(
	bridgeAddr common.Address,
	syncBlockChunkSize uint64,
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
		bridgeAddr:         bridgeAddr,
		bridgeContractV1:   bridgeContractV1,
		bridgeContractV2:   bridgeContractV2,
		ethClient:          ethClient,
		syncBlockChunkSize: syncBlockChunkSize,
	}, nil
}

func (d *downloader) download(ctx context.Context, fromBlock uint64, downloadedCh chan block) {
	lastBlock := d.waitForNewBlocks(ctx, 0)
	for {
		select {
		case <-ctx.Done():
			close(downloadedCh)
			return
		default:
		}
		toBlock := fromBlock + d.syncBlockChunkSize
		if toBlock > lastBlock {
			toBlock = lastBlock
		}
		if fromBlock == toBlock {
			lastBlock = d.waitForNewBlocks(ctx, 0)
			continue
		}
		blocks := d.getEventsByBlockRange(ctx, fromBlock, toBlock)
		for _, b := range blocks {
			downloadedCh <- b
		}
		fromBlock = toBlock
	}
}

func (d *downloader) waitForNewBlocks(ctx context.Context, lastBlockSeen uint64) (newLastBlock uint64) {
	for {
		lastBlock, err := d.ethClient.BlockNumber(ctx)
		if err != nil {
			log.Error("error geting last block num from eth client: ", err)
			time.Sleep(retryAfterErrorPeriod)
			continue
		}
		if lastBlock > lastBlockSeen {
			return lastBlockSeen
		}
		time.Sleep(waitForNewBlocksPeriod)
	}
}

func (d *downloader) getEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) []block {
	blocks := []block{}
	logs := d.getLogs(ctx, fromBlock, toBlock)
	for _, l := range logs {
		if blocks[len(blocks)-1].Num < l.BlockNumber {
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
	for {
		logs, err := d.ethClient.FilterLogs(ctx, query)
		if err != nil {
			log.Error("error calling FilterLogs to eth client: ", err)
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
		log.Fatalf(
			"error parsing log %+v using d.bridgeContractV2.ParseClaimEvent: %v",
			l, err,
		)
		b.Events.Claims = append(b.Events.Claims, Claim{
			GlobalIndex:        claim.GlobalIndex,
			OriginNetwork:      claim.OriginNetwork,
			OriginAddress:      claim.OriginAddress,
			DestinationAddress: claim.DestinationAddress,
			Amount:             claim.Amount,
		})
	case claimEventSignaturePreEtrog:
		claim, err := d.bridgeContractV1.ParseClaimEvent(l)
		log.Fatalf(
			"error parsing log %+v using d.bridgeContractV1.ParseClaimEvent: %v",
			l, err,
		)
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
