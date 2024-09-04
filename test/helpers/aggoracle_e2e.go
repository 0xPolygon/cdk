package helpers

import (
	"context"
	"errors"
	"fmt"
	"math/big"
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
	NetworkIDL2 = uint32(1)
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
	EthTxManMockL2   *EthTxManagerMock
}

func SetupAggoracleWithEVMChain(t *testing.T) *AggoracleWithEVMChainEnv {
	ctx := context.Background()
	l1Client, syncer, gerL1Contract, gerL1Addr, bridgeL1Contract, bridgeL1Addr, authL1, rd := CommonSetup(t)
	sender, l2Client, gerL2Contract, gerL2Addr, bridgeL2Contract, bridgeL2Addr, authL2, ethTxManMockL2 := EVMSetup(t)
	oracle, err := aggoracle.New(sender, l1Client.Client(), syncer, etherman.LatestBlock, time.Millisecond*20)
	require.NoError(t, err)
	go oracle.Start(ctx)

	return &AggoracleWithEVMChainEnv{
		L1Client:         l1Client,
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
	*simulated.Backend,
	*l1infotreesync.L1InfoTreeSync,
	*gerContractL1.Globalexitrootnopush0,
	common.Address,
	*polygonzkevmbridgev2.Polygonzkevmbridgev2,
	common.Address,
	*bind.TransactOpts,
	*reorgdetector.ReorgDetector,
) {
	// Config and spin up
	ctx := context.Background()
	// Simulated L1
	privateKeyL1, err := crypto.GenerateKey()
	require.NoError(t, err)
	authL1, err := bind.NewKeyedTransactorWithChainID(privateKeyL1, big.NewInt(1337))
	require.NoError(t, err)
	l1Client, gerL1Addr, gerL1Contract, bridgeL1Addr, bridgeL1Contract, err := newSimulatedL1(authL1)
	require.NoError(t, err)
	// Reorg detector
	dbPathReorgDetector := t.TempDir()
	reorg, err := reorgdetector.New(l1Client.Client(), reorgdetector.Config{DBPath: dbPathReorgDetector})
	require.NoError(t, err)
	// Syncer
	dbPathSyncer := path.Join(t.TempDir(), "tmp.sqlite")
	syncer, err := l1infotreesync.New(ctx, dbPathSyncer, gerL1Addr, common.Address{}, 10, etherman.LatestBlock, reorg, l1Client.Client(), time.Millisecond, 0, 100*time.Millisecond, 3)
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
	*EthTxManagerMock,
) {
	privateKeyL2, err := crypto.GenerateKey()
	require.NoError(t, err)
	authL2, err := bind.NewKeyedTransactorWithChainID(privateKeyL2, big.NewInt(1337))
	require.NoError(t, err)
	l2Client, gerL2Addr, gerL2Sc, bridgeL2Addr, bridgeL2Sc, err := newSimulatedEVMAggSovereignChain(authL2)
	require.NoError(t, err)
	ethTxManMock := NewEthTxManMock(t, l2Client, authL2)
	sender, err := chaingersender.NewEVMChainGERSender(gerL2Addr, authL2.From, l2Client.Client(), ethTxManMock, 0, time.Millisecond*50)
	require.NoError(t, err)

	return sender, l2Client, gerL2Sc, gerL2Addr, bridgeL2Sc, bridgeL2Addr, authL2, ethTxManMock
}

func newSimulatedL1(auth *bind.TransactOpts) (
	client *simulated.Backend,
	gerAddr common.Address,
	gerContract *gerContractL1.Globalexitrootnopush0,
	bridgeAddr common.Address,
	bridgeContract *polygonzkevmbridgev2.Polygonzkevmbridgev2,
	err error,
) {
	ctx := context.Background()
	privateKeyL1, err := crypto.GenerateKey()
	if err != nil {
		return
	}
	authDeployer, err := bind.NewKeyedTransactorWithChainID(privateKeyL1, big.NewInt(1337))
	balance, _ := new(big.Int).SetString("10000000000000000000000000", 10) //nolint:gomnd
	address := auth.From
	genesisAlloc := map[common.Address]types.Account{
		address: {
			Balance: balance,
		},
		authDeployer.From: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999) //nolint:gomnd
	client = simulated.NewBackend(genesisAlloc, simulated.WithBlockGasLimit(blockGasLimit))

	bridgeImplementationAddr, _, _, err := polygonzkevmbridgev2.DeployPolygonzkevmbridgev2(authDeployer, client.Client())
	if err != nil {
		return
	}
	client.Commit()

	nonce, err := client.Client().PendingNonceAt(ctx, authDeployer.From)
	if err != nil {
		return
	}
	precalculatedAddr := crypto.CreateAddress(authDeployer.From, nonce+1)
	bridgeABI, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	if err != nil {
		return
	}
	if bridgeABI == nil {
		err = errors.New("GetABI returned nil")
		return
	}
	dataCallProxy, err := bridgeABI.Pack("initialize",
		uint32(0),        // networkIDMainnet
		common.Address{}, // gasTokenAddressMainnet"
		uint32(0),        // gasTokenNetworkMainnet
		precalculatedAddr,
		common.Address{},
		[]byte{}, // gasTokenMetadata
	)
	if err != nil {
		return
	}
	bridgeAddr, _, _, err = transparentupgradableproxy.DeployTransparentupgradableproxy(
		authDeployer,
		client.Client(),
		bridgeImplementationAddr,
		authDeployer.From,
		dataCallProxy,
	)
	if err != nil {
		return
	}
	client.Commit()
	bridgeContract, err = polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridgeAddr, client.Client())
	if err != nil {
		return
	}
	checkGERAddr, err := bridgeContract.GlobalExitRootManager(&bind.CallOpts{Pending: false})
	if err != nil {
		return
	}
	if precalculatedAddr != checkGERAddr {
		err = fmt.Errorf("error deploying bridge, unexpected GER addr. Expected %s. Actual %s", precalculatedAddr.Hex(), checkGERAddr.Hex())
	}

	gerAddr, _, gerContract, err = gerContractL1.DeployGlobalexitrootnopush0(authDeployer, client.Client(), auth.From, bridgeAddr)

	client.Commit()
	if precalculatedAddr != gerAddr {
		err = fmt.Errorf("error calculating addr. Expected %s. Actual %s", precalculatedAddr.Hex(), gerAddr.Hex())
	}
	return
}

func newSimulatedEVMAggSovereignChain(auth *bind.TransactOpts) (
	client *simulated.Backend,
	gerAddr common.Address,
	gerContract *gerContractEVMChain.Pessimisticglobalexitrootnopush0,
	bridgeAddr common.Address,
	bridgeContract *polygonzkevmbridgev2.Polygonzkevmbridgev2,
	err error,
) {
	ctx := context.Background()
	privateKeyL1, err := crypto.GenerateKey()
	if err != nil {
		return
	}
	authDeployer, err := bind.NewKeyedTransactorWithChainID(privateKeyL1, big.NewInt(1337))
	balance, _ := new(big.Int).SetString("10000000000000000000000000", 10) //nolint:gomnd
	address := auth.From
	precalculatedBridgeAddr := crypto.CreateAddress(authDeployer.From, 1)
	genesisAlloc := map[common.Address]types.Account{
		address: {
			Balance: balance,
		},
		authDeployer.From: {
			Balance: balance,
		},
		precalculatedBridgeAddr: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(999999999999999999) //nolint:gomnd
	client = simulated.NewBackend(genesisAlloc, simulated.WithBlockGasLimit(blockGasLimit))

	bridgeImplementationAddr, _, _, err := polygonzkevmbridgev2.DeployPolygonzkevmbridgev2(authDeployer, client.Client())
	if err != nil {
		return
	}
	client.Commit()

	nonce, err := client.Client().PendingNonceAt(ctx, authDeployer.From)
	if err != nil {
		return
	}
	precalculatedAddr := crypto.CreateAddress(authDeployer.From, nonce+1)
	bridgeABI, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	if err != nil {
		return
	}
	if bridgeABI == nil {
		err = errors.New("GetABI returned nil")
		return
	}
	dataCallProxy, err := bridgeABI.Pack("initialize",
		NetworkIDL2,
		common.Address{}, // gasTokenAddressMainnet"
		uint32(0),        // gasTokenNetworkMainnet
		precalculatedAddr,
		common.Address{},
		[]byte{}, // gasTokenMetadata
	)
	if err != nil {
		return
	}
	bridgeAddr, _, _, err = transparentupgradableproxy.DeployTransparentupgradableproxy(
		authDeployer,
		client.Client(),
		bridgeImplementationAddr,
		authDeployer.From,
		dataCallProxy,
	)
	if err != nil {
		return
	}
	if bridgeAddr != precalculatedBridgeAddr {
		err = fmt.Errorf("error calculating bridge addr. Expected: %s. Actual: %s", precalculatedBridgeAddr, bridgeAddr)
		return
	}
	client.Commit()
	bridgeContract, err = polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridgeAddr, client.Client())
	if err != nil {
		return
	}
	checkGERAddr, err := bridgeContract.GlobalExitRootManager(&bind.CallOpts{})
	if err != nil {
		return
	}
	if precalculatedAddr != checkGERAddr {
		err = errors.New("error deploying bridge")
	}

	gerAddr, _, gerContract, err = gerContractEVMChain.DeployPessimisticglobalexitrootnopush0(authDeployer, client.Client(), auth.From)
	if err != nil {
		return
	}
	client.Commit()

	_GLOBAL_EXIT_ROOT_SETTER_ROLE := common.HexToHash("0x7b95520991dfda409891be0afa2635b63540f92ee996fda0bf695a166e5c5176")
	_, err = gerContract.GrantRole(authDeployer, _GLOBAL_EXIT_ROOT_SETTER_ROLE, auth.From)
	client.Commit()
	hasRole, _ := gerContract.HasRole(&bind.CallOpts{Pending: false}, _GLOBAL_EXIT_ROOT_SETTER_ROLE, auth.From)
	if !hasRole {
		err = errors.New("failed to set role")
	}
	if precalculatedAddr != gerAddr {
		err = errors.New("error calculating addr")
	}
	return
}
