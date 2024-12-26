package lastgersync_test

import (
	"context"
	"fmt"
	"math/big"
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
	l1Env, l2Env := helpers.NewL1EnvWithL2EVM(t)
	dbPathSyncer := path.Join(t.TempDir(), "lastgersyncTestE2E.sqlite")
	syncer, err := lastgersync.New(
		ctx,
		dbPathSyncer,
		l2Env.ReorgDetector,
		l2Env.SimBackend.Client(),
		l2Env.GERAddr,
		l1Env.InfoTreeSync,
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
		_, err := l1Env.GERContract.UpdateExitRoot(l1Env.Auth, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)
		l1Env.SimBackend.Commit()
		time.Sleep(time.Millisecond * 150)
		expectedGER, err := l1Env.GERContract.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		_, err = l2Env.GERContract.InsertGlobalExitRoot(l2Env.Auth, expectedGER)
		require.NoError(t, err)
		l2Env.SimBackend.Commit()
		gerIndex, err := l2Env.GERContract.GlobalExitRootMap(nil, expectedGER)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(int64(i+1)), gerIndex, fmt.Sprintf("iteration %d, GER: %s is not updated on L2", i, common.Bytes2Hex(expectedGER[:])))

		// Wait for syncer to catch up
		lb, err := l2Env.SimBackend.Client().BlockNumber(ctx)
		require.NoError(t, err)
		helpers.RequireProcessorUpdated(t, syncer, lb)

		e, err := syncer.GetFirstGERAfterL1InfoTreeIndex(ctx, uint32(i))
		require.NoError(t, err, fmt.Sprint("iteration: ", i))
		require.Equal(t, common.Hash(expectedGER), e.GlobalExitRoot, fmt.Sprint("iteration: ", i))
	}
}
