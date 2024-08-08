package bridgesync_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry-paris/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

func newSimulatedClient(auth *bind.TransactOpts) (
	client *simulated.Backend,
	bridgeAddr common.Address,
	bridgeContract *polygonzkevmbridgev2.Polygonzkevmbridgev2,
	err error,
) {
	// ctx := context.Background()
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
	client.Commit()
	return
}

func TestBridgeEventE2E(t *testing.T) {
	ctx := context.Background()
	dbPathSyncer := t.TempDir()
	dbPathReorg := t.TempDir()
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	require.NoError(t, err)
	client, bridgeAddr, bridgeSc, err := newSimulatedClient(auth)
	require.NoError(t, err)
	rd, err := reorgdetector.New(ctx, client.Client(), dbPathReorg)
	go rd.Start(ctx)

	syncer, err := bridgesync.NewL1(ctx, dbPathSyncer, bridgeAddr, 10, etherman.LatestBlock, rd, client.Client(), 0, time.Millisecond*10)
	require.NoError(t, err)
	go syncer.Start(ctx)

	// Send bridge txs
	expectedBridges := []bridgesync.Bridge{}
	for i := 0; i < 100; i++ {
		bridge := bridgesync.Bridge{
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
		client.Commit()
		receipt, err := client.Client().TransactionReceipt(ctx, tx.Hash())
		require.NoError(t, err)
		require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)
		expectedBridges = append(expectedBridges, bridge)
	}

	// Wait for syncer to catch up
	syncerUpToDate := false
	var errMsg string
	for i := 0; i < 10; i++ {
		lpb, err := syncer.GetLastProcessedBlock(ctx)
		require.NoError(t, err)
		lb, err := client.Client().BlockNumber(ctx)
		require.NoError(t, err)
		if lpb == lb {
			syncerUpToDate = true
			break
		}
		time.Sleep(time.Millisecond * 10)
		errMsg = fmt.Sprintf("last block from client: %d, last block from syncer: %d", lb, lpb)
	}
	require.True(t, syncerUpToDate, errMsg)

	// Get bridges
	lastBlcok, err := client.Client().BlockNumber(ctx)
	require.NoError(t, err)
	events, err := syncer.GetClaimsAndBridges(ctx, 0, lastBlcok)
	require.NoError(t, err)
	actualBridges := []bridgesync.Bridge{}
	for _, event := range events {
		if event.Bridge != nil {
			actualBridges = append(actualBridges, *event.Bridge)
		}
	}

	// Assert bridges
	require.Equal(t, expectedBridges, actualBridges)
}

// TODO: test claims and claims + bridges combined