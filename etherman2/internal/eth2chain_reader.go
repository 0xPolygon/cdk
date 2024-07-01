package internal

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum"
)

type Etherman2ChainReader struct {
	L1EthClient interface {
		ethereum.ChainReader
		ethereum.ChainIDReader
	}
}

/*
func LoadContract[T any](address common.Address, backend bind.ContractBackend, funcConstructor func(common.Address, bind.ContractBackend) (T, error)) (T, error) {
	zkevm, err := funcConstructor(address, backend)
	if err != nil {
		var nilObject T
		return nilObject, err
	}
	return zkevm, nil
}
*/

func NewEtherman2ChainReader(l1client interface{}) (*Etherman2ChainReader, error) {
	client, ok := l1client.(interface {
		ethereum.ChainReader
		ethereum.ChainIDReader
	})
	if !ok {
		return nil, errors.New("l1client does not implement required interfaces")
	}
	return &Etherman2ChainReader{L1EthClient: client}, nil
}

func (e *Etherman2ChainReader) LastL1BlockNumber(ctx context.Context) (uint64, error) {
	if e == nil || e.L1EthClient == nil {
		return 0, nil
	}
	block, err := e.L1EthClient.BlockByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}
	return block.NumberU64(), nil
}
func (e *Etherman2ChainReader) ChainID(ctx context.Context) (uint64, error) {
	if e == nil || e.L1EthClient == nil {
		return 0, nil
	}
	id, err := e.L1EthClient.ChainID(ctx)
	if err != nil {
		return 0, err
	}
	return id.Uint64(), nil
}
