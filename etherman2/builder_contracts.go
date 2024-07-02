package etherman2

import (
	"log"

	internal_contracts "github.com/0xPolygon/cdk/etherman2/internal/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Etherman2ContractBuilder struct {
	build struct {
		result *Contracts
	}
	params struct {
		rpcURL           string
		RollupElderberry *common.Address
	}
}

func NewEtherman2ContractBuilder(rpcURL string) *Etherman2ContractBuilder {
	res := &Etherman2ContractBuilder{}
	res.params.rpcURL = rpcURL
	return res
}

func (e *Etherman2ContractBuilder) AddRollupElderberry(address common.Address) *Etherman2ContractBuilder {
	if e.build.result != nil {
		log.Fatal("Cannot add contracts after build")
	}
	e.params.RollupElderberry = &address
	return e
}

func (e *Etherman2ContractBuilder) Build() (*Contracts, error) {
	if e.build.result != nil {
		return e.build.result, nil
	}
	l1Client, err := ethclient.Dial(e.params.rpcURL)
	if err != nil {
		return nil, err
	}
	e.build.result = &Contracts{}
	if e.params.RollupElderberry != nil {
		rollupElderberry, err := internal_contracts.NewContractRollupElderberry(*e.params.RollupElderberry, l1Client)
		if err != nil {
			return nil, err
		}
		e.build.result.RollupElderberry = rollupElderberry
	}
	return e.build.result, nil
}
