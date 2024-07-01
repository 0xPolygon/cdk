package localbridgesync

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridge"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridgev2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
}

type downloader struct {
	downloadedCh       chan block
	bridgeAddr         common.Address
	bridgeContractV2   *polygonzkevmbridgev2.Polygonzkevmbridgev2
	bridgeContractV1   *polygonzkevmbridge.Polygonzkevmbridge
	ethClient          EthClienter
	blockToBlocks      map[uint64]struct{ from, to uint64 }
	syncBlockChunkSize uint64
}

func newDownloader() (*downloader, error) {
	return nil, errors.New("not implemented")
}

func (d *downloader) download(ctx context.Context, fromBlock uint64, downloadedCh chan block) {
	d.downloadedCh = downloadedCh
	for {
		select {
		case <-ctx.Done():
			close(downloadedCh)
			return
		default:
		}
		lastBlock, err := d.ethClient.BlockNumber(ctx)
		if err != nil {
			// TODO: handle error
			return
		}
		toBlock := fromBlock + d.syncBlockChunkSize
		if toBlock > lastBlock {
			toBlock = lastBlock
		}
		if fromBlock == toBlock {
			time.Sleep(time.Millisecond * 100) // sleep 100ms for the L2 to produce more blocks
			continue
		}
		blocks, err := d.getEventsByBlockRange(ctx, fromBlock, toBlock)
		if err != nil {
			// TODO: handle error
			return
		}
		for _, b := range blocks {
			d.downloadedCh <- b
		}
		fromBlock = toBlock
	}
}

func (d *downloader) getEventsByBlockRange(ctx context.Context, fromBlock, toBlock uint64) ([]block, error) {
	blocks := []block{}
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

	logs, err := d.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}
	for _, l := range logs {
		lastBlock := blocks[len(blocks)-1]
		if lastBlock.Num < l.BlockNumber {
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
		switch l.Topics[0] {
		case bridgeEventSignature:
			bridge, err := d.bridgeContractV2.ParseBridgeEvent(l)
			if err != nil {
				return nil, err
			}
			blocks[len(blocks)-1].Events.Bridges = append(blocks[len(blocks)-1].Events.Bridges, Bridge{
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
				return nil, err
			}
			blocks[len(blocks)-1].Events.Claims = append(blocks[len(blocks)-1].Events.Claims, Claim{
				GlobalIndex:        claim.GlobalIndex,
				OriginNetwork:      claim.OriginNetwork,
				OriginAddress:      claim.OriginAddress,
				DestinationAddress: claim.DestinationAddress,
				Amount:             claim.Amount,
			})
		case claimEventSignaturePreEtrog:
			claim, err := d.bridgeContractV1.ParseClaimEvent(l)
			if err != nil {
				return nil, err
			}
			blocks[len(blocks)-1].Events.Claims = append(blocks[len(blocks)-1].Events.Claims, Claim{
				// WARNING: is it safe to convert Index --> GlobalIndex???
				// according to Jesus, yes!
				GlobalIndex:        big.NewInt(int64(claim.Index)),
				OriginNetwork:      claim.OriginNetwork,
				OriginAddress:      claim.OriginAddress,
				DestinationAddress: claim.DestinationAddress,
				Amount:             claim.Amount,
			})
		default:
			return nil, errors.New("unexpected topic")
		}
	}

	return blocks, nil
}

func (d *downloader) getBlockHeader(ctx context.Context, blockNum uint64) (blockHeader, error) {
	bn := big.NewInt(int64(blockNum))
	block, err := d.ethClient.BlockByNumber(ctx, bn)
	if err != nil {
		return blockHeader{}, err
	}
	return blockHeader{
		Num:  block.NumberU64(),
		Hash: block.Hash(),
	}, nil
}
