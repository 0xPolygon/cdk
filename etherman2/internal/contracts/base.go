package internal_contracts

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractBase[T any] struct {
	contractBind *T
	address      common.Address
}

type contractConstructorFunc[T any] func(address common.Address, backend bind.ContractBackend) (*T, error)

func NewContractBase[T any](constructor contractConstructorFunc[T], address common.Address, backend bind.ContractBackend) (*ContractBase[T], error) {
	contractBind, err := constructor(address, backend)
	if err != nil {
		return nil, err
	}

	return &ContractBase[T]{
		contractBind: contractBind,
		address:      address,
	}, nil
}

func (e *ContractBase[T]) GetContract() *T {
	return e.contractBind
}

func (e *ContractBase[T]) GetAddress() common.Address {
	return e.address
}
