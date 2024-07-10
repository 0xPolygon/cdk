package localbridgesync

import (
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonzkevmbridgemock"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonzkevmglobalexitrootl2mock"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

type simulatedL2 struct {
	V1, V2 common.Address
	Client *simulated.Backend
	auth   *bind.TransactOpts
	ger    *polygonzkevmglobalexitrootl2mock.Polygonzkevmglobalexitrootl2mock
}

func newSimulatedL2(t *testing.T, auth *bind.TransactOpts) *simulatedL2 {
	require.NotNil(t, auth)
	// 10000000 ETH in wei
	balance, _ := new(big.Int).SetString("10000000000000000000000000", 10) //nolint:gomnd
	address := auth.From
	genesisAlloc := map[common.Address]types.Account{
		address: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999) //nolint:gomnd
	client := simulated.NewBackend(genesisAlloc, simulated.WithBlockGasLimit(blockGasLimit))

	// Deploy contract bridge v1 mock
	v1Addr, _, v1Contract, err := polygonzkevmbridgemock.DeployPolygonzkevmbridgemock(auth, client.Client())
	require.NoError(t, err)
	gerAddr, _, gerContract, err := polygonzkevmglobalexitrootl2mock.DeployPolygonzkevmglobalexitrootl2mock(auth, client.Client(), v1Addr)
	require.NoError(t, err)
	_, err = v1Contract.Initialize(auth, 1, gerAddr, common.Address{})
	require.NoError(t, err)

	return &simulatedL2{
		ger: gerContract,
	}
}
