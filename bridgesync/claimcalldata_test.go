package bridgesync

import (
	"context"
	"math/big"
	"os/exec"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/test/contracts/claimmock"
	"github.com/0xPolygon/cdk/test/contracts/claimmockcaller"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	description   string
	bridgeAddr    common.Address
	log           types.Log
	expectedClaim Claim
}

func TestClaimCalldata(t *testing.T) {
	testCases := []testCase{}
	// Setup Docker L1
	ctx := context.Background()
	msg, err := exec.Command("bash", "-l", "-c", "docker compose up -d").CombinedOutput()
	require.NoError(t, err, string(msg))
	time.Sleep(time.Second * 1)
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

	// Deploy contracts
	bridgeAddr, _, bridgeContract, err := claimmock.DeployClaimmock(auth, client)
	require.NoError(t, err)
	_, _, claimCaller, err := claimmockcaller.DeployClaimmockcaller(auth, client, bridgeAddr)
	require.NoError(t, err)

	proofLocal := [32][32]byte{}
	proofLocalH := [32]common.Hash{}
	proofLocal[5] = common.HexToHash("beef")
	proofLocalH[5] = common.HexToHash("beef")
	proofRollup := [32][32]byte{}
	proofRollupH := [32]common.Hash{}
	proofRollup[4] = common.HexToHash("a1fa")
	proofRollupH[4] = common.HexToHash("a1fa")
	expectedClaim := Claim{
		GlobalIndex:         big.NewInt(420),
		OriginNetwork:       69,
		OriginAddress:       common.HexToAddress("ffaaffaa"),
		DestinationAddress:  common.HexToAddress("123456789"),
		Amount:              big.NewInt(3),
		MainnetExitRoot:     common.HexToHash("5ca1e"),
		RollupExitRoot:      common.HexToHash("dead"),
		ProofLocalExitRoot:  proofLocalH,
		ProofRollupExitRoot: proofRollupH,
		DestinationNetwork:  0,
		Metadata:            []byte{},
	}
	auth.GasLimit = 999999 // for some reason gas estimation fails :(

	// direct call claim asset
	tx, err := bridgeContract.ClaimAsset(
		auth,
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		0,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		nil,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err := client.TransactionReceipt(ctx, tx.Hash())
	expectedClaim.IsMessage = false
	testCases = append(testCases, testCase{
		description:   "direct call to claim asset",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// indirect call claim asset
	tx, err = claimCaller.ClaimAsset(
		auth,
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		0,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		nil,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	expectedClaim.IsMessage = false
	testCases = append(testCases, testCase{
		description:   "indirect call to claim asset",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// direct call claim message
	tx, err = bridgeContract.ClaimMessage(
		auth,
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		0,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		nil,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	expectedClaim.IsMessage = true
	testCases = append(testCases, testCase{
		description:   "direct call to claim message",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// indirect call claim message
	tx, err = claimCaller.ClaimMessage(
		auth,
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		0,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		nil,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	expectedClaim.IsMessage = true
	testCases = append(testCases, testCase{
		description:   "indirect call to claim message",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			claimEvent, err := bridgeContract.ParseClaimEvent(tc.log)
			require.NoError(t, err)
			actualClaim := Claim{
				GlobalIndex:        claimEvent.GlobalIndex,
				OriginNetwork:      claimEvent.OriginNetwork,
				OriginAddress:      claimEvent.OriginAddress,
				DestinationAddress: claimEvent.DestinationAddress,
				Amount:             claimEvent.Amount,
			}
			err = setClaimCalldata(client, tc.bridgeAddr, tc.log.TxHash, &actualClaim)
			require.NoError(t, err)
			require.Equal(t, tc.expectedClaim, actualClaim)
		})
	}
}