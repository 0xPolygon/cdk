package lastgersync_test

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/lastgersync"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	ctx := context.Background()
	env := helpers.NewE2EEnvWithEVML2(t)
	dbPathSyncer := path.Join(t.TempDir(), "file::memory:?cache=shared")
	syncer, err := lastgersync.New(
		ctx,
		dbPathSyncer,
		env.ReorgDetectorL1,
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
		lb, err := env.L2Client.Client().BlockNumber(ctx)
		require.NoError(t, err)
		helpers.RequireProcessorUpdated(t, syncer, lb)

		e, err := syncer.GetFirstGERAfterL1InfoTreeIndex(ctx, uint32(i))
		require.NoError(t, err, fmt.Sprint("iteration: ", i))
		require.Equal(t, common.Hash(expectedGER), e.GlobalExitRoot, fmt.Sprint("iteration: ", i))
	}
}
