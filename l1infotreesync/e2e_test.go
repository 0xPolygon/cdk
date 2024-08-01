package l1infotreesync

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/globalexitrootnopush0"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/reorgdetector"
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
	gerContract *globalexitrootnopush0.Globalexitrootnopush0,
	err error,
) {
	balance, _ := new(big.Int).SetString("10000000000000000000000000", 10) //nolint:gomnd
	address := auth.From
	genesisAlloc := map[common.Address]types.Account{
		address: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999) //nolint:gomnd
	client = simulated.NewBackend(genesisAlloc, simulated.WithBlockGasLimit(blockGasLimit))

	gerAddr, _, gerContract, err = globalexitrootnopush0.DeployGlobalexitrootnopush0(auth, client.Client(), auth.From, auth.From)

	client.Commit()
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
	client, gerAddr, gerSc, err := newSimulatedClient(auth)
	require.NoError(t, err)
	syncer, err := New(ctx, dbPath, gerAddr, 10, etherman.LatestBlock, rdm, client.Client(), time.Millisecond, 0, 100*time.Millisecond, 3)
	require.NoError(t, err)
	go syncer.Start(ctx)

	// Update GER 10 times
	// TODO: test syncer restart
	for i := 0; i < 10; i++ {
		_, err := gerSc.UpdateExitRoot(auth, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)
		client.Commit()
		// Let the processor catch up
		time.Sleep(time.Millisecond * 10)

		expectedRoot, err := gerSc.GetRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		expectedGER, err := gerSc.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		info2, err := syncer.GetInfoByIndex(ctx, uint32(i))
		require.NoError(t, err, fmt.Sprintf("index: %d, root: %s", i, common.Bytes2Hex(expectedRoot[:])))
		require.Equal(t, common.Hash(expectedGER), info2.GlobalExitRoot, fmt.Sprintf("index: %d", i))
		info, err := syncer.GetInfoByRoot(ctx, expectedRoot)
		require.NoError(t, err, fmt.Sprintf("index: %d, expected root: %s, actual root: %s", i, common.Bytes2Hex(expectedRoot[:]), info2.L1InfoTreeRoot))
		require.Equal(t, common.Hash(expectedRoot), info.L1InfoTreeRoot)
		require.Equal(t, info, info2)
	}
}
