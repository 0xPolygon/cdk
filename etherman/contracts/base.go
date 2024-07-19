package contracts

import (
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractBase[T any] struct {
	contract     *T
	address      common.Address
	contractName NameType
	version      VersionType
}

type contractConstructorFunc[T any] func(address common.Address, backend bind.ContractBackend) (*T, error)

func NewContractBase[T any](constructor contractConstructorFunc[T], address common.Address, backend bind.ContractBackend,
	name NameType, version VersionType) (*ContractBase[T], error) {
	contractBind, err := constructor(address, backend)
	if err != nil {
		log.Errorf("failed to bind contract %s at address %s. Err:%w", name, address.String(), err)
		return nil, err
	}

	return &ContractBase[T]{
		contract:     contractBind,
		address:      address,
		contractName: name,
		version:      version,
	}, nil
}

func (e *ContractBase[T]) Contract() *T {
	return e.contract
}

func (e *ContractBase[T]) Address() common.Address {
	return e.address
}

func (e *ContractBase[T]) Name() string {
	return string(e.contractName)
}

func (e *ContractBase[T]) Version() string {
	return string(e.version)
}

func (e *ContractBase[T]) String() string {
	return e.Version() + "/" + e.Name() + "@" + e.Address().String()
}
