package bridgesync

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridge"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
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

func buildAppender(client EthClienter, bridge common.Address) (sync.LogAppenderMap, error) {
	bridgeContractV1, err := polygonzkevmbridge.NewPolygonzkevmbridge(bridge, client)
	if err != nil {
		return nil, err
	}
	bridgeContractV2, err := polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridge, client)
	if err != nil {
		return nil, err
	}
	appender := make(sync.LogAppenderMap)

	appender[bridgeEventSignature] = func(b *sync.EVMBlock, l types.Log) error {
		bridge, err := bridgeContractV2.ParseBridgeEvent(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using d.bridgeContractV2.ParseBridgeEvent: %v",
				l, err,
			)
		}
		b.Events = append(b.Events, Event{Bridge: &Bridge{
			LeafType:           bridge.LeafType,
			OriginNetwork:      bridge.OriginNetwork,
			OriginAddress:      bridge.OriginAddress,
			DestinationNetwork: bridge.DestinationNetwork,
			DestinationAddress: bridge.DestinationAddress,
			Amount:             bridge.Amount,
			Metadata:           bridge.Metadata,
			DepositCount:       bridge.DepositCount,
		}})
		return nil
	}

	appender[claimEventSignature] = func(b *sync.EVMBlock, l types.Log) error {
		claim, err := bridgeContractV2.ParseClaimEvent(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using d.bridgeContractV2.ParseClaimEvent: %v",
				l, err,
			)
		}
		b.Events = append(b.Events, Event{Claim: &Claim{
			GlobalIndex:        claim.GlobalIndex,
			OriginNetwork:      claim.OriginNetwork,
			OriginAddress:      claim.OriginAddress,
			DestinationAddress: claim.DestinationAddress,
			Amount:             claim.Amount,
		}})
		return nil
	}

	appender[claimEventSignaturePreEtrog] = func(b *sync.EVMBlock, l types.Log) error {
		claim, err := bridgeContractV1.ParseClaimEvent(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using d.bridgeContractV1.ParseClaimEvent: %v",
				l, err,
			)
		}
		b.Events = append(b.Events, Event{Claim: &Claim{
			GlobalIndex:        big.NewInt(int64(claim.Index)),
			OriginNetwork:      claim.OriginNetwork,
			OriginAddress:      claim.OriginAddress,
			DestinationAddress: claim.DestinationAddress,
			Amount:             claim.Amount,
		}})
		return nil
	}

	return appender, nil
}

type fixedClient struct {
	*ethclient.Client
}

func (c *fixedClient) BlockByNumber(context.Context, rpc.BlockNumber) (*types.Block, error) {
	return nil, nil
}

func setClaimCalldata(ctx context.Context, client *ethclient.Client, txHash common.Hash, claim *Claim) error {
	tracer := tracers.NewAPI(client)
	tracerType := "callTracer"
	trace, err := tracer.TraceTransaction(ctx, txHash, &tracers.TraceConfig{Tracer: &tracerType})
	if err != nil {
		return err
	}
	log.Debug(trace)
	return nil
}

func decodeClaimCall(data []byte) {}
