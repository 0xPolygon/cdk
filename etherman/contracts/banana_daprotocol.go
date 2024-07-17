package contracts

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/idataavailabilityprotocol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractDAProtocolBanana struct {
	*ContractBase[idataavailabilityprotocol.Idataavailabilityprotocol]
}

func NewContractDAProtocolBanana(address common.Address, backend bind.ContractBackend) (*ContractDAProtocolBanana, error) {
	base, err := NewContractBase(idataavailabilityprotocol.NewIdataavailabilityprotocol, address, backend, ContractNameDAProtocol, VersionBanana)
	if err != nil {
		return nil, err
	}
	return &ContractDAProtocolBanana{
		ContractBase: base,
	}, nil
}
