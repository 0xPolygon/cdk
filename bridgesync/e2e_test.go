package bridgesync_test

import (
	"context"
	"fmt"
	"math/big"
	"path"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry-paris/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

func newSimulatedClient(t *testing.T, auth *bind.TransactOpts) (
	client *simulated.Backend,
	bridgeAddr common.Address,
	bridgeContract *polygonzkevmbridgev2.Polygonzkevmbridgev2,
) {
	t.Helper()
	var err error
	balance, _ := big.NewInt(0).SetString("10000000000000000000000000", 10) //nolint:gomnd
	address := auth.From
	genesisAlloc := map[common.Address]types.Account{
		address: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999) //nolint:gomnd
	client = simulated.NewBackend(genesisAlloc, simulated.WithBlockGasLimit(blockGasLimit))

	bridgeAddr, _, bridgeContract, err = polygonzkevmbridgev2.DeployPolygonzkevmbridgev2(auth, client.Client())
	require.NoError(t, err)
	client.Commit()
	return
}

func TestBridgeEventE2E(t *testing.T) {
	ctx := context.Background()
	dbPathSyncer := path.Join(t.TempDir(), "file::memory:?cache=shared")
	dbPathReorg := t.TempDir()
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	require.NoError(t, err)
	client, bridgeAddr, bridgeSc := newSimulatedClient(t, auth)
	rd, err := reorgdetector.New(client.Client(), reorgdetector.Config{DBPath: dbPathReorg})
	require.NoError(t, err)
	go rd.Start(ctx)

	testClient := helpers.TestClient{ClientRenamed: client.Client()}
	syncer, err := bridgesync.NewL1(ctx, dbPathSyncer, bridgeAddr, 10, etherman.LatestBlock, rd, testClient, 0, time.Millisecond*10, 0, 0)
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
			Metadata:           nil,
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
		receipt, err := client.Client().TransactionReceipt(ctx, tx.Hash())
		require.NoError(t, err)
		require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)
		expectedBridges = append(expectedBridges, bridge)
	}

	// Wait for syncer to catch up
	syncerUpToDate := false
	var errMsg string
	lb, err := client.Client().BlockNumber(ctx)
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
	lastBlock, err := client.Client().BlockNumber(ctx)
	require.NoError(t, err)
	actualBridges, err := syncer.GetBridges(ctx, 0, lastBlock)
	require.NoError(t, err)

	// Assert bridges
	require.Equal(t, expectedBridges, actualBridges)
}

// TODO: test claims and claims + bridges combined
