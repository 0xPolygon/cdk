package l1infotreesync_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/globalexitrootnopush0"
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
	gerContract *globalexitrootnopush0.Globalexitrootnopush0,
	verifyContract *verifybatchesmock.Verifybatchesmock,
	err error,
) {
	ctx := context.Background()
	balance, _ := new(big.Int).SetString("10000000000000000000000000", 10) //nolint:gomnd
	address := auth.From
	genesisAlloc := map[common.Address]types.Account{
		address: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999) //nolint:gomnd
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

	gerAddr, _, gerContract, err = globalexitrootnopush0.DeployGlobalexitrootnopush0(auth, client.Client(), verifyAddr, auth.From)
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
	dbPath := t.TempDir()
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
		// Let the processor catch up
		time.Sleep(time.Millisecond * 100)
		receipt, err := client.Client().TransactionReceipt(ctx, tx.Hash())
		require.NoError(t, err)
		require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

		expectedGER, err := gerSc.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		info, err := syncer.GetInfoByIndex(ctx, uint32(i+1))
		require.NoError(t, err)
		require.Equal(t, common.Hash(expectedGER), info.GlobalExitRoot, fmt.Sprintf("index: %d", i))
		require.Equal(t, receipt.BlockNumber.Uint64(), info.BlockNumber)

		expectedRoot, err := gerSc.GetRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		actualRoot, err := syncer.GetL1InfoTreeRootByIndex(ctx, uint32(i+1))
		require.NoError(t, err)
		require.Equal(t, common.Hash(expectedRoot), actualRoot)
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
			require.True(t, len(receipt.Logs) == 1+i%2)

			expectedRollupExitRoot, err := verifySC.GetRollupExitRoot(&bind.CallOpts{Pending: false})
			require.NoError(t, err)
			actualRollupExitRoot, err := syncer.GetLastRollupExitRoot(ctx)
			require.NoError(t, err)
			require.Equal(t, common.Hash(expectedRollupExitRoot), actualRollupExitRoot, fmt.Sprintf("rollupID: %d, i: %d", rollupID, i))
		}
	}
}

func TestFinalised(t *testing.T) {
	ctx := context.Background()
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	require.NoError(t, err)
	client, _, _, _, _, err := newSimulatedClient(auth)
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		client.Commit()
	}

	n4, err := client.Client().HeaderByNumber(ctx, big.NewInt(-4))
	require.NoError(t, err)
	fmt.Println("-4", n4.Number)
	n3, err := client.Client().HeaderByNumber(ctx, big.NewInt(-3))
	require.NoError(t, err)
	fmt.Println("-3", n3.Number)
	n2, err := client.Client().HeaderByNumber(ctx, big.NewInt(-2))
	require.NoError(t, err)
	fmt.Println("-2", n2.Number)
	n1, err := client.Client().HeaderByNumber(ctx, big.NewInt(-1))
	require.NoError(t, err)
	fmt.Println("-1", n1.Number)
	n0, err := client.Client().HeaderByNumber(ctx, nil)
	require.NoError(t, err)
	fmt.Println("0", n0.Number)
	fmt.Printf("amount of blocks latest - finalised: %d", n0.Number.Uint64()-n3.Number.Uint64())
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
	dbPathSyncer := t.TempDir()
	dbPathReorg := t.TempDir()
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	require.NoError(t, err)
	client, gerAddr, verifyAddr, gerSc, verifySC, err := newSimulatedClient(auth)
	require.NoError(t, err)
	rd, err := reorgdetector.New(ctx, client.Client(), dbPathReorg)
	go rd.Start(ctx)
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
				client.Fork(reorgBlock.Hash())
			}
		}
	}

	syncerUpToDate := false
	var errMsg string
	for i := 0; i < 50; i++ {
		lpb, err := syncer.GetLastProcessedBlock(ctx)
		require.NoError(t, err)
		lb, err := client.Client().BlockNumber(ctx)
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
	require.Equal(t, common.Hash(expectedRollupExitRoot), actualRollupExitRoot)

	// Assert L1 Info tree root
	expectedL1InfoRoot, err := gerSc.GetRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	expectedGER, err := gerSc.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	index, actualL1InfoRoot, err := syncer.GetLastL1InfoTreeRootAndIndex(ctx)
	require.NoError(t, err)
	info, err := syncer.GetInfoByIndex(ctx, index)
	require.NoError(t, err, fmt.Sprintf("index: %d", index))

	require.Equal(t, common.Hash(expectedL1InfoRoot), actualL1InfoRoot)
	require.Equal(t, common.Hash(expectedGER), info.GlobalExitRoot, fmt.Sprintf("%+v", info))
}
