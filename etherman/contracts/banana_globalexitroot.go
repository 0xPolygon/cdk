package contracts

import (
	"fmt"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractGlobalExitRootBanana struct {
	*ContractBase[polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2]
}

func NewContractGlobalExitRootBanana(address common.Address, backend bind.ContractBackend) (*ContractGlobalExitRootBanana, error) {
	base, err := NewContractBase(polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2, address, backend, ContractNameGlobalExitRoot, VersionBanana)
	if err != nil {
		return nil, err
	}
	return &ContractGlobalExitRootBanana{
		ContractBase: base,
	}, nil
}

func (e *ContractGlobalExitRootBanana) GetL1InfoRoot(indexL1InfoRoot uint32) (common.Hash, error) {
	// Get lastL1InfoTreeRoot (if index==0 then root=0, no call is needed)
	var (
		lastL1InfoTreeRoot common.Hash
		err                error
	)

	if indexL1InfoRoot > 0 {
		lastL1InfoTreeRoot, err = e.GetContract().L1InfoRootMap(&bind.CallOpts{Pending: false}, indexL1InfoRoot)
		if err != nil {
			log.Errorf("error calling SC globalexitroot L1InfoLeafMap: %v", err)
		}
	}

	return lastL1InfoTreeRoot, err
}

func (e *ContractGlobalExitRootBanana) L1InfoIndexToRoot(indexLeaf uint32) (common.Hash, error) {
	var (
		lastL1InfoTreeRoot common.Hash
		err                error
	)

	if indexLeaf > 0 {
		lastL1InfoTreeRoot, err = e.GetContract().L1InfoRootMap(&bind.CallOpts{Pending: false}, indexLeaf)
		if err != nil {
			errC := fmt.Errorf("error calling SC globalexitroot L1InfoIndexToRoot(%d) err: %w", indexLeaf, err)
			log.Errorf("%v", errC)
			return common.Hash{}, errC
		}
	}

	return lastL1InfoTreeRoot, err
}
