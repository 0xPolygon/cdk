package bridgesync_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestBridgeEventE2E(t *testing.T) {
	const (
		totalBridges          = 1000
		totalReorgs           = 700
		maxReorgDepth         = 5
		reorgEveryXIterations = 7 // every X blocks go back [1,maxReorgDepth] blocks
	)
	env := helpers.NewE2EEnvWithEVML2(t)
	ctx := context.Background()

	// Send bridge txs
	bridgesSent := 0
	reorgs := 0
	expectedBridges := []bridgesync.Bridge{}
	lastDepositCount := uint32(0)
	for i := 1; i > 0; i++ {
		// Send bridge
		bridge := bridgesync.Bridge{
			Amount:             big.NewInt(0),
			DepositCount:       lastDepositCount,
			DestinationNetwork: uint32(i),
			DestinationAddress: common.HexToAddress("f00"),
			Metadata:           []byte{},
		}
		lastDepositCount++
		tx, err := env.BridgeL1Contract.BridgeAsset(
			env.AuthL1,
			bridge.DestinationNetwork,
			bridge.DestinationAddress,
			bridge.Amount,
			bridge.OriginAddress,
			true, nil,
		)
		require.NoError(t, err)
		helpers.CommitBlocks(t, env.L1Client, 1, time.Millisecond)
		bn, err := env.L1Client.Client().BlockNumber(ctx)
		require.NoError(t, err)
		bridge.BlockNum = bn
		receipt, err := env.L1Client.Client().TransactionReceipt(ctx, tx.Hash())
		require.NoError(t, err)
		require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)
		expectedBridges = append(expectedBridges, bridge)
		bridgesSent++

		// Trigger reorg
		if i%reorgEveryXIterations == 0 {
			blocksToReorg := 1 + i%maxReorgDepth
			bn, err := env.L1Client.Client().BlockNumber(ctx)
			require.NoError(t, err)
			helpers.Reorg(t, env.L1Client, uint64(blocksToReorg))
			// Clean expected bridges
			lastValidBlock := bn - uint64(blocksToReorg)
			reorgEffective := false
			for i := len(expectedBridges) - 1; i >= 0; i-- {
				if expectedBridges[i].BlockNum > lastValidBlock {
					log.Debugf("removing expectedBridge with depositCount %d due to reorg", expectedBridges[i].DepositCount)
					lastDepositCount = expectedBridges[i].DepositCount
					expectedBridges = expectedBridges[0:i]
					reorgEffective = true
					bridgesSent--
				}
			}
			if reorgEffective {
				reorgs++
				log.Debug("reorgs: ", reorgs)
			}
		}

		// Finish condition
		if bridgesSent >= totalBridges && reorgs >= totalReorgs {
			break
		}
	}

	// Wait for syncer to catch up
	time.Sleep(time.Second * 2) // sleeping since the processor could be up to date, but have pending reorgs
	lb, err := env.L1Client.Client().BlockNumber(ctx)
	require.NoError(t, err)
	helpers.RequireProcessorUpdated(t, env.BridgeL1Sync, lb)

	// Get bridges
	lastBlock, err := env.L1Client.Client().BlockNumber(ctx)
	require.NoError(t, err)
	actualBridges, err := env.BridgeL1Sync.GetBridges(ctx, 0, lastBlock)
	require.NoError(t, err)

	// Assert bridges
	expectedRoot, err := env.BridgeL1Contract.GetRoot(nil)
	require.NoError(t, err)
	root, err := env.BridgeL1Sync.GetExitRootByIndex(ctx, expectedBridges[len(expectedBridges)-1].DepositCount)
	require.NoError(t, err)
	require.Equal(t, common.Hash(expectedRoot).Hex(), root.Hash.Hex())
	require.Equal(t, expectedBridges, actualBridges)
}
