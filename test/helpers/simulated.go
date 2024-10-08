package helpers

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana-paris/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry-paris/polygonzkevmbridgev2"
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
	ChainID = 1337
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
	gerContract *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2
}

func NewTestClient(t *testing.T) *TestClient {
	t.Helper()

	deployerPrivateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	deployerAuth, err := bind.NewKeyedTransactorWithChainID(deployerPrivateKey, big.NewInt(ChainID))
	require.NoError(t, err)

	deployerBalance, ok := big.NewInt(0).SetString("10000000000000000000000000", 10)
	require.True(t, ok)

	userPrivateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	userAuth, err := bind.NewKeyedTransactorWithChainID(userPrivateKey, big.NewInt(ChainID))
	require.NoError(t, err)

	userBalance, ok := big.NewInt(0).SetString("10000000000000000000000000", 10)
	require.True(t, ok)

	client := simulated.NewBackend(map[common.Address]types.Account{
		deployerAuth.From: {
			Balance: deployerBalance,
		},
		userAuth.From: {
			Balance: userBalance,
		},
	}, simulated.WithBlockGasLimit(999999999999999999))

	// Commit first block
	client.Commit()

	// Deploy bridge contract
	bridgeAddr, _, bridgeContract, err := polygonzkevmbridgev2.DeployPolygonzkevmbridgev2(deployerAuth, client.Client())
	require.NoError(t, err)
	client.Commit()

	// Deploy verify batches mock contract
	nonce, err := client.Client().PendingNonceAt(context.Background(), userAuth.From)
	require.NoError(t, err)
	precalculatedAddr := crypto.CreateAddress(userAuth.From, nonce+1)
	verifyAddr, _, verifyContract, err := verifybatchesmock.DeployVerifybatchesmock(deployerAuth, client.Client(), precalculatedAddr)
	require.NoError(t, err)
	client.Commit()

	// Deploy global exit root contract
	gerAddr, _, gerContract, err := polygonzkevmglobalexitrootv2.DeployPolygonzkevmglobalexitrootv2(deployerAuth, client.Client(), verifyAddr, userAuth.From)
	require.NoError(t, err)
	client.Commit()
	require.Equal(t, precalculatedAddr, gerAddr)

	return &TestClient{
		SClient:        client.Client(),
		Backend:        client,
		userAuth:       userAuth,
		deployerAuth:   deployerAuth,
		bridgeAddr:     bridgeAddr,
		bridgeContract: bridgeContract,
		verifyAddr:     verifyAddr,
		verifyContract: verifyContract,
		gerAddr:        gerAddr,
		gerContract:    gerContract,
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

func (tc TestClient) Polygonzkevmglobalexitrootv2() (common.Address, *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2) {
	return tc.gerAddr, tc.gerContract
}
