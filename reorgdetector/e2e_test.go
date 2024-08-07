package reorgdetector_test

import (
	context "context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
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

func TestE2E(t *testing.T) {
	ctx := context.Background()

	// Simulated L1
	privateKeyL1, err := crypto.GenerateKey()
	require.NoError(t, err)
	authL1, err := bind.NewKeyedTransactorWithChainID(privateKeyL1, big.NewInt(1337))
	require.NoError(t, err)
	clientL1 := newSimulatedL1(t, authL1)
	require.NoError(t, err)

	// Reorg detector
	dbPathReorgDetector := t.TempDir()
	reorgDetector, err := reorgdetector.New(ctx, clientL1.Client(), dbPathReorgDetector)
	require.NoError(t, err)

	/*ch := make(chan *types.Header, 10)
	sub, err := clientL1.Client().SubscribeNewHead(ctx, ch)
	require.NoError(t, err)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sub.Err():
				return
			case header := <-ch:
				err := reorgDetector.AddBlockToTrack(ctx, "test", header.Number.Uint64(), header.Hash())
				require.NoError(t, err)
			}
		}
	}()*/

	reorgDetector.Start(ctx)

	reorgSubscription, err := reorgDetector.Subscribe("test")
	require.NoError(t, err)

	go func() {
		firstReorgedBlock := <-reorgSubscription.FirstReorgedBlock
		fmt.Println("firstReorgedBlock", firstReorgedBlock)
	}()

	for i := 0; i < 20; i++ {
		block := clientL1.Commit()
		header, err := clientL1.Client().HeaderByHash(ctx, block)
		require.NoError(t, err)
		err = reorgDetector.AddBlockToTrack(ctx, "test", header.Number.Uint64(), header.Hash())
		require.NoError(t, err)

		// Reorg every 4 blocks with 2 blocks depth
		if i%4 == 0 {
			reorgBlock, err := clientL1.Client().BlockByNumber(ctx, big.NewInt(int64(i-2)))
			require.NoError(t, err)
			err = clientL1.Fork(reorgBlock.Hash())
			require.NoError(t, err)
		}

		time.Sleep(time.Second)
	}
}
