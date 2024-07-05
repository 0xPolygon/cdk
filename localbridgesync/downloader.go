package localbridgesync

import (
	"context"
	"errors"
	"math/big"

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

type downloader struct {
	downloadedCh     chan batch
	bridgeAddr       common.Address
	bridgeContractV2 *polygonzkevmbridgev2.Polygonzkevmbridgev2
	bridgeContractV1 *polygonzkevmbridge.Polygonzkevmbridge
	ethClient        ethereum.LogFilterer
	batchToBlocks    map[uint64]struct{ from, to uint64 }
}

func newDownloader() (*downloader, error) {
	return nil, errors.New("not implemented")
}

func (d *downloader) download(ctx context.Context, fromBatchNum uint64) {
	// how to get first blokc associated to batch num??? --> zkevm_getBatchByNumber
	/*
			"result": {
		            "name": "Batch",
		            "value": {
		              "number": "0x1",
		              "coinbase": "0x0000000000000000000000000000000000000001",
		              "stateRoot": "0x0000000000000000000000000000000000000000000000000000000000000001",
		              "globalExitRoot": "0x0000000000000000000000000000000000000000000000000000000000000002",
		              "mainnetExitRoot": "0x0000000000000000000000000000000000000000000000000000000000000003",
		              "rollupExitRoot": "0x0000000000000000000000000000000000000000000000000000000000000004",
		              "localExitRoot": "0x0000000000000000000000000000000000000000000000000000000000000005",
		              "accInputHash": "0x0000000000000000000000000000000000000000000000000000000000000006",
		              "timestamp": "0x642af31f",
		              "sendSequencesTxHash": "0x0000000000000000000000000000000000000000000000000000000000000007",
		              "verifyBatchTxHash": "0x0000000000000000000000000000000000000000000000000000000000000008",
		              "transactions": [
		                "0x0000000000000000000000000000000000000000000000000000000000000009",
		                "0x0000000000000000000000000000000000000000000000000000000000000010",
		                "0x0000000000000000000000000000000000000000000000000000000000000011"
		              ]
		            }
		          }

				  flacky flacky double checky
	*/
	for {
		// zkevm_batchNumberByBlockNumber
	}
}

type block struct {
	num     uint64
	hash    common.Hash
	bridges []Bridge
	claims  []Claim
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
		if lastBlock.num < l.BlockNumber {
			blocks = append(blocks, block{
				num:     l.BlockNumber,
				hash:    l.BlockHash,
				claims:  []Claim{},
				bridges: []Bridge{},
			})
		}
		switch l.Topics[0] {
		case bridgeEventSignature:
			bridge, err := d.bridgeContractV2.ParseBridgeEvent(l)
			if err != nil {
				return nil, err
			}
			blocks[len(blocks)-1].bridges = append(blocks[len(blocks)-1].bridges, Bridge{
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
			blocks[len(blocks)-1].claims = append(blocks[len(blocks)-1].claims, Claim{
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
			blocks[len(blocks)-1].claims = append(blocks[len(blocks)-1].claims, Claim{
				// WARNING: is it safe to convert Index --> GlobalIndex???
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
