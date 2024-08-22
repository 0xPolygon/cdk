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
	const produceBlocks = 29
	const reorgPeriod = 5
	const trackBlockPeriod = 4
	const reorgDepth = 2
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

	reorgDetector, err := New(clientL1.Client(), testDir)
	require.NoError(t, err)

	err = reorgDetector.Start(ctx)
	require.NoError(t, err)

	reorgSub, err := reorgDetector.Subscribe(subID)
	require.NoError(t, err)

	canonicalChain := make(map[uint64]common.Hash)
	trackedBlocks := make(map[uint64]common.Hash)
	lastReorgOn := uint64(0)
	for i := 1; lastReorgOn <= produceBlocks; i++ {
		block := clientL1.Commit()
		time.Sleep(time.Millisecond * 100)

		header, err := clientL1.Client().HeaderByHash(ctx, block)
		require.NoError(t, err)
		headerNumber := header.Number.Uint64()

		canonicalChain[headerNumber] = header.Hash()

		// Add block to track every "trackBlockPeriod" blocks
		if headerNumber%trackBlockPeriod == 0 {
			if _, ok := trackedBlocks[headerNumber]; !ok {
				err = reorgDetector.AddBlockToTrack(ctx, subID, header.Number.Uint64(), header.Hash())
				require.NoError(t, err)
				trackedBlocks[headerNumber] = header.Hash()
			}
		}

		// Reorg every "reorgPeriod" blocks with "reorgDepth" blocks depth
		if headerNumber > lastReorgOn && headerNumber%reorgPeriod == 0 {
			lastReorgOn = headerNumber

			reorgBlock, err := clientL1.Client().BlockByNumber(ctx, big.NewInt(int64(headerNumber-reorgDepth)))
			require.NoError(t, err)

			err = clientL1.Fork(reorgBlock.Hash())
			require.NoError(t, err)
		}
	}

	// Commit some blocks to ensure reorgs are detected
	for i := 0; i < reorgPeriod; i++ {
		clientL1.Commit()
	}

	// Expect reorgs on block
	expectReorgOn := make(map[uint64]bool)
	for num, hash := range canonicalChain {
		if _, ok := trackedBlocks[num]; !ok {
			continue
		}

		if trackedBlocks[num] != hash {
			expectReorgOn[num] = false
		}
	}

	// Wait for reorg notifications, expect len(expectReorgOn) notifications
	for range expectReorgOn {
		firstReorgedBlock := <-reorgSub.ReorgedBlock
		reorgSub.ReorgProcessed <- true

		fmt.Println("firstReorgedBlock", firstReorgedBlock)

		processed, ok := expectReorgOn[firstReorgedBlock]
		require.True(t, ok)
		require.False(t, processed)

		expectReorgOn[firstReorgedBlock] = true
	}

	// Make sure all processed
	for _, processed := range expectReorgOn {
		require.True(t, processed)
	}
}
