package contracts

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonvalidiumetrog"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractRollupBanana struct {
	*ContractBase[polygonvalidiumetrog.Polygonvalidiumetrog]
}

func NewContractRollupBanana(address common.Address, backend bind.ContractBackend) (*ContractRollupBanana, error) {
	base, err := NewContractBase(polygonvalidiumetrog.NewPolygonvalidiumetrog, address, backend, ContractNameRollup, VersionBanana)
	if err != nil {
		return nil, err
	}
	return &ContractRollupBanana{
		ContractBase: base,
	}, nil
}

func (e *ContractRollupBanana) TrustedSequencer() (common.Address, error) {
	return e.GetContract().TrustedSequencer(&bind.CallOpts{Pending: false})
}

// LastAccInputHash gets the last acc input hash from the SC
func (e *ContractRollupBanana) LastAccInputHash() (common.Hash, error) {
	return e.GetContract().LastAccInputHash(&bind.CallOpts{Pending: false})
}

func (e *ContractRollupBanana) DataAvailabilityProtocol() (common.Address, error) {
	return e.GetContract().DataAvailabilityProtocol(&bind.CallOpts{Pending: false})
}

func (e *ContractRollupBanana) TrustedSequencerURL() (string, error) {
	return e.GetContract().TrustedSequencerURL(&bind.CallOpts{Pending: false})
}
