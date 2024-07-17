package contracts

import (
	"fmt"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type ContractGlobalExitRootElderberry struct {
	*ContractBase[polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2]
}

func NewContractGlobalExitRootElderberry(address common.Address, backend bind.ContractBackend) (*ContractGlobalExitRootElderberry, error) {
	base, err := NewContractBase(polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2, address, backend, ContractNameGlobalExitRoot, VersionElderberry)
	if err != nil {
		return nil, err
	}
	return &ContractGlobalExitRootElderberry{
		ContractBase: base,
	}, nil
}

func (e *ContractGlobalExitRootElderberry) L1InfoIndexToRoot(indexLeaf uint32) (common.Hash, error) {
	errC := fmt.Errorf("Contract %s doesn't implement L1InfoIndexToRoot. Err:%w", e.String(), ErrNotImplemented)
	log.Errorf("%v", errC)
	return common.Hash{}, errC
}
