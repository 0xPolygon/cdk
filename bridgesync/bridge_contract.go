package bridgesync

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonzkevmbridgev2"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type BridgeContract struct {
	bridgeAddr common.Address
	Contract   *polygonzkevmbridgev2.Polygonzkevmbridgev2
}

func NewBridgeContract(bridgeAddr common.Address, backend bind.ContractBackend) (*BridgeContract, error) {
	contract, err := polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridgeAddr, backend)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate bridge contract at addr %s.  Err:%w", bridgeAddr.String(), err)
	}

	return &BridgeContract{
		bridgeAddr: bridgeAddr,
		Contract:   contract,
	}, nil
}

// Returns LastUpdatedDepositCount for a specific blockNumber
func (b *BridgeContract) LastUpdatedDepositCount(ctx context.Context, blockNumber uint64) (uint32, error) {
	opts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).SetUint64(blockNumber),
	}
	return b.Contract.LastUpdatedDepositCount(opts)
}
