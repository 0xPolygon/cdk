package etherman2

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/ethereum/go-ethereum/common"
)

type Etherman2Contract[T any] interface {
	GetContract() *T
	GetAddress() common.Address
}

type Etherman2ContractRollupCommon interface {
	TrustedSequencer() (common.Address, error)
}

type Etherman2ContractRollupElderberry interface {
	Etherman2Contract[polygonvalidiumetrog.Polygonvalidiumetrog]
	Etherman2ContractRollupCommon
}

type Contracts struct {
	RollupElderberry Etherman2ContractRollupElderberry
}
