package l1infotreesync

import (
	"context"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

func newSimulatedClient(auth *bind.TransactOpts) (
	client *simulated.Backend,
	gerAddr common.Address,
	gerContract *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
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

	gerAddr, _, gerContract, err = polygonzkevmglobalexitrootv2.DeployPolygonzkevmglobalexitrootv2(auth, client.Client(), auth.From, auth.From)

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
	client, gerAddr, gerSc, err := newSimulatedClient(auth)
	syncer, err := New(ctx, dbPath, gerAddr, 10, etherman.LatestBlock, rdm, client.Client(), 32)
	require.NoError(t, err)
	go syncer.Sync(ctx)

	// Update GER 10 times
	for i := 0; i < 10; i++ {
		gerSc.UpdateExitRoot(auth, common.HexToHash(strconv.Itoa(i)))
		client.Commit()
		// Let the processor catch up
		time.Sleep(time.Millisecond * 100)

		expectedRoot, err := gerSc.GetRoot(nil)
		require.NoError(t, err)
		info, err := syncer.GetInfoByRoot(ctx, expectedRoot)
		require.NoError(t, err)
		require.Equal(t, expectedRoot, info.L1InfoTreeRoot)
		info2, err := syncer.GetInfoByIndex(ctx, uint32(i))
		require.NoError(t, err)
		require.Equal(t, info, info2)
	}
}
