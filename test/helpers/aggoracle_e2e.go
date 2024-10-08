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
	NetworkIDL2        = uint32(1)
	chainID            = 1337
	initialBalance     = "10000000000000000000000000"
	blockGasLimit      = uint64(999999999999999999)
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
	EthTxManMockL2   *EthTxManagerMock
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
	*TestClient,
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
	// Simulated L1
	privateKeyL1, err := crypto.GenerateKey()
	require.NoError(t, err)
	authL1, err := bind.NewKeyedTransactorWithChainID(privateKeyL1, big.NewInt(chainID))
	require.NoError(t, err)
	l1Client, gerL1Addr, gerL1Contract, bridgeL1Addr, bridgeL1Contract, err := newSimulatedL1(t)
	require.NoError(t, err)
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
	*EthTxManagerMock,
) {
	t.Helper()

	privateKeyL2, err := crypto.GenerateKey()
	require.NoError(t, err)
	authL2, err := bind.NewKeyedTransactorWithChainID(privateKeyL2, big.NewInt(chainID))
	require.NoError(t, err)
	l2Client, gerL2Addr, gerL2Sc, bridgeL2Addr, bridgeL2Sc, err := newSimulatedEVMAggSovereignChain(authL2)
	require.NoError(t, err)
	ethTxManMock := NewEthTxManMock(t, l2Client, authL2)
	sender, err := chaingersender.NewEVMChainGERSender(log.GetDefaultLogger(),
		gerL2Addr, authL2.From, l2Client.Client(), ethTxManMock, 0, time.Millisecond*50) //nolint:mnd
	require.NoError(t, err)

	return sender, l2Client, gerL2Sc, gerL2Addr, bridgeL2Sc, bridgeL2Addr, authL2, ethTxManMock
}

func newSimulatedL1(t *testing.T) (
	client *TestClient,
	gerAddr common.Address,
	gerContract *gerContractL1.Globalexitrootnopush0,
	bridgeAddr common.Address,
	bridgeContract *polygonzkevmbridgev2.Polygonzkevmbridgev2,
	err error,
) {
	t.Helper()

	ctx := context.Background()

	client = NewTestClient(t)
	auth := client.UserAuth()
	bridgeImplementationAddr, _ := client.Polygonzkevmbridgev2()

	nonce, err := client.SClient.PendingNonceAt(ctx, auth.From)
	require.NoError(t, err)

	precalculatedAddr := crypto.CreateAddress(auth.From, nonce+1)
	bridgeABI, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	require.NoError(t, err)

	if bridgeABI == nil {
		err = errors.New("GetABI returned nil")
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("failed to get bridge ABI: %w", err)
	}

	dataCallProxy, err := bridgeABI.Pack("initialize",
		uint32(0),        // networkIDMainnet
		common.Address{}, // gasTokenAddressMainnet"
		uint32(0),        // gasTokenNetworkMainnet
		precalculatedAddr,
		common.Address{},
		[]byte{}, // gasTokenMetadata
	)
	require.NoError(t, err)

	bridgeAddr, _, _, err = transparentupgradableproxy.DeployTransparentupgradableproxy(
		auth,
		client.SClient,
		bridgeImplementationAddr,
		auth.From,
		dataCallProxy,
	)
	require.NoError(t, err)
	client.Commit()

	bridgeContract, err = polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridgeAddr, client.SClient)
	require.NoError(t, err)

	checkGERAddr, err := bridgeContract.GlobalExitRootManager(&bind.CallOpts{Pending: false})
	require.NoError(t, err)
	if precalculatedAddr != checkGERAddr {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf(
			"error deploying bridge, unexpected GER addr. Expected %s. Actual %s",
			precalculatedAddr.Hex(), checkGERAddr.Hex(),
		)
	}

	gerAddr, _, gerContract, err = gerContractL1.DeployGlobalexitrootnopush0(client.DeployerAuth(), client.SClient, auth.From, bridgeAddr)
	require.NoError(t, err)
	client.Commit()

	if precalculatedAddr != gerAddr {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf(
			"error calculating GER address. Expected %s. Actual %s",
			precalculatedAddr.Hex(), gerAddr.Hex(),
		)
	}

	return client, gerAddr, gerContract, bridgeAddr, bridgeContract, nil
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
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	authDeployer, err := bind.NewKeyedTransactorWithChainID(privateKeyL1, big.NewInt(chainID))
	if err != nil {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	balance, _ := new(big.Int).SetString(initialBalance, 10) //nolint:mnd
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

	const blockGasLimit = uint64(999999999999999999)
	client = simulated.NewBackend(genesisAlloc, simulated.WithBlockGasLimit(blockGasLimit))

	bridgeImplementationAddr, _, _, err := polygonzkevmbridgev2.DeployPolygonzkevmbridgev2(authDeployer, client.Client())
	if err != nil {
		return nil, common.Address{}, nil,
			common.Address{}, nil,
			fmt.Errorf("failed to deploy bridge implementation: %w", err)
	}
	client.Commit()

	nonce, err := client.Client().PendingNonceAt(ctx, authDeployer.From)
	if err != nil {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("failed to get pending nonce: %w", err)
	}
	precalculatedAddr := crypto.CreateAddress(authDeployer.From, nonce+1)

	bridgeABI, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	if err != nil {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("failed to get bridge ABI: %w", err)
	}
	if bridgeABI == nil {
		err = errors.New("GetABI returned nil")
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("failed to get bridge ABI: %w", err)
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
		return nil, common.Address{}, nil,
			common.Address{}, nil,
			fmt.Errorf("failed to pack data for proxy initialization: %w", err)
	}

	bridgeAddr, _, _, err = transparentupgradableproxy.DeployTransparentupgradableproxy(
		authDeployer,
		client.Client(),
		bridgeImplementationAddr,
		authDeployer.From,
		dataCallProxy,
	)
	if err != nil {
		return nil, common.Address{}, nil,
			common.Address{}, nil,
			fmt.Errorf("failed to deploy transparent upgradable proxy: %w", err)
	}
	if bridgeAddr != precalculatedBridgeAddr {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf(
			"error calculating bridge addr. Expected: %s. Actual: %s",
			precalculatedBridgeAddr, bridgeAddr,
		)
	}
	client.Commit()

	bridgeContract, err = polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridgeAddr, client.Client())
	if err != nil {
		return nil, common.Address{}, nil,
			common.Address{}, nil,
			fmt.Errorf("failed to create bridge contract instance: %w", err)
	}

	checkGERAddr, err := bridgeContract.GlobalExitRootManager(&bind.CallOpts{})
	if err != nil {
		return nil, common.Address{}, nil,
			common.Address{}, nil,
			fmt.Errorf("failed to get Global Exit Root Manager: %w", err)
	}
	if precalculatedAddr != checkGERAddr {
		return nil, common.Address{}, nil, common.Address{}, nil, errors.New(
			"error deploying bridge, unexpected GER Manager address",
		)
	}

	gerAddr, _, gerContract, err = gerContractEVMChain.DeployPessimisticglobalexitrootnopush0(
		authDeployer, client.Client(), auth.From)
	if err != nil {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("failed to deploy GER contract: %w", err)
	}
	client.Commit()

	globalExitRootSetterRole := common.HexToHash("0x7b95520991dfda409891be0afa2635b63540f92ee996fda0bf695a166e5c5176")
	_, err = gerContract.GrantRole(authDeployer, globalExitRootSetterRole, auth.From)
	if err != nil {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("failed to grant role to GER contract: %w", err)
	}
	client.Commit()

	hasRole, _ := gerContract.HasRole(&bind.CallOpts{Pending: false}, globalExitRootSetterRole, auth.From)
	if !hasRole {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("failed to set role for GER contract")
	}

	if precalculatedAddr != gerAddr {
		return nil, common.Address{}, nil, common.Address{}, nil, fmt.Errorf("error calculating GER address")
	}

	return client, gerAddr, gerContract, bridgeAddr, bridgeContract, nil
}
