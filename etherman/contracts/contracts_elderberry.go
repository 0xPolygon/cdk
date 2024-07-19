package contracts

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonrollupmanager"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/etherman/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type GlobalExitRootElderberryType = ContractBase[polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2]
type RollupElderberryType = ContractBase[polygonvalidiumetrog.Polygonvalidiumetrog]
type RollupManagerElderberryType = ContractBase[polygonrollupmanager.Polygonrollupmanager]

type ContractsElderberry struct {
	GlobalExitRoot GlobalExitRootElderberryType
	Rollup         RollupElderberryType
	RollupManager  RollupManagerElderberryType
}

func NewContractsElderberry(cfg config.L1Config, backend bind.ContractBackend) (*ContractsElderberry, error) {
	ger, err := NewContractBase(polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2, cfg.GlobalExitRootManagerAddr, backend, ContractNameGlobalExitRoot, VersionElderberry)
	if err != nil {
		return nil, err
	}
	rollup, err := NewContractBase(polygonvalidiumetrog.NewPolygonvalidiumetrog, cfg.ZkEVMAddr, backend, ContractNameRollup, VersionElderberry)
	if err != nil {
		return nil, err
	}

	rollupManager, err := NewContractBase(polygonrollupmanager.NewPolygonrollupmanager, cfg.RollupManagerAddr, backend, ContractNameRollupManager, VersionElderberry)
	if err != nil {
		return nil, err
	}

	return &ContractsElderberry{
		GlobalExitRoot: *ger,
		Rollup:         *rollup,
		RollupManager:  *rollupManager,
	}, nil
}

func (c *ContractsElderberry) String() string {
	return "RollupManager: " + c.RollupManager.String() + "\nGlobalExitRoot: " + c.GlobalExitRoot.String() + "\nRollup: " + c.Rollup.String()

}
