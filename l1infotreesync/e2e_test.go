package l1infotreesync

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
	rdm := NewReorgDetectorMock(t)
	rdm.On("Subscribe", reorgDetectorID).Return(&reorgdetector.Subscription{}, nil)
	rdm.On("AddBlockToTrack", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	client, gerAddr, verifyAddr, gerSc, verifySC, err := newSimulatedClient(auth)
	require.NoError(t, err)
	syncer, err := New(ctx, dbPath, gerAddr, verifyAddr, 10, etherman.LatestBlock, rdm, client.Client(), time.Millisecond, 0, 100*time.Millisecond, 3)
	require.NoError(t, err)
	go syncer.Start(ctx)

	// Update GER 10 times
	for i := 0; i < 10; i++ {
		tx, err := gerSc.UpdateExitRoot(auth, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)
		client.Commit()
		// Let the processor catch up
		time.Sleep(time.Millisecond * 10)
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
		actualRoot, err := syncer.GetL1InfoTreeRootByIndex(ctx, uint32(i))
		require.NoError(t, err)
		require.Equal(t, common.Hash(expectedRoot), actualRoot)
	}

	// Update 10 rollups 10 times
	for rollupID := uint32(1); rollupID < 10; rollupID++ {
		for i := 0; i < 10; i++ {
			newLocalExitRoot := common.HexToHash(strconv.Itoa(int(rollupID)) + "ffff" + strconv.Itoa(i))
			tx, err := verifySC.VerifyBatches(auth, rollupID, 0, newLocalExitRoot, common.Hash{}, true)
			require.NoError(t, err)
			client.Commit()
			// Let the processor catch up
			time.Sleep(time.Millisecond * 100)
			receipt, err := client.Client().TransactionReceipt(ctx, tx.Hash())
			require.NoError(t, err)
			require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)
			require.True(t, len(receipt.Logs) == 2)

			expectedRollupExitRoot, err := verifySC.GetRollupExitRoot(&bind.CallOpts{Pending: false})
			require.NoError(t, err)
			actualRollupExitRoot, err := syncer.GetLastRollupExitRoot(ctx)
			require.NoError(t, err)
			require.Equal(t, common.Hash(expectedRollupExitRoot), actualRollupExitRoot, fmt.Sprintf("rollupID: %d, i: %d", rollupID, i))
		}
	}
}
