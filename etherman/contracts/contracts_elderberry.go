package contracts

import (
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonrollupmanager"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/etherman/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type GlobalExitRootElderberryType struct {
	*polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2
	*ContractBase
}

type RollupElderberryType struct {
	*polygonvalidiumetrog.Polygonvalidiumetrog
	*ContractBase
}

type RollupManagerElderberryType struct {
	*polygonrollupmanager.Polygonrollupmanager
	*ContractBase
}

type ContractsElderberry struct {
	GlobalExitRoot GlobalExitRootElderberryType
	Rollup         RollupElderberryType
	RollupManager  RollupManagerElderberryType
}

func NewContractsElderberry(cfg config.L1Config, backend bind.ContractBackend) (*ContractsElderberry, error) {
	ger, err := NewContractMagic[GlobalExitRootElderberryType](
		polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2,
		cfg.GlobalExitRootManagerAddr,
		backend,
		ContractNameGlobalExitRoot,
		VersionElderberry,
	)
	if err != nil {
		return nil, err
	}
	rollup, err := NewContractMagic[RollupElderberryType](
		polygonvalidiumetrog.NewPolygonvalidiumetrog,
		cfg.ZkEVMAddr,
		backend,
		ContractNameRollup,
		VersionElderberry,
	)
	if err != nil {
		return nil, err
	}

	rollupManager, err := NewContractMagic[RollupManagerElderberryType](
		polygonrollupmanager.NewPolygonrollupmanager,
		cfg.RollupManagerAddr,
		backend,
		ContractNameRollupManager,
		VersionElderberry,
	)
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
	return "RollupManager: " + c.RollupManager.String() + "\nGlobalExitRoot: " +
		c.GlobalExitRoot.String() + "\nRollup: " + c.Rollup.String()
}
