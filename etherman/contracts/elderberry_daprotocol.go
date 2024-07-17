package contracts

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/idataavailabilityprotocol"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractDAProtocolElderberry struct {
	*ContractBase[idataavailabilityprotocol.Idataavailabilityprotocol]
}

func NewContractDAProtocolElderberry(address common.Address, backend bind.ContractBackend) (*ContractDAProtocolElderberry, error) {
	base, err := NewContractBase(idataavailabilityprotocol.NewIdataavailabilityprotocol, address, backend, ContractNameDAProtocol, VersionElderberry)
	if err != nil {
		return nil, err
	}
	return &ContractDAProtocolElderberry{
		ContractBase: base,
	}, nil
}
