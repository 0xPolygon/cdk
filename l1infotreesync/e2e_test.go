package l1infotreesync_test

import (
	"context"
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
	mocks_l1infotreesync "github.com/0xPolygon/cdk/l1infotreesync/mocks"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/test/contracts/verifybatchesmock"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newSimulatedClient(t *testing.T) (
	*simulated.Backend,
	*bind.TransactOpts,
	common.Address,
	common.Address,
	*polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
	*verifybatchesmock.Verifybatchesmock,
) {
	t.Helper()

	ctx := context.Background()
	client, setup := helpers.SimulatedBackend(t, nil, 0)

	nonce, err := client.Client().PendingNonceAt(ctx, setup.UserAuth.From)
	require.NoError(t, err)

	precalculatedAddr := crypto.CreateAddress(setup.UserAuth.From, nonce+1)
	verifyAddr, _, verifyContract, err := verifybatchesmock.DeployVerifybatchesmock(setup.UserAuth, client.Client(), precalculatedAddr)
	require.NoError(t, err)
	client.Commit()

	gerAddr, _, gerContract, err := polygonzkevmglobalexitrootv2.DeployPolygonzkevmglobalexitrootv2(setup.UserAuth, client.Client(), verifyAddr, setup.UserAuth.From)
	require.NoError(t, err)
	client.Commit()

	require.Equal(t, precalculatedAddr, gerAddr)

	return client, setup.UserAuth, gerAddr, verifyAddr, gerContract, verifyContract
}

func TestE2E(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	dbPath := path.Join(t.TempDir(), "L1InfoTreeTest.sqlite")

	rdm := mocks_l1infotreesync.NewReorgDetectorMock(t)
	rdm.On("Subscribe", mock.Anything).Return(&reorgdetector.Subscription{}, nil)
	rdm.On("AddBlockToTrack", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	client, auth, gerAddr, verifyAddr, gerSc, verifySC := newSimulatedClient(t)
	syncer, err := l1infotreesync.New(ctx, dbPath, gerAddr, verifyAddr, 10, etherman.LatestBlock, rdm, client.Client(), time.Millisecond, 0, 100*time.Millisecond, 3,
		l1infotreesync.FlagAllowWrongContractsAddrs)
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

	// Restart syncer
	cancelCtx()
	ctx = context.Background()
	go syncer.Start(ctx)

	// Update 3 rollups (verify batches event) 3 times
	for rollupID := uint32(1); rollupID < 3; rollupID++ {
		for i := 0; i < 3; i++ {
			newLocalExitRoot := common.HexToHash(strconv.Itoa(int(rollupID)) + "ffff" + strconv.Itoa(i))
			tx, err := verifySC.VerifyBatches(auth, rollupID, 0, newLocalExitRoot, common.Hash{}, i%2 != 0)
			require.NoError(t, err)
			client.Commit()
			receipt, err := client.Client().TransactionReceipt(ctx, tx.Hash())
			require.NoError(t, err)
			require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)
			require.True(t, len(receipt.Logs) == 1+i%2+i%2)

			// Let the processor catch
			processorUpdated := false
			for i := 0; i < 30; i++ {
				lpb, err := syncer.GetLastProcessedBlock(ctx)
				require.NoError(t, err)
				if receipt.BlockNumber.Uint64() == lpb {
					processorUpdated = true
					break
				}
				time.Sleep(time.Millisecond * 10)
			}
			require.True(t, processorUpdated)

			// Assert rollup exit root
			expectedRollupExitRoot, err := verifySC.GetRollupExitRoot(&bind.CallOpts{Pending: false})
			require.NoError(t, err)
			actualRollupExitRoot, err := syncer.GetLastRollupExitRoot(ctx)
			require.NoError(t, err)
			require.Equal(t, common.Hash(expectedRollupExitRoot), actualRollupExitRoot.Hash, fmt.Sprintf("rollupID: %d, i: %d", rollupID, i))

			// Assert verify batches
			expectedVerify := l1infotreesync.VerifyBatches{
				BlockNumber:    receipt.BlockNumber.Uint64(),
				BlockPosition:  uint64(i%2 + i%2),
				RollupID:       rollupID,
				ExitRoot:       newLocalExitRoot,
				Aggregator:     auth.From,
				RollupExitRoot: expectedRollupExitRoot,
			}
			actualVerify, err := syncer.GetLastVerifiedBatches(rollupID)
			require.NoError(t, err)
			require.Equal(t, expectedVerify, *actualVerify)
		}
	}
}

func TestWithReorgs(t *testing.T) {
	ctx := context.Background()
	dbPathSyncer := path.Join(t.TempDir(), "file::memory:?cache=shared")
	dbPathReorg := path.Join(t.TempDir(), "file::memory:?cache=shared")

	client, auth, gerAddr, verifyAddr, gerSc, verifySC := newSimulatedClient(t)

	rd, err := reorgdetector.New(client.Client(), reorgdetector.Config{DBPath: dbPathReorg, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 30)})
	require.NoError(t, err)
	require.NoError(t, rd.Start(ctx))

	syncer, err := l1infotreesync.New(ctx, dbPathSyncer, gerAddr, verifyAddr, 10, etherman.LatestBlock, rd, client.Client(), time.Millisecond, 0, time.Second, 25,
		l1infotreesync.FlagAllowWrongContractsAddrs)
	require.NoError(t, err)
	go syncer.Start(ctx)

	// Commit block
	header, err := client.Client().HeaderByHash(ctx, client.Commit()) // Block 3
	require.NoError(t, err)
	reorgFrom := header.Hash()
	fmt.Println("start from header:", header.Number)

	updateL1InfoTreeAndRollupExitTree := func(i int, rollupID uint32) {
		// Update L1 Info Tree
		_, err := gerSc.UpdateExitRoot(auth, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)

		// Update L1 Info Tree + Rollup Exit Tree
		newLocalExitRoot := common.HexToHash(strconv.Itoa(i) + "ffff" + strconv.Itoa(1))
		_, err = verifySC.VerifyBatchesTrustedAggregator(auth, rollupID, 0, newLocalExitRoot, common.Hash{}, true)
		require.NoError(t, err)

		// Update Rollup Exit Tree
		newLocalExitRoot = common.HexToHash(strconv.Itoa(i) + "ffff" + strconv.Itoa(2))
		_, err = verifySC.VerifyBatchesTrustedAggregator(auth, rollupID, 0, newLocalExitRoot, common.Hash{}, false)
		require.NoError(t, err)
	}

	// create some events and update the trees
	updateL1InfoTreeAndRollupExitTree(1, 1)

	// Block 4
	commitBlocks(t, client, 1, time.Second*5)

	// Make sure syncer is up to date
	waitForSyncerToCatchUp(ctx, t, syncer, client)

	// Assert rollup exit root
	expectedRollupExitRoot, err := verifySC.GetRollupExitRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	actualRollupExitRoot, err := syncer.GetLastRollupExitRoot(ctx)
	require.NoError(t, err)
	t.Log("exit roots", common.Hash(expectedRollupExitRoot), actualRollupExitRoot.Hash)
	require.Equal(t, common.Hash(expectedRollupExitRoot), actualRollupExitRoot.Hash)

	// Assert L1 Info tree root
	expectedL1InfoRoot, err := gerSc.GetRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	expectedGER, err := gerSc.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	actualL1InfoRoot, err := syncer.GetLastL1InfoTreeRoot(ctx)
	require.NoError(t, err)
	info, err := syncer.GetInfoByIndex(ctx, actualL1InfoRoot.Index)
	require.NoError(t, err)

	require.Equal(t, common.Hash(expectedL1InfoRoot), actualL1InfoRoot.Hash)
	require.Equal(t, common.Hash(expectedGER), info.GlobalExitRoot, fmt.Sprintf("%+v", info))

	// Forking from block 3
	err = client.Fork(reorgFrom)
	require.NoError(t, err)

	// Block 4, 5, 6 after the fork
	commitBlocks(t, client, 3, time.Millisecond*500)

	// Make sure syncer is up to date
	waitForSyncerToCatchUp(ctx, t, syncer, client)

	// Assert rollup exit root after the fork - should be zero since there are no events in the block after the fork
	expectedRollupExitRoot, err = verifySC.GetRollupExitRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	actualRollupExitRoot, err = syncer.GetLastRollupExitRoot(ctx)
	require.ErrorContains(t, err, "not found") // rollup exit tree reorged, it does not have any exits in it
	t.Log("exit roots", common.Hash(expectedRollupExitRoot), actualRollupExitRoot.Hash)
	require.Equal(t, common.Hash(expectedRollupExitRoot), actualRollupExitRoot.Hash)

	// Forking from block 3 again
	err = client.Fork(reorgFrom)
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	// create some events and update the trees
	updateL1InfoTreeAndRollupExitTree(2, 1)

	// Block 4, 5, 6, 7 after the fork
	commitBlocks(t, client, 4, time.Millisecond*100)

	// Make sure syncer is up to date
	waitForSyncerToCatchUp(ctx, t, syncer, client)

	// Assert rollup exit root after the fork - should be zero since there are no events in the block after the fork
	expectedRollupExitRoot, err = verifySC.GetRollupExitRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	actualRollupExitRoot, err = syncer.GetLastRollupExitRoot(ctx)
	require.NoError(t, err)
	t.Log("exit roots", common.Hash(expectedRollupExitRoot), actualRollupExitRoot.Hash)
	require.Equal(t, common.Hash(expectedRollupExitRoot), actualRollupExitRoot.Hash)
}

func TestStressAndReorgs(t *testing.T) {
	const (
		totalIterations       = 3
		blocksInIteration     = 140
		reorgEveryXIterations = 70
		reorgSizeInBlocks     = 2
		maxRollupID           = 31
		extraBlocksToMine     = 10
	)

	ctx := context.Background()
	dbPathSyncer := path.Join(t.TempDir(), "file:TestStressAndReorgs:memory:?cache=shared")
	dbPathReorg := path.Join(t.TempDir(), "file::memory:?cache=shared")

	client, auth, gerAddr, verifyAddr, gerSc, verifySC := newSimulatedClient(t)

	rd, err := reorgdetector.New(client.Client(), reorgdetector.Config{DBPath: dbPathReorg, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)})
	require.NoError(t, err)
	require.NoError(t, rd.Start(ctx))

	syncer, err := l1infotreesync.New(ctx, dbPathSyncer, gerAddr, verifyAddr, 10, etherman.LatestBlock, rd, client.Client(), time.Millisecond, 0, time.Second, 100,
		l1infotreesync.FlagAllowWrongContractsAddrs)
	require.NoError(t, err)
	go syncer.Start(ctx)

	updateL1InfoTreeAndRollupExitTree := func(i, j int, rollupID uint32) {
		// Update L1 Info Tree
		_, err := gerSc.UpdateExitRoot(auth, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)

		// Update L1 Info Tree + Rollup Exit Tree
		newLocalExitRoot := common.HexToHash(strconv.Itoa(i) + "ffff" + strconv.Itoa(j))
		_, err = verifySC.VerifyBatches(auth, rollupID, 0, newLocalExitRoot, common.Hash{}, true)
		require.NoError(t, err)

		// Update Rollup Exit Tree
		newLocalExitRoot = common.HexToHash(strconv.Itoa(i) + "fffa" + strconv.Itoa(j))
		_, err = verifySC.VerifyBatches(auth, rollupID, 0, newLocalExitRoot, common.Hash{}, false)
		require.NoError(t, err)
	}

	for i := 1; i <= totalIterations; i++ {
		for j := 1; j <= blocksInIteration; j++ {
			commitBlocks(t, client, 1, time.Millisecond*10)

			if j%reorgEveryXIterations == 0 {
				currentBlockNum, err := client.Client().BlockNumber(ctx)
				require.NoError(t, err)

				block, err := client.Client().BlockByNumber(ctx, big.NewInt(int64(currentBlockNum-reorgSizeInBlocks)))
				log.Debugf("reorging until block %d. Current block %d (before reorg)", block.NumberU64(), currentBlockNum)
				require.NoError(t, err)
				reorgFrom := block.Hash()
				err = client.Fork(reorgFrom)
				require.NoError(t, err)
			} else {
				updateL1InfoTreeAndRollupExitTree(i, j, uint32(j%maxRollupID)+1)
			}
		}
	}

	commitBlocks(t, client, 1, time.Millisecond*10)

	waitForSyncerToCatchUp(ctx, t, syncer, client)

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

	t.Logf("expectedL1InfoRoot: %s", common.Hash(expectedL1InfoRoot).String())
	require.Equal(t, common.Hash(expectedGER), info.GlobalExitRoot, fmt.Sprintf("%+v", info))
	require.Equal(t, common.Hash(expectedL1InfoRoot), lastRoot.Hash)
}

func waitForSyncerToCatchUp(ctx context.Context, t *testing.T, syncer *l1infotreesync.L1InfoTreeSync, client *simulated.Backend) {
	t.Helper()

	syncerUpToDate := false
	var errMsg string

	for i := 0; i < 200; i++ {
		lpb, err := syncer.GetLastProcessedBlock(ctx)
		require.NoError(t, err)
		lb, err := client.Client().BlockNumber(ctx)
		require.NoError(t, err)
		if lpb == lb {
			syncerUpToDate = true
			break
		}
		time.Sleep(time.Second / 2)
		errMsg = fmt.Sprintf("last block from client: %d, last block from syncer: %d", lb, lpb)
	}

	require.True(t, syncerUpToDate, errMsg)
}

// commitBlocks commits the specified number of blocks with the given client and waits for the specified duration after each block
func commitBlocks(t *testing.T, client *simulated.Backend, numBlocks int, waitDuration time.Duration) {
	t.Helper()

	for i := 0; i < numBlocks; i++ {
		client.Commit()
		time.Sleep(waitDuration)
	}
}
