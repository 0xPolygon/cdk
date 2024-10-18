package aggoracle_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	gerContractL1 "github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/globalexitrootnopush0"
	"github.com/0xPolygon/cdk/aggoracle"
	"github.com/0xPolygon/cdk/test/aggoraclehelpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

func TestEVM(t *testing.T) {
	env := aggoraclehelpers.SetupAggoracleWithEVMChain(t)
	runTest(t, env.GERL1Contract, env.AggOracleSender, env.L1Client, env.AuthL1)
}

func runTest(
	t *testing.T,
	gerL1Contract *gerContractL1.Globalexitrootnopush0,
	sender aggoracle.ChainSender,
	l1Client *simulated.Backend,
	authL1 *bind.TransactOpts,
) {
	t.Helper()

	for i := 0; i < 10; i++ {
		_, err := gerL1Contract.UpdateExitRoot(authL1, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)
		l1Client.Commit()
		time.Sleep(time.Millisecond * 150)
		expectedGER, err := gerL1Contract.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		isInjected, err := sender.IsGERAlreadyInjected(expectedGER)
		require.NoError(t, err)
		require.True(t, isInjected, fmt.Sprintf("iteration %d, GER: %s", i, common.Bytes2Hex(expectedGER[:])))
	}
}
