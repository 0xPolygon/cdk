package reorgdetector

import (
	"context"
	"fmt"
	"math/big"
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

	reorgDetector := New(clientL1.Client())

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
				err := reorgDetector.AddBlockToTrack(ctx, "test", header.Number.Uint64(), header.Hash(), header.ParentHash)
				require.NoError(t, err)
			}
		}
	}()

	expectedReorgBlocks := make(map[uint64]bool)
	lastReorgOn := int64(0)
	for i := 1; lastReorgOn <= produceBlocks; i++ {
		block := clientL1.Commit()
		time.Sleep(time.Millisecond)

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

	for range expectedReorgBlocks {
		firstReorgedBlock := <-reorgSub.FirstReorgedBlock
		reorgSub.ReorgProcessed <- true

		fmt.Println("firstReorgedBlock", firstReorgedBlock)

		processed, ok := expectedReorgBlocks[firstReorgedBlock]
		require.True(t, ok)
		require.False(t, processed)

		expectedReorgBlocks[firstReorgedBlock] = true
	}

	for _, processed := range expectedReorgBlocks {
		require.True(t, processed)
	}
}
