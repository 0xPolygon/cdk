package helpers

import (
	"context"
	"math/big"
	"path"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/l2-sovereign-chain-paris/globalexitrootmanagerl2sovereignchain"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/l2-sovereign-chain-paris/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/l2-sovereign-chain-paris/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/aggoracle"
	"github.com/0xPolygon/cdk/aggoracle/chaingersender"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/test/contracts/transparentupgradableproxy"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

type AggoracleWithEVMChain struct {
	L1Environment
	L2Environment
	AggOracle   *aggoracle.AggOracle
	NetworkIDL2 uint32
}

// CommonEnvironment contains common setup results used in both L1 and L2 network setups.
type CommonEnvironment struct {
	SimBackend     *simulated.Backend
	GERAddr        common.Address
	BridgeContract *polygonzkevmbridgev2.Polygonzkevmbridgev2
	BridgeAddr     common.Address
	Auth           *bind.TransactOpts
	ReorgDetector  *reorgdetector.ReorgDetector
	BridgeSync     *bridgesync.BridgeSync
}

// L1Environment contains setup results for L1 network.
type L1Environment struct {
	CommonEnvironment
	GERContract  *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2
	InfoTreeSync *l1infotreesync.L1InfoTreeSync
}

// L2Environment contains setup results for L1 network.
type L2Environment struct {
	CommonEnvironment
	GERContract      *globalexitrootmanagerl2sovereignchain.Globalexitrootmanagerl2sovereignchain
	AggoracleSender  aggoracle.ChainSender
	EthTxManagerMock *EthTxManagerMock
}

// NewE2EEnvWithEVML2 creates a new E2E environment with EVM L1 and L2 chains.
func NewE2EEnvWithEVML2(t *testing.T) *AggoracleWithEVMChain {
	t.Helper()

	ctx := context.Background()
	// Setup L1
	l1Setup := L1Setup(t)

	// Setup L2 EVM
	l2Setup := L2Setup(t)

	oracle, err := aggoracle.New(
		log.GetDefaultLogger(), l2Setup.AggoracleSender,
		l1Setup.SimBackend.Client(), l1Setup.InfoTreeSync,
		etherman.LatestBlock, time.Millisecond*20, //nolint:mnd
	)
	require.NoError(t, err)
	go oracle.Start(ctx)

	return &AggoracleWithEVMChain{
		L1Environment: *l1Setup,
		L2Environment: *l2Setup,
		AggOracle:     oracle,
		NetworkIDL2:   rollupID,
	}
}

// L1Setup creates a new L1 environment.
func L1Setup(t *testing.T) *L1Environment {
	t.Helper()

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
		time.Millisecond, 0, periodRetry,
		retries, l1infotreesync.FlagAllowWrongContractsAddrs,
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

	return &L1Environment{
		CommonEnvironment: CommonEnvironment{
			SimBackend:     l1Client,
			GERAddr:        gerL1Addr,
			BridgeContract: bridgeL1Contract,
			BridgeAddr:     bridgeL1Addr,
			Auth:           authL1,
			ReorgDetector:  rdL1,
			BridgeSync:     bridgeL1Sync,
		},
		GERContract:  gerL1Contract,
		InfoTreeSync: l1InfoTreeSync,
	}
}

// L2Setup creates a new L2 environment.
func L2Setup(t *testing.T) *L2Environment {
	t.Helper()

	l2Client, authL2, gerL2Addr, gerL2Contract,
		bridgeL2Addr, bridgeL2Contract := newSimulatedEVML2SovereignChain(t)

	ethTxManagerMock := NewEthTxManMock(t, l2Client, authL2)

	const gerCheckFrequency = time.Millisecond * 50
	sender, err := chaingersender.NewEVMChainGERSender(
		log.GetDefaultLogger(), gerL2Addr, l2Client.Client(),
		ethTxManagerMock, 0, gerCheckFrequency,
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

	return &L2Environment{
		CommonEnvironment: CommonEnvironment{
			SimBackend:     l2Client,
			GERAddr:        gerL2Addr,
			BridgeContract: bridgeL2Contract,
			BridgeAddr:     bridgeL2Addr,
			Auth:           authL2,
			ReorgDetector:  rdL2,
			BridgeSync:     bridgeL2Sync,
		},
		GERContract:      gerL2Contract,
		AggoracleSender:  sender,
		EthTxManagerMock: ethTxManagerMock,
	}
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

	deployerAuth, err := CreateAccount(big.NewInt(chainID))
	require.NoError(t, err)

	client, setup := NewSimulatedBackend(t, nil, deployerAuth)

	ctx := context.Background()
	nonce, err := client.Client().PendingNonceAt(ctx, setup.DeployerAuth.From)
	require.NoError(t, err)

	// DeployBridge function sends two transactions (bridge and proxy contract deployment)
	calculatedGERAddr := crypto.CreateAddress(setup.DeployerAuth.From, nonce+2) //nolint:mnd

	err = setup.DeployBridge(client, calculatedGERAddr, 0)
	require.NoError(t, err)

	gerAddr, _, gerContract, err := polygonzkevmglobalexitrootv2.DeployPolygonzkevmglobalexitrootv2(
		setup.DeployerAuth, client.Client(),
		setup.UserAuth.From, setup.BridgeProxyAddr)
	require.NoError(t, err)
	client.Commit()

	require.Equal(t, calculatedGERAddr, gerAddr)

	return client, setup.UserAuth, gerAddr, gerContract, setup.BridgeProxyAddr, setup.BridgeProxyContract
}

func newSimulatedEVML2SovereignChain(t *testing.T) (
	*simulated.Backend,
	*bind.TransactOpts,
	common.Address,
	*globalexitrootmanagerl2sovereignchain.Globalexitrootmanagerl2sovereignchain,
	common.Address,
	*polygonzkevmbridgev2.Polygonzkevmbridgev2,
) {
	t.Helper()

	deployerAuth, err := CreateAccount(big.NewInt(chainID))
	require.NoError(t, err)

	premineBalance, ok := new(big.Int).SetString(defaultBalance, base10)
	require.True(t, ok)

	const deployedContractsCount = 3
	l2BridgeAddr := crypto.CreateAddress(deployerAuth.From, deployedContractsCount)

	genesisAllocMap := map[common.Address]types.Account{l2BridgeAddr: {Balance: premineBalance}}
	client, setup := NewSimulatedBackend(t, genesisAllocMap, deployerAuth)

	// Deploy L2 GER manager contract
	gerL2Addr, _, _, err := globalexitrootmanagerl2sovereignchain.DeployGlobalexitrootmanagerl2sovereignchain(
		setup.DeployerAuth, client.Client(), setup.BridgeProxyAddr)
	require.NoError(t, err)
	client.Commit()

	// Prepare initialize data that are going to be called by the L2 GER proxy contract
	gerL2Abi, err := globalexitrootmanagerl2sovereignchain.Globalexitrootmanagerl2sovereignchainMetaData.GetAbi()
	require.NoError(t, err)
	require.NotNil(t, gerL2Abi)

	gerL2InitData, err := gerL2Abi.Pack("initialize", setup.UserAuth.From, setup.UserAuth.From)
	require.NoError(t, err)

	// Deploy L2 GER manager proxy contract
	gerProxyAddr, _, _, err := transparentupgradableproxy.DeployTransparentupgradableproxy(
		setup.DeployerAuth,
		client.Client(),
		gerL2Addr,
		setup.DeployerAuth.From,
		gerL2InitData,
	)
	require.NoError(t, err)
	client.Commit()

	// Create L2 GER manager contract binding
	gerL2Contract, err := globalexitrootmanagerl2sovereignchain.NewGlobalexitrootmanagerl2sovereignchain(
		gerProxyAddr, client.Client())
	require.NoError(t, err)

	err = setup.DeployBridge(client, gerProxyAddr, rollupID)
	require.NoError(t, err)
	require.Equal(t, l2BridgeAddr, setup.BridgeProxyAddr)

	bridgeGERAddr, err := setup.BridgeProxyContract.GlobalExitRootManager(nil)
	require.NoError(t, err)
	require.Equal(t, gerProxyAddr, bridgeGERAddr)

	return client, setup.UserAuth, gerProxyAddr, gerL2Contract, setup.BridgeProxyAddr, setup.BridgeProxyContract
}
