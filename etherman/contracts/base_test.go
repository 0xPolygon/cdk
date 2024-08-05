package contracts_test

import (
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestNewContractMagic(t *testing.T) {
	ger, err := contracts.NewContractMagic[contracts.GlobalExitRootBananaType](polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2, common.Address{}, nil, contracts.ContractNameGlobalExitRoot, contracts.VersionBanana)
	require.NoError(t, err)
	require.NotNil(t, ger)
}
