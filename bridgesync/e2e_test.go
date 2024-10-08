package bridgesync_test

import (
	"context"
	"fmt"
	"math/big"
	"path"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestBridgeEventE2E(t *testing.T) {
	ctx := context.Background()
	dbPathSyncer := path.Join(t.TempDir(), "file::memory:?cache=shared")
	dbPathReorg := t.TempDir()

	client := helpers.NewTestClient(t)
	auth := client.UserAuth()
	bridgeAddr, bridgeSc := client.Polygonzkevmbridgev2()

	rd, err := reorgdetector.New(client.SClient, reorgdetector.Config{DBPath: dbPathReorg})
	require.NoError(t, err)

	go rd.Start(ctx) //nolint:errcheck

	syncer, err := bridgesync.NewL1(ctx, dbPathSyncer, bridgeAddr, 10, etherman.LatestBlock, rd, client, 0, time.Millisecond*10, 0, 0)
	require.NoError(t, err)

	go syncer.Start(ctx)

	// Send bridge txs
	expectedBridges := []bridgesync.Bridge{}

	for i := 0; i < 100; i++ {
		bridge := bridgesync.Bridge{
			BlockNum:           uint64(2 + i),
			Amount:             big.NewInt(0),
			DepositCount:       uint32(i),
			DestinationNetwork: 3,
			DestinationAddress: common.HexToAddress("f00"),
			Metadata:           []byte{},
		}
		tx, err := bridgeSc.BridgeAsset(
			auth,
			bridge.DestinationNetwork,
			bridge.DestinationAddress,
			bridge.Amount,
			bridge.OriginAddress,
			false, nil,
		)
		require.NoError(t, err)
		client.Commit()
		receipt, err := client.SClient.TransactionReceipt(ctx, tx.Hash())
		require.NoError(t, err)
		require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)
		expectedBridges = append(expectedBridges, bridge)
	}

	// Wait for syncer to catch up
	syncerUpToDate := false

	var errMsg string
	lb, err := client.SClient.BlockNumber(ctx)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		lpb, err := syncer.GetLastProcessedBlock(ctx)
		require.NoError(t, err)
		if lpb == lb {
			syncerUpToDate = true

			break
		}

		time.Sleep(time.Millisecond * 100)
		errMsg = fmt.Sprintf("last block from client: %d, last block from syncer: %d", lb, lpb)
	}
	require.True(t, syncerUpToDate, errMsg)

	// Get bridges
	lastBlock, err := client.SClient.BlockNumber(ctx)
	require.NoError(t, err)
	actualBridges, err := syncer.GetBridges(ctx, 0, lastBlock)
	require.NoError(t, err)

	// Assert bridges
	require.Equal(t, expectedBridges, actualBridges)
}

// TODO: test claims and claims + bridges combined
