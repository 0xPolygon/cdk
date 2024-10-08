package helpers

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana-paris/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry-paris/polygonzkevmbridgev2"
	gerContractL1 "github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/globalexitrootnopush0"
	gerContractEVMChain "github.com/0xPolygon/cdk-contracts-tooling/contracts/manual/pessimisticglobalexitrootnopush0"
	"github.com/0xPolygon/cdk/test/contracts/transparentupgradableproxy"
	"github.com/0xPolygon/cdk/test/contracts/verifybatchesmock"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

const (
	ChainID        = 1337
	defaultBalance = "10000000000000000000000000"
)

type SClient simulated.Client

type TestClient struct {
	SClient
	Backend      *simulated.Backend
	userAuth     *bind.TransactOpts
	deployerAuth *bind.TransactOpts

	bridgeAddr     common.Address
	bridgeContract *polygonzkevmbridgev2.Polygonzkevmbridgev2

	verifyAddr     common.Address
	verifyContract *verifybatchesmock.Verifybatchesmock

	gerAddr     common.Address
	gerContract *gerContractL1.Globalexitrootnopush0

	zkevmGerAddr     common.Address
	zkevmGerContract *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2

	peGerAddr     common.Address
	peGerContract *gerContractEVMChain.Pessimisticglobalexitrootnopush0
}

func NewTestClient(t *testing.T, networkID uint32) *TestClient {
	t.Helper()

	deployerPrivateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	deployerAuth, err := bind.NewKeyedTransactorWithChainID(deployerPrivateKey, big.NewInt(ChainID))
	require.NoError(t, err)

	deployerBalance, ok := big.NewInt(0).SetString(defaultBalance, 10) //nolint:mnd
	require.True(t, ok)

	userPrivateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	userAuth, err := bind.NewKeyedTransactorWithChainID(userPrivateKey, big.NewInt(ChainID))
	require.NoError(t, err)

	userBalance, ok := big.NewInt(0).SetString(defaultBalance, 10) //nolint:mnd
	require.True(t, ok)

	bridgeBalance, ok := big.NewInt(0).SetString(defaultBalance, 10) //nolint:mnd
	require.True(t, ok)

	precalculatedBridgeAddr := crypto.CreateAddress(deployerAuth.From, 1)
	client := simulated.NewBackend(map[common.Address]types.Account{
		deployerAuth.From: {
			Balance: deployerBalance,
		},
		userAuth.From: {
			Balance: userBalance,
		},
		precalculatedBridgeAddr: {
			Balance: bridgeBalance,
		},
	}, simulated.WithBlockGasLimit(999999999999999999)) //nolint:mnd

	// Commit first block
	client.Commit()

	// Deploy bridge contract
	bridgeAddr, _, bridgeContract, err := polygonzkevmbridgev2.DeployPolygonzkevmbridgev2(deployerAuth, client.Client()) //nolint:wastedassign
	require.NoError(t, err)
	client.Commit()

	// Deploy proxy for the bridge contract
	{
		nonce, err := client.Client().PendingNonceAt(context.Background(), deployerAuth.From)
		require.NoError(t, err)
		precalculatedAddr := crypto.CreateAddress(deployerAuth.From, nonce+1)

		bridgeABI, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
		require.NoError(t, err)
		require.NotNil(t, bridgeABI)

		dataCallProxy, err := bridgeABI.Pack("initialize",
			networkID,
			common.Address{}, // gasTokenAddressMainnet
			uint32(0),        // gasTokenNetworkMainnet
			precalculatedAddr,
			common.Address{},
			[]byte{}, // gasTokenMetadata
		)
		require.NoError(t, err)

		bridgeAddr, _, _, err = transparentupgradableproxy.DeployTransparentupgradableproxy(
			deployerAuth,
			client.Client(),
			bridgeAddr,
			deployerAuth.From,
			dataCallProxy,
		)
		require.NoError(t, err)
		require.Equal(t, precalculatedBridgeAddr, bridgeAddr)
		client.Commit()

		bridgeContract, err = polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridgeAddr, client.Client())
		require.NoError(t, err)

		checkGERAddr, err := bridgeContract.GlobalExitRootManager(&bind.CallOpts{})
		require.NoError(t, err)
		require.Equal(t, precalculatedAddr, checkGERAddr)
	}

	// Deploy verify batches mock contract
	nonce, err := client.Client().PendingNonceAt(context.Background(), deployerAuth.From)
	require.NoError(t, err)
	precalculatedAddr := crypto.CreateAddress(deployerAuth.From, nonce+1)
	verifyAddr, _, verifyContract, err := verifybatchesmock.DeployVerifybatchesmock(deployerAuth, client.Client(), precalculatedAddr)
	require.NoError(t, err)
	client.Commit()

	// Deploy zkevm global exit root contract
	zkevmGerAddr, _, zkevmGerContract, err := polygonzkevmglobalexitrootv2.DeployPolygonzkevmglobalexitrootv2(deployerAuth, client.Client(), verifyAddr, userAuth.From)
	require.NoError(t, err)
	client.Commit()
	require.Equal(t, precalculatedAddr, zkevmGerAddr)

	// Deploy pessimistic global exit root contract
	nonce, err = client.Client().PendingNonceAt(context.Background(), deployerAuth.From)
	require.NoError(t, err)
	precalculatedAddr = crypto.CreateAddress(deployerAuth.From, nonce+1)
	peGerAddr, _, peGerContract, err := gerContractEVMChain.DeployPessimisticglobalexitrootnopush0(
		deployerAuth, client.Client(), userAuth.From)
	require.NoError(t, err)
	client.Commit()
	//	require.Equal(t, precalculatedAddr, peGerAddr)

	// Grant role
	{
		globalExitRootSetterRole := common.HexToHash("0x7b95520991dfda409891be0afa2635b63540f92ee996fda0bf695a166e5c5176")
		_, err = peGerContract.GrantRole(deployerAuth, globalExitRootSetterRole, userAuth.From)
		require.NoError(t, err)
		client.Commit()

		hasRole, _ := peGerContract.HasRole(&bind.CallOpts{Pending: false}, globalExitRootSetterRole, userAuth.From)
		require.True(t, hasRole)
	}

	// Deploy global exit root contract
	gerAddr, _, gerContract, err := gerContractL1.DeployGlobalexitrootnopush0(deployerAuth, client.Client(), userAuth.From, bridgeAddr)
	require.NoError(t, err)
	client.Commit()

	return &TestClient{
		SClient:          client.Client(),
		Backend:          client,
		userAuth:         userAuth,
		deployerAuth:     deployerAuth,
		bridgeAddr:       bridgeAddr,
		bridgeContract:   bridgeContract,
		verifyAddr:       verifyAddr,
		verifyContract:   verifyContract,
		gerAddr:          gerAddr,
		gerContract:      gerContract,
		zkevmGerAddr:     zkevmGerAddr,
		zkevmGerContract: zkevmGerContract,
		peGerAddr:        peGerAddr,
		peGerContract:    peGerContract,
	}
}

func (tc TestClient) Client() *rpc.Client {
	return nil
}

func (tc TestClient) Commit() common.Hash {
	return tc.Backend.Commit()
}

func (tc TestClient) UserAuth() *bind.TransactOpts {
	return tc.userAuth
}

func (tc TestClient) DeployerAuth() *bind.TransactOpts {
	return tc.deployerAuth
}

func (tc TestClient) Polygonzkevmbridgev2() (common.Address, *polygonzkevmbridgev2.Polygonzkevmbridgev2) {
	return tc.bridgeAddr, tc.bridgeContract
}

func (tc TestClient) Verifybatchesmock() (common.Address, *verifybatchesmock.Verifybatchesmock) {
	return tc.verifyAddr, tc.verifyContract
}

func (tc TestClient) Globalexitrootnopush0() (common.Address, *gerContractL1.Globalexitrootnopush0) {
	return tc.gerAddr, tc.gerContract
}

func (tc TestClient) Polygonzkevmglobalexitrootv2() (common.Address, *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2) {
	return tc.zkevmGerAddr, tc.zkevmGerContract
}

func (tc TestClient) Pessimisticglobalexitrootnopush0() (common.Address, *gerContractEVMChain.Pessimisticglobalexitrootnopush0) {
	return tc.peGerAddr, tc.peGerContract
}
