package contracts

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractRollupElderberry struct {
	*ContractBase[polygonvalidiumetrog.Polygonvalidiumetrog]
}

func NewContractRollupElderberry(address common.Address, backend bind.ContractBackend) (*ContractRollupElderberry, error) {
	base, err := NewContractBase(polygonvalidiumetrog.NewPolygonvalidiumetrog, address, backend, ContractNameRollup, VersionElderberry)
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

// LastAccInputHash gets the last acc input hash from the SC
func (e *ContractRollupElderberry) LastAccInputHash() (common.Hash, error) {
	return e.GetContract().LastAccInputHash(&bind.CallOpts{Pending: false})
}

func (e *ContractRollupElderberry) DataAvailabilityProtocol() (common.Address, error) {
	return e.GetContract().DataAvailabilityProtocol(&bind.CallOpts{Pending: false})
}

func (e *ContractRollupElderberry) TrustedSequencerURL() (string, error) {
	return e.GetContract().TrustedSequencerURL(&bind.CallOpts{Pending: false})
}
