package helpers

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/l2-sovereign-chain-paris/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk/log"
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
	defaultBalance       = "100000000000000000000000000"
	chainID              = 1337

	base10 = 10
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
	UserAuth            *bind.TransactOpts
	DeployerAuth        *bind.TransactOpts
	BridgeProxyAddr     common.Address
	BridgeProxyContract *polygonzkevmbridgev2.Polygonzkevmbridgev2
}

// DeployBridge deploys the bridge contract
func (s *SimulatedBackendSetup) DeployBridge(client *simulated.Backend,
	gerAddr common.Address, networkID uint32) error {
	// Deploy zkevm bridge contract
	bridgeAddr, _, _, err := polygonzkevmbridgev2.DeployPolygonzkevmbridgev2(s.DeployerAuth, client.Client())
	if err != nil {
		return err
	}
	client.Commit()

	// Create proxy contract for the bridge
	var (
		bridgeProxyAddr     common.Address
		bridgeProxyContract *polygonzkevmbridgev2.Polygonzkevmbridgev2
	)

	bridgeABI, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	if err != nil {
		return err
	}

	dataCallProxy, err := bridgeABI.Pack("initialize",
		networkID,
		common.Address{}, // gasTokenAddressMainnet
		uint32(0),        // gasTokenNetworkMainnet
		gerAddr,          // global exit root manager
		common.Address{}, // rollup manager
		[]byte{},         // gasTokenMetadata
	)
	if err != nil {
		return err
	}

	bridgeProxyAddr, _, _, err = transparentupgradableproxy.DeployTransparentupgradableproxy(
		s.DeployerAuth,
		client.Client(),
		bridgeAddr,
		s.DeployerAuth.From,
		dataCallProxy,
	)
	if err != nil {
		return err
	}
	client.Commit()

	bridgeProxyContract, err = polygonzkevmbridgev2.NewPolygonzkevmbridgev2(bridgeProxyAddr, client.Client())
	if err != nil {
		return err
	}

	actualGERAddr, err := bridgeProxyContract.GlobalExitRootManager(&bind.CallOpts{})
	if err != nil {
		return err
	}

	if gerAddr != actualGERAddr {
		return fmt.Errorf("mismatch between expected %s and actual %s GER addresses on bridge contract (%s)",
			gerAddr, actualGERAddr, bridgeProxyAddr)
	}

	s.BridgeProxyAddr = bridgeProxyAddr
	s.BridgeProxyContract = bridgeProxyContract

	bridgeBalance, err := client.Client().BalanceAt(context.Background(), bridgeProxyAddr, nil)
	if err != nil {
		return err
	}

	log.Debugf("Bridge@%s, balance=%d\n", bridgeProxyAddr, bridgeBalance)

	return nil
}

// NewSimulatedBackend creates a simulated backend with two accounts: user and deployer.
func NewSimulatedBackend(t *testing.T,
	balances map[common.Address]types.Account,
	deployerAuth *bind.TransactOpts) (*simulated.Backend, *SimulatedBackendSetup) {
	t.Helper()

	// Define default balance
	balance, ok := new(big.Int).SetString(defaultBalance, 10) //nolint:mnd
	require.Truef(t, ok, "failed to set balance")

	// Create user account
	userPK, err := crypto.GenerateKey()
	require.NoError(t, err)
	userAuth, err := bind.NewKeyedTransactorWithChainID(userPK, big.NewInt(chainID))
	require.NoError(t, err)

	// Create deployer account
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

	setup := &SimulatedBackendSetup{
		UserAuth:     userAuth,
		DeployerAuth: deployerAuth,
	}

	return client, setup
}

func CreateAccount(chainID *big.Int) (*bind.TransactOpts, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	return bind.NewKeyedTransactorWithChainID(privateKey, chainID)
}
