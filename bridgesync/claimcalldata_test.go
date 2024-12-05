package bridgesync

import (
	"context"
	"math/big"
	"os/exec"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/test/contracts/claimmock"
	"github.com/0xPolygon/cdk/test/contracts/claimmockcaller"
	"github.com/0xPolygon/cdk/test/contracts/claimmocktest"
	tree "github.com/0xPolygon/cdk/tree/types"
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
	log.Debug("starting docker")
	ctx := context.Background()
	msg, err := exec.Command("bash", "-l", "-c", "docker compose up -d").CombinedOutput()
	require.NoError(t, err, string(msg))
	time.Sleep(time.Second * 1)
	defer func() {
		msg, err = exec.Command("bash", "-l", "-c", "docker compose down").CombinedOutput()
		require.NoError(t, err, string(msg))
	}()
	log.Debug("docker started")
	client, err := ethclient.Dial("http://localhost:8545")
	require.NoError(t, err)
	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(0).SetUint64(1337))
	require.NoError(t, err)

	// Deploy contracts
	bridgeAddr, _, bridgeContract, err := claimmock.DeployClaimmock(auth, client)
	require.NoError(t, err)
	claimCallerAddr, _, claimCaller, err := claimmockcaller.DeployClaimmockcaller(auth, client, bridgeAddr)
	require.NoError(t, err)
	_, _, claimTest, err := claimmocktest.DeployClaimmocktest(auth, client, bridgeAddr, claimCallerAddr)
	require.NoError(t, err)

	proofLocal := [32][32]byte{}
	proofLocalH := tree.Proof{}
	proofLocal[5] = common.HexToHash("beef")
	proofLocalH[5] = common.HexToHash("beef")
	proofRollup := [32][32]byte{}
	proofRollupH := tree.Proof{}
	proofRollup[4] = common.HexToHash("a1fa")
	proofRollupH[4] = common.HexToHash("a1fa")
	expectedClaim := Claim{
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
		GlobalExitRoot:      crypto.Keccak256Hash(common.HexToHash("5ca1e").Bytes(), common.HexToHash("dead").Bytes()),
	}
	expectedClaim2 := Claim{
		OriginNetwork:       87,
		OriginAddress:       common.HexToAddress("eebbeebb"),
		DestinationAddress:  common.HexToAddress("2233445566"),
		Amount:              big.NewInt(4),
		MainnetExitRoot:     common.HexToHash("5ca1e"),
		RollupExitRoot:      common.HexToHash("dead"),
		ProofLocalExitRoot:  proofLocalH,
		ProofRollupExitRoot: proofRollupH,
		DestinationNetwork:  0,
		Metadata:            []byte{},
		GlobalExitRoot:      crypto.Keccak256Hash(common.HexToHash("5ca1e").Bytes(), common.HexToHash("dead").Bytes()),
	}
	expectedClaim3 := Claim{
		OriginNetwork:       69,
		OriginAddress:       common.HexToAddress("ffaaffaa"),
		DestinationAddress:  common.HexToAddress("2233445566"),
		Amount:              big.NewInt(5),
		MainnetExitRoot:     common.HexToHash("5ca1e"),
		RollupExitRoot:      common.HexToHash("dead"),
		ProofLocalExitRoot:  proofLocalH,
		ProofRollupExitRoot: proofRollupH,
		DestinationNetwork:  0,
		Metadata:            []byte{},
		GlobalExitRoot:      crypto.Keccak256Hash(common.HexToHash("5ca1e").Bytes(), common.HexToHash("dead").Bytes()),
	}
	auth.GasLimit = 999999 // for some reason gas estimation fails :(

	abi, err := claimmock.ClaimmockMetaData.GetAbi()
	require.NoError(t, err)

	// direct call claim asset
	expectedClaim.GlobalIndex = big.NewInt(421)
	expectedClaim.IsMessage = false
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
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "direct call to claim asset",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// indirect call claim asset
	expectedClaim.IsMessage = false
	expectedClaim.GlobalIndex = big.NewInt(422)
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
		false,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "indirect call to claim asset",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// indirect call claim asset bytes
	expectedClaim.GlobalIndex = big.NewInt(423)
	expectedClaim.IsMessage = false
	expectedClaimBytes, err := abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.ClaimBytes(
		auth,
		expectedClaimBytes,
		false,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "indirect call to claim asset bytes",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// direct call claim message
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(424)
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
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "direct call to claim message",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// indirect call claim message
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(425)
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
		false,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "indirect call to claim message",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// indirect call claim message bytes
	expectedClaim.GlobalIndex = big.NewInt(426)
	expectedClaim.IsMessage = true
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.ClaimBytes(
		auth,
		expectedClaimBytes,
		false,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "indirect call to claim message bytes",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// indirect call claim message bytes
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim.IsMessage = true
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.ClaimBytes(
		auth,
		expectedClaimBytes,
		true,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	log.Infof("%+v", r.Logs)

	reverted := [2]bool{false, false}

	// 2 indirect call claim message (same global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(427)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err := abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim message 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim message 2 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim2,
	})

	// 2 indirect call claim message (diff global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(428)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(429)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim message 1 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim message 2 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim2,
	})

	reverted = [2]bool{false, true}

	// 2 indirect call claim message (same global index) (1 ok, 1 reverted)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(430)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(430)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim message (same globalIndex) (1 ok, 1 reverted)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// 2 indirect call claim message (diff global index) (1 ok, 1 reverted)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(431)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(432)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim message (diff globalIndex) (1 ok, 1 reverted)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	reverted = [2]bool{true, false}

	// 2 indirect call claim message (same global index) (1 reverted, 1 ok)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(433)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(433)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim message (same globalIndex) (reverted,ok)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim2,
	})

	// 2 indirect call claim message (diff global index) (1 reverted, 1 ok)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(434)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(435)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim message (diff globalIndex) (reverted,ok)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim2,
	})

	reverted = [2]bool{false, false}

	// 2 indirect call claim asset (same global index)
	expectedClaim.IsMessage = false
	expectedClaim.GlobalIndex = big.NewInt(436)
	expectedClaim2.IsMessage = false
	expectedClaim2.GlobalIndex = big.NewInt(436)
	expectedClaimBytes, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim asset 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim asset 2 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim2,
	})

	// 2 indirect call claim asset (diff global index)
	expectedClaim.IsMessage = false
	expectedClaim.GlobalIndex = big.NewInt(437)
	expectedClaim2.IsMessage = false
	expectedClaim2.GlobalIndex = big.NewInt(438)
	expectedClaimBytes, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim asset 1 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim asset 2 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim2,
	})

	reverted = [2]bool{false, true}

	// 2 indirect call claim asset (same global index) (1 ok, 1 reverted)
	expectedClaim.IsMessage = false
	expectedClaim.GlobalIndex = big.NewInt(439)
	expectedClaim2.IsMessage = false
	expectedClaim2.GlobalIndex = big.NewInt(439)
	expectedClaimBytes, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim asset (same globalIndex) (1 ok, 1 reverted)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	// 2 indirect call claim message (diff global index) (1 ok, 1 reverted)
	expectedClaim.IsMessage = false
	expectedClaim.GlobalIndex = big.NewInt(440)
	expectedClaim2.IsMessage = false
	expectedClaim2.GlobalIndex = big.NewInt(441)
	expectedClaimBytes, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim asset (diff globalIndex) (1 ok, 1 reverted)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	reverted = [2]bool{true, false}

	// 2 indirect call claim asset (same global index) (1 reverted, 1 ok)
	expectedClaim.IsMessage = false
	expectedClaim.GlobalIndex = big.NewInt(442)
	expectedClaim2.IsMessage = false
	expectedClaim2.GlobalIndex = big.NewInt(442)
	expectedClaimBytes, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim asset (same globalIndex) (reverted,ok)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim2,
	})

	// 2 indirect call claim asset (diff global index) (1 reverted, 1 ok)
	expectedClaim.IsMessage = false
	expectedClaim.GlobalIndex = big.NewInt(443)
	expectedClaim2.IsMessage = false
	expectedClaim2.GlobalIndex = big.NewInt(444)
	expectedClaimBytes, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimAsset",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimCaller.Claim2Bytes(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect call claim asset (diff globalIndex) (reverted,ok)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim2,
	})

	// indirect + indirect call claim message bytes
	expectedClaim.GlobalIndex = big.NewInt(426)
	expectedClaim.IsMessage = true
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.ClaimTestInternal(
		auth,
		expectedClaimBytes,
		false,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "indirect + indirect call to claim message bytes",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	reverted = [2]bool{false, false}

	// 2 indirect + indirect call claim message (same global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(427)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim2TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		reverted,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 indirect + indirect call claim message 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "2 indirect + indirect call claim message 2 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim2,
	})

	reverted3 := [3]bool{false, false, false}

	// 3 ok (indirectx2, indirect, indirectx2) call claim message (same global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(427)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(427)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err := abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "3 ok (indirectx2, indirect, indirectx2) call claim message 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "3 ok (indirectx2, indirect, indirectx2) call claim message 2 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim2,
	})
	testCases = append(testCases, testCase{
		description:   "3 ok (indirectx2, indirect, indirectx2) call claim message 3 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[2],
		expectedClaim: expectedClaim3,
	})

	// 3 ok (indirectx2, indirect, indirectx2) call claim message (diff global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(428)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(429)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "3 ok (indirectx2, indirect, indirectx2) call claim message 1 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "3 ok (indirectx2, indirect, indirectx2) call claim message 2 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim2,
	})
	testCases = append(testCases, testCase{
		description:   "3 ok (indirectx2, indirect, indirectx2) call claim message 3 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[2],
		expectedClaim: expectedClaim3,
	})

	reverted3 = [3]bool{true, false, false}

	// 1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message (diff global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(428)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(429)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message 1 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim2,
	})
	testCases = append(testCases, testCase{
		description:   "1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message 2 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim3,
	})

	reverted3 = [3]bool{false, true, false}

	// 1 ok 1 ko 1 ok (indirectx2, indirect, indirectx2) call claim message (diff global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(428)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(429)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message 1 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message 2 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim3,
	})

	reverted3 = [3]bool{false, false, true}

	// 1 ok 1 ok 1 ko (indirectx2, indirect, indirectx2) call claim message (diff global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(428)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(429)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "1 ok 1 ok 1 ko (indirectx2, indirect, indirectx2) call claim message 1 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "1 ok 1 ok 1 ko (indirectx2, indirect, indirectx2) call claim message 2 (diff globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim2,
	})

	reverted3 = [3]bool{true, false, false}

	// 1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message (same global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(427)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(427)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim2,
	})
	testCases = append(testCases, testCase{
		description:   "1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message 2 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim3,
	})

	reverted3 = [3]bool{false, true, false}

	// 1 ok 1 ko 1 ok (indirectx2, indirect, indirectx2) call claim message (same global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(427)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(427)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "1 ko 2 ok (indirectx2, indirect, indirectx2) call claim message 2 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim3,
	})

	reverted3 = [3]bool{false, false, true}

	// 1 ok 1 ok 1 ko (indirectx2, indirect, indirectx2) call claim message (same global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(427)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(427)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "1 ok 1 ok 1 ko (indirectx2, indirect, indirectx2) call claim message 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})
	testCases = append(testCases, testCase{
		description:   "1 ok 1 ok 1 ko (indirectx2, indirect, indirectx2) call claim message 2 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[1],
		expectedClaim: expectedClaim2,
	})

	reverted3 = [3]bool{true, true, false}

	// 2 ko 1 ok (indirectx2, indirect, indirectx2) call claim message (same global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(427)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(427)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "2 ko 1 ok (indirectx2, indirect, indirectx2) call claim message 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim3,
	})

	reverted3 = [3]bool{false, true, true}

	// 1 ok 2 ko (indirectx2, indirect, indirectx2) call claim message (same global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(427)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(427)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "1 ok 2 ko (indirectx2, indirect, indirectx2) call claim message 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim,
	})

	reverted3 = [3]bool{true, false, true}

	// 1 ko 1 ok 1 ko (indirectx2, indirect, indirectx2) call claim message (same global index)
	expectedClaim.IsMessage = true
	expectedClaim.GlobalIndex = big.NewInt(427)
	expectedClaim2.IsMessage = true
	expectedClaim2.GlobalIndex = big.NewInt(427)
	expectedClaim3.IsMessage = true
	expectedClaim3.GlobalIndex = big.NewInt(427)
	expectedClaimBytes, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim.GlobalIndex,
		expectedClaim.MainnetExitRoot,
		expectedClaim.RollupExitRoot,
		expectedClaim.OriginNetwork,
		expectedClaim.OriginAddress,
		expectedClaim.DestinationNetwork,
		expectedClaim.DestinationAddress,
		expectedClaim.Amount,
		expectedClaim.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes2, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim2.GlobalIndex,
		expectedClaim2.MainnetExitRoot,
		expectedClaim2.RollupExitRoot,
		expectedClaim2.OriginNetwork,
		expectedClaim2.OriginAddress,
		expectedClaim2.DestinationNetwork,
		expectedClaim2.DestinationAddress,
		expectedClaim2.Amount,
		expectedClaim2.Metadata,
	)
	require.NoError(t, err)
	expectedClaimBytes3, err = abi.Pack(
		"claimMessage",
		proofLocal,
		proofRollup,
		expectedClaim3.GlobalIndex,
		expectedClaim3.MainnetExitRoot,
		expectedClaim3.RollupExitRoot,
		expectedClaim3.OriginNetwork,
		expectedClaim3.OriginAddress,
		expectedClaim3.DestinationNetwork,
		expectedClaim3.DestinationAddress,
		expectedClaim3.Amount,
		expectedClaim3.Metadata,
	)
	require.NoError(t, err)
	tx, err = claimTest.Claim3TestInternal(
		auth,
		expectedClaimBytes,
		expectedClaimBytes2,
		expectedClaimBytes3,
		reverted3,
	)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)
	r, err = client.TransactionReceipt(ctx, tx.Hash())
	require.NoError(t, err)
	testCases = append(testCases, testCase{
		description:   "1 ko 1 ok 1 ko (indirectx2, indirect, indirectx2) call claim message 1 (same globalIndex)",
		bridgeAddr:    bridgeAddr,
		log:           *r.Logs[0],
		expectedClaim: expectedClaim2,
	})

	for _, tc := range testCases {
		log.Info(tc.description)
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
