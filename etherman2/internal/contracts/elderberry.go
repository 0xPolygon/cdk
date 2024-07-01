package internal_contracts

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractRollupElderberry struct {
	*ContractBase[polygonvalidiumetrog.Polygonvalidiumetrog]
}

func NewContractRollupElderberry(address common.Address, backend bind.ContractBackend) (*ContractRollupElderberry, error) {
	//ZkEVM, err := polygonvalidiumetrog.NewPolygonvalidiumetrog(address, backend)
	base, err := NewContractBase(polygonvalidiumetrog.NewPolygonvalidiumetrog, address, backend)
	if err != nil {
		return nil, err
	}
	return &ContractRollupElderberry{
		ContractBase: base,
	}, nil
}

func (e *ContractRollupElderberry) TrustedSequencer() (common.Address, error) {
	return e.GetContract().TrustedSequencer(&bind.CallOpts{Pending: false})
}
