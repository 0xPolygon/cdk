package etherman

import (
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type ContractsBanana struct {
	//DAProtocol     *ContractDAProtocolBanana
	GlobalExitRoot *contracts.ContractGlobalExitRootBanana
	RollupManager  *contracts.ContractRollupManangerBanana
	ZkEVM          *contracts.ContractRollupBanana
}

func (c *ContractsBanana) String() string {
	return "RollupManager: " + c.RollupManager.String() + "\nGlobalExitRoot: " + c.GlobalExitRoot.String() + "\nZkEVM: " + c.ZkEVM.String()

}

func newContractsBanana(cfg L1Config, backend bind.ContractBackend) (*ContractsBanana, error) {

	globalExitRoot, err := contracts.NewContractGlobalExitRootBanana(cfg.GlobalExitRootManagerAddr, backend)
	if err != nil {
		return nil, err
	}

	rollupManager, err := contracts.NewContractRollupManangerBanana(cfg.RollupManagerAddr, backend)
	if err != nil {
		return nil, err
	}

	zkEVM, err := contracts.NewContractRollupBanana(cfg.ZkEVMAddr, backend)
	if err != nil {
		return nil, err
	}

	return &ContractsBanana{
		GlobalExitRoot: globalExitRoot,
		RollupManager:  rollupManager,
		ZkEVM:          zkEVM,
	}, nil
}

type ContractsElderberry struct {

	//DAProtocol     *ContractDAProtocolBanana
	RollupManager  *contracts.ContractRollupManangerElderberry
	GlobalExitRoot *contracts.ContractGlobalExitRootElderberry
	ZkEVM          *contracts.ContractRollupElderberry
}

func (c *ContractsElderberry) String() string {
	return "RollupManager: " + c.RollupManager.String() + "\nGlobalExitRoot: " + c.GlobalExitRoot.String() + "\nZkEVM: " + c.ZkEVM.String()

}

func newContractsElderberry(cfg L1Config, backend bind.ContractBackend) (*ContractsElderberry, error) {

	globalExitRoot, err := contracts.NewContractGlobalExitRootElderberry(cfg.GlobalExitRootManagerAddr, backend)
	if err != nil {
		return nil, err
	}
	rollupManager, err := contracts.NewContractRollupManangerElderberry(cfg.RollupManagerAddr, backend)
	if err != nil {
		return nil, err
	}

	zkEVM, err := contracts.NewContractRollupElderberry(cfg.ZkEVMAddr, backend)
	if err != nil {
		return nil, err
	}

	return &ContractsElderberry{
		GlobalExitRoot: globalExitRoot,
		RollupManager:  rollupManager,
		ZkEVM:          zkEVM,
	}, nil
}

type Contracts struct {
	Banana          ContractsBanana
	Elderberry      ContractsElderberry
	contractVersion contracts.VersionType
}

func (c *Contracts) String() string {
	return "default_contract_version: " + string(c.contractVersion) + "\n  Banana: \n" + c.Banana.String() + "\n Elderberry:\n" + c.Elderberry.String()
}

func NewContracts(cfg L1Config, backend bind.ContractBackend, defaultVersion contracts.VersionType) (*Contracts, error) {
	banana, err := newContractsBanana(cfg, backend)
	if err != nil {
		return nil, err
	}

	elderberry, err := newContractsElderberry(cfg, backend)
	if err != nil {
		return nil, err
	}

	return &Contracts{
		Banana:          *banana,
		Elderberry:      *elderberry,
		contractVersion: defaultVersion,
	}, nil
}

func (c *Contracts) RollupManager(specificVersion *contracts.VersionType) contracts.RollupManagerContractor {
	useVersion := c.contractVersion
	if specificVersion != nil {
		useVersion = *specificVersion
	}
	if useVersion == contracts.VersionBanana {
		return c.Banana.RollupManager
	} else if useVersion == contracts.VersionElderberry {
		return c.Elderberry.RollupManager
	} else {
		log.Errorf("RollupManager unknown version %s", useVersion)
		return nil
	}
}

func (c *Contracts) GlobalExitRoot(specificVersion *contracts.VersionType) contracts.GlobalExitRootContractor {
	useVersion := c.contractVersion
	if specificVersion != nil {
		useVersion = *specificVersion
	}
	if useVersion == contracts.VersionBanana {
		return c.Banana.GlobalExitRoot
	} else if useVersion == contracts.VersionElderberry {
		return c.Elderberry.GlobalExitRoot
	} else {
		log.Errorf("RollupManager unknown version %s", useVersion)
		return nil
	}
}

func (c *Contracts) ZkEVM(specificVersion *contracts.VersionType) contracts.RollupContractor {
	useVersion := c.contractVersion
	if specificVersion != nil {
		useVersion = *specificVersion
	}
	if useVersion == contracts.VersionBanana {
		return c.Banana.ZkEVM
	} else if useVersion == contracts.VersionElderberry {
		return c.Elderberry.ZkEVM
	} else {
		log.Errorf("ZkEVM unknown version %s", useVersion)
		return nil
	}
}
