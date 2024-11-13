package bridgesync_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestBridgeEventE2E(t *testing.T) {
	env := helpers.NewE2EEnvWithEVML2(t)
	ctx := context.Background()

	// Send bridge txs
	expectedBridges := []bridgesync.Bridge{}

	for i := 0; i < 100; i++ {
		bridge := bridgesync.Bridge{
			BlockNum:           uint64(4 + i),
			Amount:             big.NewInt(0),
			DepositCount:       uint32(i),
			DestinationNetwork: 3,
			DestinationAddress: common.HexToAddress("f00"),
			Metadata:           []byte{},
		}
		tx, err := env.BridgeL1Contract.BridgeAsset(
			env.AuthL1,
			bridge.DestinationNetwork,
			bridge.DestinationAddress,
			bridge.Amount,
			bridge.OriginAddress,
			false, nil,
		)
		require.NoError(t, err)
		env.L1Client.Commit()
		receipt, err := env.L1Client.Client().TransactionReceipt(ctx, tx.Hash())
		require.NoError(t, err)
		require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)
		expectedBridges = append(expectedBridges, bridge)
	}

	// TODO: replace wait with the one in  package helpers once merged
	// Wait for syncer to catch up
	lb, err := env.L1Client.Client().BlockNumber(ctx)
	require.NoError(t, err)
	helpers.RequireProcessorUpdated(t, env.BridgeL1Sync, lb)

	// Get bridges
	lastBlock, err := env.L1Client.Client().BlockNumber(ctx)
	require.NoError(t, err)
	actualBridges, err := env.BridgeL1Sync.GetBridges(ctx, 0, lastBlock)
	require.NoError(t, err)

	// Assert bridges
	require.Equal(t, expectedBridges, actualBridges)
}

// TODO: test claims and claims + bridges combined
