package contracts

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractBase[T any] struct {
	contractBind *T
	address      common.Address
	contractName NameType
	version      VersionType
}

type contractConstructorFunc[T any] func(address common.Address, backend bind.ContractBackend) (*T, error)

func NewContractBase[T any](constructor contractConstructorFunc[T], address common.Address, backend bind.ContractBackend,
	name NameType, version VersionType) (*ContractBase[T], error) {
	contractBind, err := constructor(address, backend)
	if err != nil {
		return nil, err
	}

	return &ContractBase[T]{
		contractBind: contractBind,
		address:      address,
		contractName: name,
		version:      version,
	}, nil
}

func (e *ContractBase[T]) GetContract() *T {
	return e.contractBind
}

func (e *ContractBase[T]) GetAddress() common.Address {
	return e.address
}

func (e *ContractBase[T]) GetName() string {
	return string(e.contractName)
}

func (e *ContractBase[T]) GetVersion() string {
	return string(e.version)
}

func (e *ContractBase[T]) String() string {
	return e.GetVersion() + "/" + e.GetName() + "@" + e.address.String()
}
