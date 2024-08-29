package bridgesync

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridge"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridgev2"
	rpcTypes "github.com/0xPolygon/cdk-rpc/types"
	"github.com/0xPolygon/cdk/sync"
	"github.com/0xPolygon/cdk/tree"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang-collections/collections/stack"
)

var (
	bridgeEventSignature        = crypto.Keccak256Hash([]byte("BridgeEvent(uint8,uint32,address,uint32,address,uint256,bytes,uint32)"))
	claimEventSignature         = crypto.Keccak256Hash([]byte("ClaimEvent(uint256,uint32,address,address,uint256)"))
	claimEventSignaturePreEtrog = crypto.Keccak256Hash([]byte("ClaimEvent(uint32,uint32,address,address,uint256)"))
	methodIDClaimAsset          = common.Hex2Bytes("ccaa2d11")
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

type call struct {
	To    common.Address   `json:"to"`
	Value *rpcTypes.ArgBig `json:"value"`
	// Err   *string           `json:"error"`
	Input rpcTypes.ArgBytes `json:"input"`
	Calls []call            `json:"calls"`
}

type tracerCfg struct {
	Tracer string `json:"tracer"`
}

func setClaimCalldata(client *ethclient.Client, bridgeAddr common.Address, txHash common.Hash, claim *Claim) error {
	c := &call{}
	err := client.Client().Call(c, "debug_traceTransaction", txHash, tracerCfg{Tracer: "callTracer"})
	if err != nil {
		return err
	}

	// find the claim linked to the event using DFS
	// TODO: take into account potential reverts that may be found on the path,
	// and other edge cases
	callStack := stack.New()
	callStack.Push(*c)
	for {
		if callStack.Len() == 0 {
			break
		}
		currentCall := callStack.Pop().(call)
		if currentCall.To == bridgeAddr {
			found, err := setClaimIfFoundOnInput(
				currentCall.Input,
				claim,
			)
			if err != nil {
				return err
			}
			if found {
				return nil
			}
		}
		for _, c := range currentCall.Calls {
			callStack.Push(c)
		}
	}
	return ErrNotFound
}

func setClaimIfFoundOnInput(input []byte, claim *Claim) (bool, error) {
	smcAbi, err := abi.JSON(strings.NewReader(polygonzkevmbridgev2.Polygonzkevmbridgev2ABI))
	if err != nil {
		return false, err
	}
	methodId := input[:4]

	// Ignore other methods
	if !bytes.Equal(methodId, methodIDClaimAsset) {
		return false, nil
	}

	// Recover Method from signature and ABI
	method, err := smcAbi.MethodById(methodId)
	if err != nil {
		return false, err
	}

	/* Unpack method inputs
	claimAsset(
		0: smtProofLocalExitRoot,
		1: smtProofRollupExitRoot,
		2: globalIndex,
		3: mainnetExitRoot,
		4: rollupExitRoot,
		5: originNetwork,
		6: originTokenAddress,
		7: destinationNetwork,
		8: destinationAddress,
		9: amount,
		10: metadata,
	)
	*/
	data, err := method.Inputs.Unpack(input[4:])
	if err != nil {
		return false, err
	}

	// TODO: support both claim asset & message, check if previous versions need special treatment
	// TODO: ignore claim messages that don't have value
	actualGlobalIndex := data[2].(*big.Int)
	if actualGlobalIndex.Cmp(claim.GlobalIndex) != 0 {
		// not the claim we're looking for
		return false, nil
	} else {
		proofLER := [tree.DefaultHeight]common.Hash{}
		proofLERBytes := data[0].([32][32]byte)
		proofRER := [tree.DefaultHeight]common.Hash{}
		proofRERBytes := data[1].([32][32]byte)
		for i := 0; i < int(tree.DefaultHeight); i++ {
			proofLER[i] = proofLERBytes[i]
			proofRER[i] = proofRERBytes[i]
		}
		// TODO: add ALL the data, hard to know what we're gonna need in the future
		claim.ProofLocalExitRoot = proofLER
		claim.ProofRollupExitRoot = proofRER
		claim.MainnetExitRoot = data[3].([32]byte)
		claim.RollupExitRoot = data[4].([32]byte)
		claim.Amount = data[9].(*big.Int)
		return true, nil
	}
}
