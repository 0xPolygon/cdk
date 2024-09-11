package contracts

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonrollupmanager"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/etherman/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type GlobalExitRootBananaType struct {
	*polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2
	*ContractBase
}

type RollupManagerBananaType struct {
	*polygonrollupmanager.Polygonrollupmanager
	*ContractBase
}

type RollupBananaType struct {
	*polygonvalidiumetrog.Polygonvalidiumetrog
	*ContractBase
}

type ContractsBanana struct {
	GlobalExitRoot GlobalExitRootBananaType
	Rollup         RollupBananaType
	RollupManager  RollupManagerBananaType
}

func NewContractsBanana(cfg config.L1Config, backend bind.ContractBackend) (*ContractsBanana, error) {
	ger, err := NewContractMagic[GlobalExitRootBananaType](
		polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2,
		cfg.GlobalExitRootManagerAddr,
		backend, ContractNameGlobalExitRoot, VersionBanana)
	if err != nil {
		return nil, err
	}
	rollup, err := NewContractMagic[RollupBananaType](
		polygonvalidiumetrog.NewPolygonvalidiumetrog, cfg.ZkEVMAddr,
		backend, ContractNameRollup, VersionBanana)
	if err != nil {
		return nil, err
	}

	rollupManager, err := NewContractMagic[RollupManagerBananaType](
		polygonrollupmanager.NewPolygonrollupmanager, cfg.RollupManagerAddr,
		backend, ContractNameRollupManager, VersionBanana)
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
	return "RollupManager: " + c.RollupManager.String() + "\nGlobalExitRoot: " +
		c.GlobalExitRoot.String() + "\nRollup: " + c.Rollup.String()
}
