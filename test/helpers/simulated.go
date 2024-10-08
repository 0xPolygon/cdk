package helpers

import (
	"math/big"
	"testing"

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

// SimulatedBackend creates a simulated backend with two accounts: user and deployer.
func SimulatedBackend(
	t *testing.T,
	balances map[common.Address]types.Account,
) (*simulated.Backend, *bind.TransactOpts, *bind.TransactOpts) {
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

	// Define balances map
	if balances == nil {
		balances = make(map[common.Address]types.Account)
	}
	balances[userAuth.From] = types.Account{Balance: balance}
	balances[deployerAuth.From] = types.Account{Balance: balance}

	client := simulated.NewBackend(balances, simulated.WithBlockGasLimit(defaultBlockGasLimit))

	return client, userAuth, deployerAuth
}
