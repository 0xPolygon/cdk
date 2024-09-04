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
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/golang-collections/collections/stack"
)

var (
	bridgeEventSignature        = crypto.Keccak256Hash([]byte("BridgeEvent(uint8,uint32,address,uint32,address,uint256,bytes,uint32)"))
	claimEventSignature         = crypto.Keccak256Hash([]byte("ClaimEvent(uint256,uint32,address,address,uint256)"))
	claimEventSignaturePreEtrog = crypto.Keccak256Hash([]byte("ClaimEvent(uint32,uint32,address,address,uint256)"))
	methodIDClaimAsset          = common.Hex2Bytes("ccaa2d11")
	methodIDClaimMessage        = common.Hex2Bytes("f5efcd79")
)

type EthClienter interface {
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainReader
	bind.ContractBackend
	Client() *rpc.Client
}

func buildAppender(client EthClienter, bridge common.Address, syncFullClaims bool) (sync.LogAppenderMap, error) {
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
		claimEvent, err := bridgeContractV2.ParseClaimEvent(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using d.bridgeContractV2.ParseClaimEvent: %v",
				l, err,
			)
		}
		claim := &Claim{
			GlobalIndex:        claimEvent.GlobalIndex,
			OriginNetwork:      claimEvent.OriginNetwork,
			OriginAddress:      claimEvent.OriginAddress,
			DestinationAddress: claimEvent.DestinationAddress,
			Amount:             claimEvent.Amount,
		}
		if syncFullClaims {
			setClaimCalldata(client, bridge, l.TxHash, claim)
		}
		b.Events = append(b.Events, Event{Claim: claim})
		return nil
	}

	appender[claimEventSignaturePreEtrog] = func(b *sync.EVMBlock, l types.Log) error {
		claimEvent, err := bridgeContractV1.ParseClaimEvent(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using d.bridgeContractV1.ParseClaimEvent: %v",
				l, err,
			)
		}
		claim := &Claim{
			GlobalIndex:        big.NewInt(int64(claimEvent.Index)),
			OriginNetwork:      claimEvent.OriginNetwork,
			OriginAddress:      claimEvent.OriginAddress,
			DestinationAddress: claimEvent.DestinationAddress,
			Amount:             claimEvent.Amount,
		}
		if syncFullClaims {
			setClaimCalldata(client, bridge, l.TxHash, claim)
		}
		b.Events = append(b.Events, Event{Claim: claim})
		return nil
	}

	return appender, nil
}

type call struct {
	To    common.Address    `json:"to"`
	Value *rpcTypes.ArgBig  `json:"value"`
	Err   *string           `json:"error"`
	Input rpcTypes.ArgBytes `json:"input"`
	Calls []call            `json:"calls"`
}

type tracerCfg struct {
	Tracer string `json:"tracer"`
}

func setClaimCalldata(client EthClienter, bridge common.Address, txHash common.Hash, claim *Claim) error {
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
		if currentCall.To == bridge {
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
	// Recover Method from signature and ABI
	method, err := smcAbi.MethodById(methodId)
	if err != nil {
		return false, err
	}
	data, err := method.Inputs.Unpack(input[4:])
	if err != nil {
		return false, err
	}
	// Ignore other methods
	if bytes.Equal(methodId, methodIDClaimAsset) || bytes.Equal(methodId, methodIDClaimMessage) {
		found, err := decodeClaimCallDataAndSetIfFound(data, claim)
		if err != nil {
			return false, err
		}
		if found {
			if bytes.Equal(methodId, methodIDClaimMessage) {
				claim.IsMessage = true
			}
			return true, nil
		}
		return false, nil
	} else {
		return false, nil
	}
	// TODO: support both claim asset & message, check if previous versions need special treatment
}

func decodeClaimCallDataAndSetIfFound(data []interface{}, claim *Claim) (bool, error) {
	/* Unpack method inputs. Note that both claimAsset and claimMessage have the same interface
	for the relevant parts
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
	claimMessage(
		0: smtProofLocalExitRoot,
		1: smtProofRollupExitRoot,
		2: globalIndex,
		3: mainnetExitRoot,
		4: rollupExitRoot,
		5: originNetwork,
		6: originAddress,
		7: destinationNetwork,
		8: destinationAddress,
		9: amount,
		10: metadata,
	)
	*/
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
		claim.ProofLocalExitRoot = proofLER
		claim.ProofRollupExitRoot = proofRER
		claim.MainnetExitRoot = data[3].([32]byte)
		claim.RollupExitRoot = data[4].([32]byte)
		claim.DestinationNetwork = data[7].(uint32)
		claim.Metadata = data[10].([]byte)
		return true, nil
	}
}
