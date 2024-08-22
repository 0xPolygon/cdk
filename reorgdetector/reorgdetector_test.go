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
	const reorgDepth = 2

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

	reorgSub, err := reorgDetector.Subscribe("test")
	require.NoError(t, err)

	ch := make(chan *types.Header, 10)
	headerSub, err := clientL1.Client().SubscribeNewHead(ctx, ch)
	require.NoError(t, err)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-headerSub.Err():
				return
			case header := <-ch:
				err := reorgDetector.AddBlockToTrack(ctx, "test", header.Number.Uint64(), header.Hash())
				require.NoError(t, err)
			}
		}
	}()

	expectedReorgBlocks := make(map[uint64]bool)
	lastReorgOn := int64(0)
	for i := 1; lastReorgOn <= produceBlocks; i++ {
		block := clientL1.Commit()
		time.Sleep(time.Millisecond * 100)

		header, err := clientL1.Client().HeaderByHash(ctx, block)
		require.NoError(t, err)
		headerNumber := header.Number.Int64()

		// Reorg every "reorgPeriod" blocks with "reorgDepth" blocks depth
		if headerNumber > lastReorgOn && headerNumber%reorgPeriod == 0 {
			lastReorgOn = headerNumber

			reorgBlock, err := clientL1.Client().BlockByNumber(ctx, big.NewInt(headerNumber-reorgDepth))
			require.NoError(t, err)

			expectedReorgBlocks[reorgBlock.NumberU64()] = false

			err = clientL1.Fork(reorgBlock.Hash())
			require.NoError(t, err)
		}
	}

	// Commit some blocks to ensure reorgs are detected
	for i := 0; i < reorgPeriod; i++ {
		clientL1.Commit()
	}

	fmt.Println("expectedReorgBlocks", expectedReorgBlocks)

	for blk := range reorgSub.ReorgedBlock {
		reorgSub.ReorgProcessed <- true
		fmt.Println("reorgSub.FirstReorgedBlock", blk)
	}

	for range expectedReorgBlocks {
		firstReorgedBlock := <-reorgSub.ReorgedBlock
		reorgSub.ReorgProcessed <- true

		fmt.Println("firstReorgedBlock", firstReorgedBlock)

		_, ok := expectedReorgBlocks[firstReorgedBlock]
		require.True(t, ok)
		//require.False(t, processed)

		expectedReorgBlocks[firstReorgedBlock] = true
	}

	for _, processed := range expectedReorgBlocks {
		require.True(t, processed)
	}
}
