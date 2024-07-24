package aggoracle_test

import (
	"context"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/globalexitrootnopush0"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/pessimisticglobalexitroot"
	"github.com/0xPolygon/cdk/aggoracle"
	"github.com/0xPolygon/cdk/aggoracle/chaingersender"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

func TestEVM(t *testing.T) {
	ctx := context.Background()
	l1Client, syncer, gerL1Contract, authL1 := commonSetup(t)
	sender := evmSetup(t)
	oracle, err := aggoracle.New(sender, l1Client.Client(), syncer, etherman.LatestBlock)
	require.NoError(t, err)
	go oracle.Start(ctx)

	runTest(t, gerL1Contract, sender, l1Client, authL1)
}

func commonSetup(t *testing.T) (
	*simulated.Backend,
	*l1infotreesync.L1InfoTreeSync,
	*globalexitrootnopush0.Globalexitrootnopush0,
	*bind.TransactOpts,
) {
	// Config and spin up
	ctx := context.Background()
	// Simulated L1
	privateKeyL1, err := crypto.GenerateKey()
	require.NoError(t, err)
	authL1, err := bind.NewKeyedTransactorWithChainID(privateKeyL1, big.NewInt(1337))
	require.NoError(t, err)
	l1Client, gerL1Addr, gerL1Contract, err := newSimulatedL1(authL1)
	require.NoError(t, err)
	// Reorg detector
	dbPathReorgDetector := t.TempDir()
	reorg, err := reorgdetector.New(ctx, l1Client.Client(), dbPathReorgDetector)
	require.NoError(t, err)
	// Syncer
	dbPathSyncer := t.TempDir()
	syncer, err := l1infotreesync.New(ctx, dbPathSyncer, gerL1Addr, 10, etherman.LatestBlock, reorg, l1Client.Client(), 32)
	require.NoError(t, err)
	go syncer.Sync(ctx)

	return l1Client, syncer, gerL1Contract, authL1
}

func evmSetup(t *testing.T) aggoracle.ChainSender {
	privateKeyL2, err := crypto.GenerateKey()
	require.NoError(t, err)
	authL2, err := bind.NewKeyedTransactorWithChainID(privateKeyL2, big.NewInt(1337))
	require.NoError(t, err)
	l2Client, gerL2Addr, _, err := newSimulatedEVMAggSovereignChain(authL2)
	require.NoError(t, err)
	sender, err := chaingersender.NewEVMChainGERSender(gerL2Addr, authL2.From, l2Client.Client(), nil, 0)
	require.NoError(t, err)

	return sender
}

func newSimulatedL1(auth *bind.TransactOpts) (
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

func newSimulatedEVMAggSovereignChain(auth *bind.TransactOpts) (
	client *simulated.Backend,
	gerAddr common.Address,
	gerContract *pessimisticglobalexitroot.Pessimisticglobalexitroot,
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

	gerAddr, _, gerContract, err = pessimisticglobalexitroot.DeployPessimisticglobalexitroot(auth, client.Client(), auth.From)

	client.Commit()
	return
}

func runTest(
	t *testing.T,
	gerL1Contract *globalexitrootnopush0.Globalexitrootnopush0,
	sender aggoracle.ChainSender,
	l1Client *simulated.Backend,
	authL1 *bind.TransactOpts,
) {
	for i := 0; i < 10; i++ {
		_, err := gerL1Contract.UpdateExitRoot(authL1, common.HexToHash(strconv.Itoa(i)))
		require.NoError(t, err)
		l1Client.Commit()
		time.Sleep(time.Second * 30)
		expectedGER, err := gerL1Contract.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		isInjected, err := sender.IsGERAlreadyInjected(expectedGER)
		require.NoError(t, err)
		require.True(t, isInjected)
	}
}
