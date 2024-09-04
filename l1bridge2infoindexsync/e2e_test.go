package l1bridge2infoindexsync_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry-paris/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry-paris/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/bridgesync"
	cdktypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1bridge2infoindexsync"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/test/contracts/transparentupgradableproxy"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func newSimulatedClient(authDeployer, authCaller *bind.TransactOpts) (
	client *simulated.Backend,
	gerAddr common.Address,
	bridgeAddr common.Address,
	gerContract *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
	bridgeContract *polygonzkevmbridgev2.Polygonzkevmbridgev2,
	err error,
) {
	ctx := context.Background()
	balance, _ := new(big.Int).SetString("10000000000000000000000000", 10) //nolint:gomnd
	genesisAlloc := map[common.Address]types.Account{
		authDeployer.From: {
			Balance: balance,
		},
		authCaller.From: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999) //nolint:gomnd
	client = simulated.NewBackend(genesisAlloc, simulated.WithBlockGasLimit(blockGasLimit))

	bridgeImplementationAddr, _, _, err := polygonzkevmbridgev2.DeployPolygonzkevmbridgev2(authDeployer, client.Client())
	if err != nil {
		return
	}
	client.Commit()

	nonce, err := client.Client().PendingNonceAt(ctx, authDeployer.From)
	if err != nil {
		return
	}
	precalculatedAddr := crypto.CreateAddress(authDeployer.From, nonce+1)
	bridgeABI, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	if err != nil {
		return
	}
	if bridgeABI == nil {
		err = errors.New("GetABI returned nil")
		return
	}
	dataCallProxy, err := bridgeABI.Pack("initialize",
		uint32(0),        // networkIDMainnet
		common.Address{}, // gasTokenAddressMainnet"
		uint32(0),        // gasTokenNetworkMainnet
		precalculatedAddr,
		common.Address{},
		[]byte{}, // gasTokenMetadata
	)
	if err != nil {
		return
	}
	bridgeAddr, _, _, err = transparentupgradableproxy.DeployTransparentupgradableproxy(
		authDeployer,
		client.Client(),
		bridgeImplementationAddr,
		authDeployer.From,
		dataCallProxy,
	)
	if err != nil {
		return
	}
	client.Commit()
	bridgeContract, err = polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridgeAddr, client.Client())
	if err != nil {
		return
	}
	checkGERAddr, err := bridgeContract.GlobalExitRootManager(&bind.CallOpts{})
	if err != nil {
		return
	}
	if precalculatedAddr != checkGERAddr {
		err = errors.New("error deploying bridge")
		return
	}

	gerAddr, _, gerContract, err = polygonzkevmglobalexitrootv2.DeployPolygonzkevmglobalexitrootv2(
		authDeployer, client.Client(), authCaller.From, bridgeAddr,
	)
	if err != nil {
		return
	}
	client.Commit()

	if precalculatedAddr != gerAddr {
		err = errors.New("error calculating addr")
	}
	return
}

func TestE2E(t *testing.T) {
	ctx := context.Background()
	dbPathBridgeSync := path.Join(t.TempDir(), "tmp.sqlite")
	dbPathL1Sync := path.Join(t.TempDir(), "tmp.sqlite")
	dbPathReorg := t.TempDir()
	dbPathL12InfoSync := t.TempDir()

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	authDeployer, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	require.NoError(t, err)
	privateKey, err = crypto.GenerateKey()
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	require.NoError(t, err)
	require.NotEqual(t, authDeployer.From, auth.From)
	client, gerAddr, bridgeAddr, gerSc, bridgeSc, err := newSimulatedClient(authDeployer, auth)
	require.NoError(t, err)
	rd, err := reorgdetector.New(client.Client(), reorgdetector.Config{DBPath: dbPathReorg, CheckReorgsInterval: cdktypes.NewDuration(time.Second)})
	require.NoError(t, err)
	require.NoError(t, rd.Start(ctx))

	testClient := helpers.TestClient{ClientRenamed: client.Client()}
	bridgeSync, err := bridgesync.NewL1(ctx, dbPathBridgeSync, bridgeAddr, 10, etherman.LatestBlock, rd, testClient, 0, time.Millisecond*10, 0, 0)
	require.NoError(t, err)
	go bridgeSync.Start(ctx)

	l1Sync, err := l1infotreesync.New(
		ctx,
		dbPathL1Sync,
		gerAddr,
		common.Address{},
		10,
		etherman.SafeBlock,
		rd,
		client.Client(),
		time.Millisecond,
		0,
		time.Millisecond,
		3,
	)
	require.NoError(t, err)
	go l1Sync.Start(ctx)

	bridge2InfoSync, err := l1bridge2infoindexsync.New(dbPathL12InfoSync, bridgeSync, l1Sync, client.Client(), 0, 0, time.Millisecond)
	require.NoError(t, err)
	go bridge2InfoSync.Start(ctx)

	// Send bridge txs
	expectedIndex := -1
	for i := 0; i < 10; i++ {
		bridge := bridgesync.Bridge{
			Amount:             big.NewInt(0),
			DestinationNetwork: 3,
			DestinationAddress: common.HexToAddress("f00"),
		}
		_, err := bridgeSc.BridgeAsset(
			auth,
			bridge.DestinationNetwork,
			bridge.DestinationAddress,
			bridge.Amount,
			bridge.OriginAddress,
			true, nil,
		)
		require.NoError(t, err)
		expectedIndex++
		client.Commit()

		// Wait for block to be finalised
		updateAtBlock, err := client.Client().BlockNumber(ctx)
		require.NoError(t, err)
		for {
			lastFinalisedBlock, err := client.Client().BlockByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
			require.NoError(t, err)
			if lastFinalisedBlock.NumberU64() >= updateAtBlock {
				break
			}
			client.Commit()
			time.Sleep(time.Microsecond)
		}

		// Wait for syncer to catch up
		syncerUpToDate := false
		var errMsg string
		lb, err := client.Client().BlockByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
		require.NoError(t, err)
		for i := 0; i < 10; i++ {
			lpb, err := bridge2InfoSync.GetLastProcessedBlock(ctx)
			require.NoError(t, err)
			if lpb == lb.NumberU64() {
				syncerUpToDate = true
				break
			}
			time.Sleep(time.Millisecond * 100)
			errMsg = fmt.Sprintf("last block from client: %d, last block from syncer: %d", lb.NumberU64(), lpb)
		}
		require.True(t, syncerUpToDate, errMsg)

		actualIndex, err := bridge2InfoSync.GetL1InfoTreeIndexByDepositCount(ctx, uint32(i))
		require.NoError(t, err)
		require.Equal(t, uint32(expectedIndex), actualIndex)

		if i%2 == 1 {
			// Update L1 info tree without a bridge on L1
			_, err = gerSc.UpdateExitRoot(auth, common.HexToHash(strconv.Itoa(i)))
			require.NoError(t, err)
			expectedIndex++
			client.Commit()
		}
	}
}
