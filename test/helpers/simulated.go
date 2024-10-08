package helpers

import (
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry-paris/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk/test/contracts/transparentupgradableproxy"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

const (
	defaultBlockGasLimit = uint64(999999999999999999)
	defaultBalance       = "10000000000000000000000000"
	chainID              = 1337
)

type ClientRenamed simulated.Client

type TestClient struct {
	ClientRenamed
}

func (tc TestClient) Client() *rpc.Client {
	return nil
}

// SimulatedBackendSetup defines the setup for a simulated backend.
type SimulatedBackendSetup struct {
	UserAuth                   *bind.TransactOpts
	DeployerAuth               *bind.TransactOpts
	EBZkevmBridgeAddr          common.Address
	EBZkevmBridgeContract      *polygonzkevmbridgev2.Polygonzkevmbridgev2
	EBZkevmBridgeProxyAddr     common.Address
	EBZkevmBridgeProxyContract *polygonzkevmbridgev2.Polygonzkevmbridgev2
}

// SimulatedBackend creates a simulated backend with two accounts: user and deployer.
func SimulatedBackend(
	t *testing.T,
	balances map[common.Address]types.Account,
	ebZkevmBridgeNetwork uint32,
) (*simulated.Backend, *SimulatedBackendSetup) {
	t.Helper()

	// Define default balance
	balance, ok := new(big.Int).SetString(defaultBalance, 10) //nolint:mnd
	require.Truef(t, ok, "failed to set balance")

	// Create user
	userPK, err := crypto.GenerateKey()
	require.NoError(t, err)
	userAuth, err := bind.NewKeyedTransactorWithChainID(userPK, big.NewInt(chainID))
	require.NoError(t, err)

	// Create deployer
	deployerPK, err := crypto.GenerateKey()
	require.NoError(t, err)
	deployerAuth, err := bind.NewKeyedTransactorWithChainID(deployerPK, big.NewInt(chainID))
	require.NoError(t, err)
	precalculatedBridgeAddr := crypto.CreateAddress(deployerAuth.From, 1)

	// Define balances map
	if balances == nil {
		balances = make(map[common.Address]types.Account)
	}
	balances[userAuth.From] = types.Account{Balance: balance}
	balances[deployerAuth.From] = types.Account{Balance: balance}
	balances[precalculatedBridgeAddr] = types.Account{Balance: balance}

	client := simulated.NewBackend(balances, simulated.WithBlockGasLimit(defaultBlockGasLimit))

	// Mine the first block
	client.Commit()

	// MUST BE DEPLOYED FIRST
	// Deploy zkevm bridge contract
	ebZkevmBridgeAddr, _, ebZkevmBridgeContract, err := polygonzkevmbridgev2.DeployPolygonzkevmbridgev2(deployerAuth, client.Client())
	require.NoError(t, err)
	client.Commit()

	// Create proxy contract for the bridge
	var ebZkevmBridgeProxyAddr common.Address
	var ebZkevmBridgeProxyContract *polygonzkevmbridgev2.Polygonzkevmbridgev2
	{
		precalculatedAddr := crypto.CreateAddress(deployerAuth.From, 2) //nolint:mnd

		bridgeABI, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
		require.NoError(t, err)
		require.NotNil(t, bridgeABI)

		dataCallProxy, err := bridgeABI.Pack("initialize",
			ebZkevmBridgeNetwork,
			common.Address{}, // gasTokenAddressMainnet
			uint32(0),        // gasTokenNetworkMainnet
			precalculatedAddr,
			common.Address{},
			[]byte{}, // gasTokenMetadata
		)
		require.NoError(t, err)

		ebZkevmBridgeProxyAddr, _, _, err = transparentupgradableproxy.DeployTransparentupgradableproxy(
			deployerAuth,
			client.Client(),
			ebZkevmBridgeAddr,
			deployerAuth.From,
			dataCallProxy,
		)
		require.NoError(t, err)
		require.Equal(t, precalculatedBridgeAddr, ebZkevmBridgeProxyAddr)
		client.Commit()

		ebZkevmBridgeProxyContract, err = polygonzkevmbridgev2.NewPolygonzkevmbridgev2(ebZkevmBridgeProxyAddr, client.Client())
		require.NoError(t, err)

		checkGERAddr, err := ebZkevmBridgeProxyContract.GlobalExitRootManager(&bind.CallOpts{})
		require.NoError(t, err)
		require.Equal(t, precalculatedAddr, checkGERAddr)
	}

	return client, &SimulatedBackendSetup{
		UserAuth:                   userAuth,
		DeployerAuth:               deployerAuth,
		EBZkevmBridgeAddr:          ebZkevmBridgeAddr,
		EBZkevmBridgeContract:      ebZkevmBridgeContract,
		EBZkevmBridgeProxyAddr:     ebZkevmBridgeProxyAddr,
		EBZkevmBridgeProxyContract: ebZkevmBridgeProxyContract,
	}
}
