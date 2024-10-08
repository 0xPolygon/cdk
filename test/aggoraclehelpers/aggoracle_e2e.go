package aggoraclehelpers

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry-paris/polygonzkevmbridgev2"
	gerContractL1 "github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/globalexitrootnopush0"
	gerContractEVMChain "github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/pessimisticglobalexitrootnopush0"
	"github.com/0xPolygon/cdk/aggoracle"
	"github.com/0xPolygon/cdk/aggoracle/chaingersender"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

const (
	NetworkIDL2        = uint32(1)
	syncBlockChunkSize = 10
	retries            = 3
	periodRetry        = time.Millisecond * 100
)

type AggoracleWithEVMChainEnv struct {
	L1Client         *simulated.Backend
	L2Client         *simulated.Backend
	L1InfoTreeSync   *l1infotreesync.L1InfoTreeSync
	GERL1Contract    *gerContractL1.Globalexitrootnopush0
	GERL1Addr        common.Address
	GERL2Contract    *gerContractEVMChain.Pessimisticglobalexitrootnopush0
	GERL2Addr        common.Address
	AuthL1           *bind.TransactOpts
	AuthL2           *bind.TransactOpts
	AggOracle        *aggoracle.AggOracle
	AggOracleSender  aggoracle.ChainSender
	ReorgDetector    *reorgdetector.ReorgDetector
	BridgeL1Contract *polygonzkevmbridgev2.Polygonzkevmbridgev2
	BridgeL1Addr     common.Address
	BridgeL2Contract *polygonzkevmbridgev2.Polygonzkevmbridgev2
	BridgeL2Addr     common.Address
	NetworkIDL2      uint32
	EthTxManMockL2   *helpers.EthTxManagerMock
}

func SetupAggoracleWithEVMChain(t *testing.T) *AggoracleWithEVMChainEnv {
	t.Helper()

	ctx := context.Background()
	l1Client, syncer, gerL1Contract, gerL1Addr, bridgeL1Contract, bridgeL1Addr, authL1, rd := CommonSetup(t)
	sender, l2Client, gerL2Contract, gerL2Addr, bridgeL2Contract, bridgeL2Addr, authL2, ethTxManMockL2 := EVMSetup(t)
	oracle, err := aggoracle.New(
		log.GetDefaultLogger(), sender,
		l1Client.SClient, syncer,
		etherman.LatestBlock, time.Millisecond*20) //nolint:mnd
	require.NoError(t, err)
	go oracle.Start(ctx)

	return &AggoracleWithEVMChainEnv{
		L1Client:         l1Client.Backend,
		L2Client:         l2Client,
		L1InfoTreeSync:   syncer,
		GERL1Contract:    gerL1Contract,
		GERL1Addr:        gerL1Addr,
		GERL2Contract:    gerL2Contract,
		GERL2Addr:        gerL2Addr,
		AuthL1:           authL1,
		AuthL2:           authL2,
		AggOracle:        oracle,
		AggOracleSender:  sender,
		ReorgDetector:    rd,
		BridgeL1Contract: bridgeL1Contract,
		BridgeL1Addr:     bridgeL1Addr,
		BridgeL2Contract: bridgeL2Contract,
		BridgeL2Addr:     bridgeL2Addr,
		NetworkIDL2:      NetworkIDL2,
		EthTxManMockL2:   ethTxManMockL2,
	}
}

func CommonSetup(t *testing.T) (
	*helpers.TestClient,
	*l1infotreesync.L1InfoTreeSync,
	*gerContractL1.Globalexitrootnopush0,
	common.Address,
	*polygonzkevmbridgev2.Polygonzkevmbridgev2,
	common.Address,
	*bind.TransactOpts,
	*reorgdetector.ReorgDetector,
) {
	t.Helper()

	// Config and spin up
	ctx := context.Background()

	l1Client := helpers.NewTestClient(t, NetworkIDL2)
	authL1 := l1Client.UserAuth()
	bridgeL1Addr, bridgeL1Contract := l1Client.Polygonzkevmbridgev2()
	gerL1Addr, gerL1Contract := l1Client.Globalexitrootnopush0()

	// Reorg detector
	dbPathReorgDetector := t.TempDir()
	reorg, err := reorgdetector.New(l1Client.SClient, reorgdetector.Config{DBPath: dbPathReorgDetector})
	require.NoError(t, err)

	// Syncer
	dbPathSyncer := path.Join(t.TempDir(), "file::memory:?cache=shared")
	syncer, err := l1infotreesync.New(ctx, dbPathSyncer,
		gerL1Addr, common.Address{},
		syncBlockChunkSize, etherman.LatestBlock,
		reorg, l1Client.SClient,
		time.Millisecond, 0, periodRetry, retries, l1infotreesync.FlagAllowWrongContractsAddrs)
	require.NoError(t, err)
	go syncer.Start(ctx)

	return l1Client, syncer, gerL1Contract, gerL1Addr, bridgeL1Contract, bridgeL1Addr, authL1, reorg
}

func EVMSetup(t *testing.T) (
	aggoracle.ChainSender,
	*simulated.Backend,
	*gerContractEVMChain.Pessimisticglobalexitrootnopush0,
	common.Address,
	*polygonzkevmbridgev2.Polygonzkevmbridgev2,
	common.Address,
	*bind.TransactOpts,
	*helpers.EthTxManagerMock,
) {
	t.Helper()

	l2Client := helpers.NewTestClient(t, NetworkIDL2)
	authL2 := l2Client.UserAuth()
	bridgeL2Addr, bridgeL2Sc := l2Client.Polygonzkevmbridgev2()
	gerL2Addr, gerL2Sc := l2Client.Pessimisticglobalexitrootnopush0()

	ethTxManMock := helpers.NewEthTxManMock(t, l2Client.Backend, authL2)
	sender, err := chaingersender.NewEVMChainGERSender(log.GetDefaultLogger(),
		gerL2Addr, authL2.From, l2Client.SClient, ethTxManMock, 0, time.Millisecond*50) //nolint:mnd
	require.NoError(t, err)

	return sender, l2Client.Backend, gerL2Sc, gerL2Addr, bridgeL2Sc, bridgeL2Addr, authL2, ethTxManMock
}
