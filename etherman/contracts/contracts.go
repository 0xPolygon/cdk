package contracts

import (
	"github.com/0xPolygon/cdk/etherman/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type Contracts struct {
	Banana     ContractsBanana
	Elderberry ContractsElderberry
}

func NewContracts(cfg config.L1Config, backend bind.ContractBackend) (*Contracts, error) {
	banana, err := NewContractsBanana(cfg, backend)
	if err != nil {
		return nil, err
	}

	elderberry, err := NewContractsElderberry(cfg, backend)
	if err != nil {
		return nil, err
	}

	return &Contracts{
		Banana:     *banana,
		Elderberry: *elderberry,
	}, nil
}

func (c *Contracts) String() string {
	return " Banana: \n" + c.Banana.String() + "\n Elderberry:\n" + c.Elderberry.String()
}
