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
	setup := helpers.NewE2EEnvWithEVML2(t)

	for i := 0; i < 10; i++ {
		_, err := setup.L1Environment.GERContract.UpdateExitRoot(setup.L1Environment.Auth, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)
		setup.L1Environment.SimBackend.Commit()

		// wait for the GER to be processed by the InfoTree syncer
		time.Sleep(time.Millisecond * 100)
		expectedGER, err := setup.L1Environment.GERContract.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)

		isInjected, err := setup.L2Environment.AggoracleSender.IsGERInjected(expectedGER)
		require.NoError(t, err)

		require.True(t, isInjected, fmt.Sprintf("iteration %d, GER: %s", i, common.Bytes2Hex(expectedGER[:])))
	}
}
