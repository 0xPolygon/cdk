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

func Test_ReorgDetectorV2(t *testing.T) {
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

	mon := NewReorgMonitor(clientL1.Client(), 100)

	sub, err := mon.Subscribe("test")
	require.NoError(t, err)

	err = mon.Start(context.Background())
	require.NoError(t, err)

	expectedReorgBlocks := make(map[uint64]struct{})
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

			expectedReorgBlocks[reorgBlock.NumberU64()] = struct{}{}

			err = clientL1.Fork(reorgBlock.Hash())
			require.NoError(t, err)
		}
	}

	// Commit some blocks to ensure reorgs are detected
	for i := 0; i < reorgPeriod; i++ {
		clientL1.Commit()
	}

	fmt.Println("Expected reorg blocks", expectedReorgBlocks)

	for firstReorgedBlock := range sub.FirstReorgedBlock {
		sub.ReorgProcessed <- true
		fmt.Println("reorg", firstReorgedBlock)
	}

	for range expectedReorgBlocks {
		//reorg := <-sub.FirstReorgedBlock
		//sub.ReorgProcessed <- true
		//fmt.Println("reorg", reorg)
		//_, ok := expectedReorgBlocks[reorg.StartBlockHeight-1]
		//require.True(t, ok, "unexpected reorg starting from", reorg.StartBlockHeight-1)
	}
}
