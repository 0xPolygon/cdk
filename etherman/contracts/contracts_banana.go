package contracts

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonrollupmanager"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/etherman/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type GlobalExitRootBananaType = ContractBase[polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2]
type RollupBananaType = ContractBase[polygonvalidiumetrog.Polygonvalidiumetrog]
type RollupManagerBananaType = ContractBase[polygonrollupmanager.Polygonrollupmanager]

type ContractsBanana struct {
	GlobalExitRoot GlobalExitRootBananaType
	Rollup         RollupBananaType
	RollupManager  RollupManagerBananaType
}

func NewContractsBanana(cfg config.L1Config, backend bind.ContractBackend) (*ContractsBanana, error) {
	ger, err := NewContractBase(polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2, cfg.GlobalExitRootManagerAddr, backend, ContractNameGlobalExitRoot, VersionBanana)
	if err != nil {
		return nil, err
	}
	rollup, err := NewContractBase(polygonvalidiumetrog.NewPolygonvalidiumetrog, cfg.ZkEVMAddr, backend, ContractNameRollup, VersionBanana)
	if err != nil {
		return nil, err
	}

	rollupManager, err := NewContractBase(polygonrollupmanager.NewPolygonrollupmanager, cfg.RollupManagerAddr, backend, ContractNameRollupManager, VersionBanana)
	if err != nil {
		return nil, err
	}

	return &ContractsBanana{
		GlobalExitRoot: *ger,
		Rollup:         *rollup,
		RollupManager:  *rollupManager,
	}, nil
}

func (c *ContractsBanana) String() string {
	return "RollupManager: " + c.RollupManager.String() + "\nGlobalExitRoot: " + c.GlobalExitRoot.String() + "\nRollup: " + c.Rollup.String()

}
