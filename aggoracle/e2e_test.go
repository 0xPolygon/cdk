package aggoracle_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestEVM(t *testing.T) {
	env := helpers.NewE2EEnvWithEVML2(t)

	for i := 0; i < 10; i++ {
		_, err := env.GERL1Contract.UpdateExitRoot(env.AuthL1, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)
		env.L1Client.Commit()

		// wait for the GER to be processed by the L1InfoTree syncer
		time.Sleep(time.Millisecond * 100)
		expectedGER, err := env.GERL1Contract.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)

		isInjected, err := env.AggOracleSender.IsGERInjected(expectedGER)
		require.NoError(t, err)

		require.True(t, isInjected, fmt.Sprintf("iteration %d, GER: %s", i, common.Bytes2Hex(expectedGER[:])))
	}
}
