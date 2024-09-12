package injectedgersync_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/injectedgersync"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	ctx := context.Background()
	env := helpers.SetupAggoracleWithEVMChain(t)
	dbPathSyncer := t.TempDir()
	syncer, err := injectedgersync.New(
		ctx,
		dbPathSyncer,
		env.ReorgDetector,
		env.L2Client.Client(),
		env.GERL2Addr,
		env.L1InfoTreeSync,
		0,
		0,
		etherman.LatestBlock,
		time.Millisecond*30,
		10,
	)
	require.NoError(t, err)
	go syncer.Start(ctx)

	for i := 0; i < 10; i++ {
		// Update GER on L1
		_, err := env.GERL1Contract.UpdateExitRoot(env.AuthL1, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)
		env.L1Client.Commit()
		time.Sleep(time.Millisecond * 150)
		expectedGER, err := env.GERL1Contract.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		isInjected, err := env.AggOracleSender.IsGERAlreadyInjected(expectedGER)
		require.NoError(t, err)
		require.True(t, isInjected, fmt.Sprintf("iteration %d, GER: %s", i, common.Bytes2Hex(expectedGER[:])))

		// Wait for syncer to catch up
		syncerUpToDate := false
		var errMsg string
		for i := 0; i < 10; i++ {
			lpb, err := syncer.GetLastProcessedBlock(ctx)
			require.NoError(t, err)
			lb, err := env.L2Client.Client().BlockNumber(ctx)
			require.NoError(t, err)
			if lpb == lb {
				syncerUpToDate = true
				break
			}
			time.Sleep(time.Millisecond * 100)
			errMsg = fmt.Sprintf("last block from client: %d, last block from syncer: %d", lb, lpb)
		}
		require.True(t, syncerUpToDate, errMsg)

		injected, err := syncer.GetFirstGERAfterL1InfoTreeIndex(ctx, uint32(i))
		require.NoError(t, err)
		require.Equal(t, common.Hash(expectedGER), injected.GlobalExitRoot)
	}
}
