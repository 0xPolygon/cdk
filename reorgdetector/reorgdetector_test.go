package reorgdetector

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

func newSimulatedL1(t *testing.T, auth *bind.TransactOpts) *simulated.Backend {
	t.Helper()

	balance, _ := new(big.Int).SetString("10000000000000000000000000", 10) //nolint:gomnd

	blockGasLimit := uint64(999999999999999999) //nolint:gomnd
	client := simulated.NewBackend(map[common.Address]types.Account{
		auth.From: {
			Balance: balance,
		},
	}, simulated.WithBlockGasLimit(blockGasLimit))
	client.Commit()

	return client
}

func newTestDir(tb testing.TB) string {
	tb.Helper()

	dir := fmt.Sprintf("/tmp/reorgdetector-temp_%v", time.Now().UTC().Format(time.RFC3339Nano))
	err := os.Mkdir(dir, 0775)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		require.NoError(tb, os.RemoveAll(dir))
	})

	return dir
}

func Test_ReorgDetector(t *testing.T) {
	const subID = "test"

	ctx := context.Background()

	// Simulated L1
	privateKeyL1, err := crypto.GenerateKey()
	require.NoError(t, err)
	authL1, err := bind.NewKeyedTransactorWithChainID(privateKeyL1, big.NewInt(1337))
	require.NoError(t, err)
	clientL1 := newSimulatedL1(t, authL1)
	require.NoError(t, err)

	// Create test DB dir
	testDir := newTestDir(t)

	reorgDetector, err := New(clientL1.Client(), Config{DBPath: testDir, CheckReorgsInterval: time.Millisecond * 100})
	require.NoError(t, err)

	err = reorgDetector.Start(ctx)
	require.NoError(t, err)

	reorgSub, err := reorgDetector.Subscribe(subID)
	require.NoError(t, err)

	remainingHeader, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit()) // Block 2
	require.NoError(t, err)
	err = reorgDetector.AddBlockToTrack(ctx, subID, remainingHeader.Number.Uint64(), remainingHeader.Hash()) // Adding block 2
	require.NoError(t, err)
	reorgHeader, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit()) // Block 3
	require.NoError(t, err)
	firstHeaderAfterReorg, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit()) // Block 4
	require.NoError(t, err)
	err = reorgDetector.AddBlockToTrack(ctx, subID, firstHeaderAfterReorg.Number.Uint64(), firstHeaderAfterReorg.Hash()) // Adding block 4
	require.NoError(t, err)
	header, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit()) // Block 5
	require.NoError(t, err)
	err = reorgDetector.AddBlockToTrack(ctx, subID, header.Number.Uint64(), header.Hash()) // Adding block 5
	require.NoError(t, err)
	err = clientL1.Fork(reorgHeader.Hash()) // Reorg on block 3
	require.NoError(t, err)
	clientL1.Commit() // Next block 4 after reorg on block 3
	clientL1.Commit() // Block 5
	clientL1.Commit() // Block 6

	// Expect reorg on added blocks 4 -> all further blocks should be removed
	select {
	case firstReorgedBlock := <-reorgSub.ReorgedBlock:
		reorgSub.ReorgProcessed <- true
		require.Equal(t, firstHeaderAfterReorg.Number.Uint64(), firstReorgedBlock)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for reorg")
	}

	// just wait a little for completion
	time.Sleep(time.Second / 5)

	headersList, ok := reorgDetector.trackedBlocks[subID]
	require.True(t, ok)
	require.Equal(t, 1, headersList.len()) // Only block 2 left
	require.Equal(t, remainingHeader.Hash(), headersList.get(2).Hash)
}
