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
		blockTime             = time.Millisecond * 10
		totalBridges          = 80
		totalReorgs           = 40
		maxReorgDepth         = 2
		reorgEveryXIterations = 4 // every X blocks go back [1,maxReorgDepth] blocks
	)
	setup := helpers.NewE2EEnvWithEVML2(t)
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
			DestinationNetwork: uint32(i + 1),
			DestinationAddress: common.HexToAddress("f00"),
			Metadata:           []byte{},
		}
		lastDepositCount++
		tx, err := setup.L1Environment.BridgeContract.BridgeAsset(
			setup.L1Environment.Auth,
			bridge.DestinationNetwork,
			bridge.DestinationAddress,
			bridge.Amount,
			bridge.OriginAddress,
			true, nil,
		)
		require.NoError(t, err)
		helpers.CommitBlocks(t, setup.L1Environment.SimBackend, 1, blockTime)
		bn, err := setup.L1Environment.SimBackend.Client().BlockNumber(ctx)
		require.NoError(t, err)
		bridge.BlockNum = bn
		receipt, err := setup.L1Environment.SimBackend.Client().TransactionReceipt(ctx, tx.Hash())
		require.NoError(t, err)
		require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)
		expectedBridges = append(expectedBridges, bridge)
		bridgesSent++

		// Trigger reorg
		if i%reorgEveryXIterations == 0 {
			blocksToReorg := 1 + i%maxReorgDepth
			bn, err := setup.L1Environment.SimBackend.Client().BlockNumber(ctx)
			require.NoError(t, err)
			helpers.Reorg(t, setup.L1Environment.SimBackend, uint64(blocksToReorg))
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
	lb, err := setup.L1Environment.SimBackend.Client().BlockNumber(ctx)
	require.NoError(t, err)
	helpers.RequireProcessorUpdated(t, setup.L1Environment.BridgeSync, lb)

	// Get bridges
	lastBlock, err := setup.L1Environment.SimBackend.Client().BlockNumber(ctx)
	require.NoError(t, err)
	actualBridges, err := setup.L1Environment.BridgeSync.GetBridges(ctx, 0, lastBlock)
	require.NoError(t, err)

	// Assert bridges
	expectedRoot, err := setup.L1Environment.BridgeContract.GetRoot(nil)
	require.NoError(t, err)
	root, err := setup.L1Environment.BridgeSync.GetExitRootByIndex(ctx, expectedBridges[len(expectedBridges)-1].DepositCount)
	require.NoError(t, err)
	require.Equal(t, common.Hash(expectedRoot).Hex(), root.Hash.Hex())
	require.Equal(t, expectedBridges, actualBridges)
}
