package bridgesync

import (
	"math/big"
	"os/exec"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func newDockerBackend(t *testing.T) {
	msg, err := exec.Command("bash", "-l", "-c", "docker compose up -d test-claimcalldata-l1").CombinedOutput()
	require.NoError(t, err, string(msg))
}

func TestClaimCalldata(t *testing.T) {
	msg, err := exec.Command("bash", "-l", "-c", "docker compose up -d").CombinedOutput()
	require.NoError(t, err, string(msg))
	defer func() {
		msg, err = exec.Command("bash", "-l", "-c", "docker compose down").CombinedOutput()
		require.NoError(t, err, string(msg))
	}()
	client, err := ethclient.Dial("http://localhost:8545")
	require.NoError(t, err)
	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(0).SetUint64(1337))
	require.NoError(t, err)

}
