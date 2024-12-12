package helpers

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/l2-sovereign-chain-paris/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/l2-sovereign-chain-paris/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/aggoracle"
	"github.com/0xPolygon/cdk/aggoracle/chaingersender"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

const (
	rollupID           = uint32(1)
	syncBlockChunkSize = 10
	retries            = 3
	periodRetry        = time.Millisecond * 100
)

type AggoracleWithEVMChainEnv struct {
	L1Client         *simulated.Backend
	L2Client         *simulated.Backend
	L1InfoTreeSync   *l1infotreesync.L1InfoTreeSync
	GERL1Contract    *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2
	GERL1Addr        common.Address
	GERL2Contract    *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2
	GERL2Addr        common.Address
	AuthL1           *bind.TransactOpts
	AuthL2           *bind.TransactOpts
	AggOracle        *aggoracle.AggOracle
	AggOracleSender  aggoracle.ChainSender
	ReorgDetectorL1  *reorgdetector.ReorgDetector
	ReorgDetectorL2  *reorgdetector.ReorgDetector
	BridgeL1Contract *polygonzkevmbridgev2.Polygonzkevmbridgev2
	BridgeL1Addr     common.Address
	BridgeL1Sync     *bridgesync.BridgeSync
	BridgeL2Contract *polygonzkevmbridgev2.Polygonzkevmbridgev2
	BridgeL2Addr     common.Address
	BridgeL2Sync     *bridgesync.BridgeSync
	NetworkIDL2      uint32
	EthTxManMockL2   *EthTxManagerMock
}

func NewE2EEnvWithEVML2(t *testing.T) *AggoracleWithEVMChainEnv {
	t.Helper()

	ctx := context.Background()
	l1Client, syncer, gerL1Contract, gerL1Addr,
		bridgeL1Contract, bridgeL1Addr, authL1, rdL1, bridgeL1Sync := CommonSetup(t)

	sender, l2Client, gerL2Contract, gerL2Addr,
		bridgeL2Contract, bridgeL2Addr, authL2,
		ethTxManMockL2, bridgeL2Sync, rdL2 := L2SetupEVM(t)
	oracle, err := aggoracle.New(
		log.GetDefaultLogger(), sender,
		l1Client.Client(), syncer,
		etherman.LatestBlock, time.Millisecond*20, //nolint:mnd
	)
	require.NoError(t, err)
	go oracle.Start(ctx)

	return &AggoracleWithEVMChainEnv{
		L1Client: l1Client,
		L2Client: l2Client,

		L1InfoTreeSync: syncer,

		GERL1Contract: gerL1Contract,
		GERL1Addr:     gerL1Addr,

		GERL2Contract: gerL2Contract,
		GERL2Addr:     gerL2Addr,

		AuthL1: authL1,
		AuthL2: authL2,

		AggOracle:       oracle,
		AggOracleSender: sender,

		ReorgDetectorL1: rdL1,
		ReorgDetectorL2: rdL2,

		BridgeL1Contract: bridgeL1Contract,
		BridgeL1Addr:     bridgeL1Addr,
		BridgeL1Sync:     bridgeL1Sync,

		BridgeL2Contract: bridgeL2Contract,
		BridgeL2Addr:     bridgeL2Addr,
		BridgeL2Sync:     bridgeL2Sync,

		NetworkIDL2:    rollupID,
		EthTxManMockL2: ethTxManMockL2,
	}
}

func CommonSetup(t *testing.T) (
	*simulated.Backend,
	*l1infotreesync.L1InfoTreeSync,
	*polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
	common.Address,
	*polygonzkevmbridgev2.Polygonzkevmbridgev2,
	common.Address,
	*bind.TransactOpts,
	*reorgdetector.ReorgDetector,
	*bridgesync.BridgeSync,
) {
	t.Helper()

	// Config and spin up
	ctx := context.Background()

	// Simulated L1
	l1Client, authL1, gerL1Addr, gerL1Contract, bridgeL1Addr, bridgeL1Contract := newSimulatedL1(t)

	// Reorg detector
	dbPathReorgDetectorL1 := path.Join(t.TempDir(), "ReorgDetectorL1.sqlite")
	rdL1, err := reorgdetector.New(l1Client.Client(), reorgdetector.Config{DBPath: dbPathReorgDetectorL1})
	require.NoError(t, err)
	go rdL1.Start(ctx) //nolint:errcheck

	// L1 info tree sync
	dbPathL1InfoTreeSync := path.Join(t.TempDir(), "L1InfoTreeSync.sqlite")
	l1InfoTreeSync, err := l1infotreesync.New(
		ctx, dbPathL1InfoTreeSync,
		gerL1Addr, common.Address{},
		syncBlockChunkSize, etherman.LatestBlock,
		rdL1, l1Client.Client(),
		time.Millisecond, 0, periodRetry, retries, l1infotreesync.FlagAllowWrongContractsAddrs,
	)
	require.NoError(t, err)

	go l1InfoTreeSync.Start(ctx)

	const (
		syncBlockChunks        = 10
		waitForNewBlocksPeriod = 10 * time.Millisecond
		originNetwork          = 1
		initialBlock           = 0
		retryPeriod            = 0
		retriesCount           = 0
	)

	// Bridge sync
	testClient := TestClient{ClientRenamed: l1Client.Client()}
	dbPathBridgeSyncL1 := path.Join(t.TempDir(), "BridgeSyncL1.sqlite")
	bridgeL1Sync, err := bridgesync.NewL1(
		ctx, dbPathBridgeSyncL1, bridgeL1Addr,
		syncBlockChunks, etherman.LatestBlock, rdL1, testClient,
		initialBlock, waitForNewBlocksPeriod, retryPeriod,
		retriesCount, originNetwork, false)
	require.NoError(t, err)

	go bridgeL1Sync.Start(ctx)

	return l1Client, l1InfoTreeSync, gerL1Contract, gerL1Addr,
		bridgeL1Contract, bridgeL1Addr, authL1, rdL1, bridgeL1Sync
}

func L2SetupEVM(t *testing.T) (
	aggoracle.ChainSender,
	*simulated.Backend,
	*polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
	common.Address,
	*polygonzkevmbridgev2.Polygonzkevmbridgev2,
	common.Address,
	*bind.TransactOpts,
	*EthTxManagerMock,
	*bridgesync.BridgeSync,
	*reorgdetector.ReorgDetector,
) {
	t.Helper()

	l2Client, authL2, gerL2Addr, gerL2Sc, bridgeL2Addr, bridgeL2Sc := newSimulatedEVMAggSovereignChain(t)
	ethTxManMock := NewEthTxManMock(t, l2Client, authL2)
	sender, err := chaingersender.NewEVMChainGERSender(
		log.GetDefaultLogger(),
		gerL2Addr, l2Client.Client(), ethTxManMock, 0, time.Millisecond*50, //nolint:mnd
	)
	require.NoError(t, err)
	ctx := context.Background()

	// Reorg detector
	dbPathReorgL2 := path.Join(t.TempDir(), "ReorgDetectorL2.sqlite")
	rdL2, err := reorgdetector.New(l2Client.Client(), reorgdetector.Config{DBPath: dbPathReorgL2})
	require.NoError(t, err)
	go rdL2.Start(ctx) //nolint:errcheck

	// Bridge sync
	dbPathL2BridgeSync := path.Join(t.TempDir(), "BridgeSyncL2.sqlite")
	testClient := TestClient{ClientRenamed: l2Client.Client()}

	const (
		syncBlockChunks        = 10
		waitForNewBlocksPeriod = 10 * time.Millisecond
		originNetwork          = 1
		initialBlock           = 0
		retryPeriod            = 0
		retriesCount           = 0
	)

	bridgeL2Sync, err := bridgesync.NewL2(
		ctx, dbPathL2BridgeSync, bridgeL2Addr, syncBlockChunks,
		etherman.LatestBlock, rdL2, testClient,
		initialBlock, waitForNewBlocksPeriod, retryPeriod,
		retriesCount, originNetwork, false)
	require.NoError(t, err)

	go bridgeL2Sync.Start(ctx)

	return sender, l2Client, gerL2Sc, gerL2Addr, bridgeL2Sc, bridgeL2Addr, authL2, ethTxManMock, bridgeL2Sync, rdL2
}

func newSimulatedL1(t *testing.T) (
	*simulated.Backend,
	*bind.TransactOpts,
	common.Address,
	*polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
	common.Address,
	*polygonzkevmbridgev2.Polygonzkevmbridgev2,
) {
	t.Helper()

	client, setup := SimulatedBackend(t, nil, 0)

	precalculatedAddr := crypto.CreateAddress(setup.DeployerAuth.From, 2) //nolint:mnd

	gerAddr, _, gerContract, err := polygonzkevmglobalexitrootv2.DeployPolygonzkevmglobalexitrootv2(
		setup.DeployerAuth, client.Client(),
		setup.UserAuth.From, setup.BridgeProxyAddr)
	require.NoError(t, err)
	client.Commit()

	require.Equal(t, precalculatedAddr, gerAddr)

	return client, setup.UserAuth, gerAddr, gerContract, setup.BridgeProxyAddr, setup.BridgeProxyContract
}

func newSimulatedEVMAggSovereignChain(t *testing.T) (
	*simulated.Backend,
	*bind.TransactOpts,
	common.Address,
	*polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
	common.Address,
	*polygonzkevmbridgev2.Polygonzkevmbridgev2,
) {
	t.Helper()

	client, setup := SimulatedBackend(t, nil, rollupID)

	// TODO: @Stefan-Ethernal deploy the GlobalExitRootManagerL2SovereignChain here instead?
	// https://github.com/0xPolygonHermez/zkevm-contracts/blob/feature/audit-remediations/contracts/v2/sovereignChains/GlobalExitRootManagerL2SovereignChain.sol
	gerAddr, _, gerContract, err := polygonzkevmglobalexitrootv2.DeployPolygonzkevmglobalexitrootv2(
		setup.DeployerAuth, client.Client(), setup.UserAuth.From, setup.BridgeProxyAddr)
	require.NoError(t, err)
	client.Commit()

	// TODO: @Stefan-Ethernal Remove?
	// globalExitRootSetterRole := common.HexToHash("0x7b95520991dfda409891be0afa2635b63540f92ee996fda0bf695a166e5c5176")
	// _, err = gerContract.GrantRole(setup.DeployerAuth, globalExitRootSetterRole, setup.UserAuth.From)
	// require.NoError(t, err)
	// client.Commit()

	// precalculatedAddr := crypto.CreateAddress(setup.DeployerAuth.From, 2) //nolint:mnd
	// hasRole, _ := gerContract.HasRole(&bind.CallOpts{Pending: false}, globalExitRootSetterRole, setup.UserAuth.From)
	// require.True(t, hasRole)
	// require.Equal(t, precalculatedAddr, gerAddr)

	return client, setup.UserAuth, gerAddr, gerContract, setup.BridgeProxyAddr, setup.BridgeProxyContract
}
