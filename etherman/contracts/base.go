package contracts

import (
	"reflect"

	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractBase struct {
	address      common.Address
	contractName NameType
	version      VersionType
}

type contractConstructorFunc[T any] func(address common.Address, backend bind.ContractBackend) (T, error)

func NewContractBase(address common.Address, backend bind.ContractBackend,
	name NameType, version VersionType) *ContractBase {
	return &ContractBase{
		address:      address,
		contractName: name,
		version:      version,
	}
}

func (e *ContractBase) Address() common.Address {
	return e.address
}

func (e *ContractBase) Name() string {
	return string(e.contractName)
}

func (e *ContractBase) Version() string {
	return string(e.version)
}

func (e *ContractBase) String() string {
	return e.Version() + "/" + e.Name() + "@" + e.Address().String()
}

func NewContractMagic[C any, T any](constructor contractConstructorFunc[T], address common.Address, backend bind.ContractBackend, name NameType, version VersionType) (*C, error) {
	contractBind, err := constructor(address, backend)
	if err != nil {
		log.Errorf("failed to bind contract %s at address %s. Err:%w", name, address.String(), err)
		return nil, err
	}
	tmp := new(C)
	values := reflect.ValueOf(tmp).Elem()
	values.FieldByIndex([]int{0}).Set(reflect.ValueOf(contractBind))
	values.FieldByIndex([]int{1}).Set(reflect.ValueOf(NewContractBase(address, backend, name, version)))
	return tmp, nil
}
