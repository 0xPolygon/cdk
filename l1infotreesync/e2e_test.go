package l1infotreesync_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana-paris/polygonzkevmglobalexitrootv2"
	cdktypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/test/contracts/verifybatchesmock"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newSimulatedClient(auth *bind.TransactOpts) (
	client *simulated.Backend,
	gerAddr common.Address,
	verifyAddr common.Address,
	gerContract *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
	verifyContract *verifybatchesmock.Verifybatchesmock,
	err error,
) {
	ctx := context.Background()
	balance, _ := new(big.Int).SetString("10000000000000000000000000", 10)
	address := auth.From
	genesisAlloc := map[common.Address]types.Account{
		address: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999)
	client = simulated.NewBackend(genesisAlloc, simulated.WithBlockGasLimit(blockGasLimit))

	nonce, err := client.Client().PendingNonceAt(ctx, auth.From)
	if err != nil {
		return
	}
	precalculatedAddr := crypto.CreateAddress(auth.From, nonce+1)
	verifyAddr, _, verifyContract, err = verifybatchesmock.DeployVerifybatchesmock(auth, client.Client(), precalculatedAddr)
	if err != nil {
		return
	}
	client.Commit()

	gerAddr, _, gerContract, err = polygonzkevmglobalexitrootv2.DeployPolygonzkevmglobalexitrootv2(auth, client.Client(), verifyAddr, auth.From)
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
	dbPath := path.Join(t.TempDir(), "file::memory:?cache=shared")
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	require.NoError(t, err)
	rdm := l1infotreesync.NewReorgDetectorMock(t)
	rdm.On("Subscribe", mock.Anything).Return(&reorgdetector.Subscription{}, nil)
	rdm.On("AddBlockToTrack", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	client, gerAddr, verifyAddr, gerSc, verifySC, err := newSimulatedClient(auth)
	require.NoError(t, err)
	syncer, err := l1infotreesync.New(ctx, dbPath, gerAddr, verifyAddr, 10, etherman.LatestBlock, rdm, client.Client(), time.Millisecond, 0, 100*time.Millisecond, 3)
	require.NoError(t, err)
	go syncer.Start(ctx)

	// Update GER 3 times
	for i := 0; i < 3; i++ {
		tx, err := gerSc.UpdateExitRoot(auth, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)
		client.Commit()
		g, err := gerSc.L1InfoRootMap(nil, uint32(i+1))
		require.NoError(t, err)
		// Let the processor catch up
		time.Sleep(time.Millisecond * 100)
		receipt, err := client.Client().TransactionReceipt(ctx, tx.Hash())
		require.NoError(t, err)
		require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		expectedGER, err := gerSc.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		info, err := syncer.GetInfoByIndex(ctx, uint32(i))
		require.NoError(t, err)
		require.Equal(t, common.Hash(expectedGER), info.GlobalExitRoot, fmt.Sprintf("index: %d", i))
		require.Equal(t, receipt.BlockNumber.Uint64(), info.BlockNumber)

		expectedRoot, err := gerSc.GetRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		require.Equal(t, g, expectedRoot)
		actualRoot, err := syncer.GetL1InfoTreeRootByIndex(ctx, uint32(i))
		require.NoError(t, err)
		require.Equal(t, common.Hash(expectedRoot), actualRoot.Hash)
	}

	// Update 3 rollups (verify batches event) 3 times
	for rollupID := uint32(1); rollupID < 3; rollupID++ {
		for i := 0; i < 3; i++ {
			newLocalExitRoot := common.HexToHash(strconv.Itoa(int(rollupID)) + "ffff" + strconv.Itoa(i))
			tx, err := verifySC.VerifyBatches(auth, rollupID, 0, newLocalExitRoot, common.Hash{}, i%2 != 0)
			require.NoError(t, err)
			client.Commit()
			// Let the processor catch up
			time.Sleep(time.Millisecond * 100)
			receipt, err := client.Client().TransactionReceipt(ctx, tx.Hash())
			require.NoError(t, err)
			require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)
			require.True(t, len(receipt.Logs) == 1+i%2+i%2)

			expectedRollupExitRoot, err := verifySC.GetRollupExitRoot(&bind.CallOpts{Pending: false})
			require.NoError(t, err)
			actualRollupExitRoot, err := syncer.GetLastRollupExitRoot(ctx)
			require.NoError(t, err)
			require.Equal(t, common.Hash(expectedRollupExitRoot), actualRollupExitRoot.Hash, fmt.Sprintf("rollupID: %d, i: %d", rollupID, i))
		}
	}
}

func TestStressAndReorgs(t *testing.T) {
	const (
		totalIterations       = 200   // Have tested with much larger number (+10k)
		enableReorgs          = false // test fails when set to true
		reorgEveryXIterations = 53
		maxReorgDepth         = 5
		maxEventsPerBlock     = 7
		maxRollups            = 31
	)

	ctx := context.Background()
	dbPathSyncer := path.Join(t.TempDir(), "file::memory:?cache=shared")
	dbPathReorg := t.TempDir()
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	require.NoError(t, err)
	client, gerAddr, verifyAddr, gerSc, verifySC, err := newSimulatedClient(auth)
	require.NoError(t, err)
	rd, err := reorgdetector.New(client.Client(), reorgdetector.Config{DBPath: dbPathReorg, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)})
	require.NoError(t, err)
	require.NoError(t, rd.Start(ctx))
	syncer, err := l1infotreesync.New(ctx, dbPathSyncer, gerAddr, verifyAddr, 10, etherman.LatestBlock, rd, client.Client(), time.Millisecond, 0, 100*time.Millisecond, 3)
	require.NoError(t, err)
	go syncer.Start(ctx)

	for i := 0; i < totalIterations; i++ {
		for j := 0; j < i%maxEventsPerBlock; j++ {
			switch j % 3 {
			case 0: // Update L1 Info Tree
				_, err := gerSc.UpdateExitRoot(auth, common.HexToHash(strconv.Itoa(i)))
				require.NoError(t, err)
			case 1: // Update L1 Info Tree + Rollup Exit Tree
				newLocalExitRoot := common.HexToHash(strconv.Itoa(i) + "ffff" + strconv.Itoa(j))
				_, err := verifySC.VerifyBatches(auth, 1+uint32(i%maxRollups), 0, newLocalExitRoot, common.Hash{}, true)
				require.NoError(t, err)
			case 2: // Update Rollup Exit Tree
				newLocalExitRoot := common.HexToHash(strconv.Itoa(i) + "ffff" + strconv.Itoa(j))
				_, err := verifySC.VerifyBatches(auth, 1+uint32(i%maxRollups), 0, newLocalExitRoot, common.Hash{}, false)
				require.NoError(t, err)
			}
		}
		client.Commit()
		time.Sleep(time.Microsecond * 30) // Sleep just enough for goroutine to switch
		if enableReorgs && i%reorgEveryXIterations == 0 {
			reorgDepth := i%maxReorgDepth + 1
			currentBlockNum, err := client.Client().BlockNumber(ctx)
			require.NoError(t, err)
			targetReorgBlockNum := currentBlockNum - uint64(reorgDepth)
			if targetReorgBlockNum < currentBlockNum { // we are dealing with uints...
				reorgBlock, err := client.Client().BlockByNumber(ctx, big.NewInt(int64(targetReorgBlockNum)))
				require.NoError(t, err)
				err = client.Fork(reorgBlock.Hash())
				require.NoError(t, err)
			}
		}
	}

	syncerUpToDate := false
	var errMsg string
	lb, err := client.Client().BlockNumber(ctx)
	require.NoError(t, err)
	for i := 0; i < 50; i++ {
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

	// Assert rollup exit root
	expectedRollupExitRoot, err := verifySC.GetRollupExitRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	actualRollupExitRoot, err := syncer.GetLastRollupExitRoot(ctx)
	require.NoError(t, err)
	require.Equal(t, common.Hash(expectedRollupExitRoot), actualRollupExitRoot.Hash)

	// Assert L1 Info tree root
	expectedL1InfoRoot, err := gerSc.GetRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	expectedGER, err := gerSc.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	lastRoot, err := syncer.GetLastL1InfoTreeRoot(ctx)
	require.NoError(t, err)
	info, err := syncer.GetInfoByIndex(ctx, lastRoot.Index)
	require.NoError(t, err, fmt.Sprintf("index: %d", lastRoot.Index))

	require.Equal(t, common.Hash(expectedL1InfoRoot), lastRoot.Hash)
	require.Equal(t, common.Hash(expectedGER), info.GlobalExitRoot, fmt.Sprintf("%+v", info))
}
