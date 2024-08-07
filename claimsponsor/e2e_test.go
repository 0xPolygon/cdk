package claimsponsor_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/claimsponsor"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestE2EL1toEVML2(t *testing.T) {
	// start other needed components
	ctx := context.Background()
	env := helpers.SetupAggoracleWithEVMChain(t)
	dbPathBridgeSyncL1 := t.TempDir()
	bridgeSyncL1, err := bridgesync.NewL1(ctx, dbPathBridgeSyncL1, env.BridgeL1Addr, 10, etherman.LatestBlock, env.ReorgDetector, env.L1Client.Client(), 0)
	require.NoError(t, err)
	go bridgeSyncL1.Start(ctx)

	// start claim sponsor
	dbPathClaimSponsor := t.TempDir()
	ethTxMock := helpers.NewEthTxManagerMock(t)
	claimer, err := claimsponsor.NewEVMClaimSponsor(dbPathClaimSponsor, env.L2Client.Client(), env.BridgeL2Addr, env.AuthL2.From, 100_000, 0, ethTxMock)
	require.NoError(t, err)
	go claimer.Start(ctx)

	// test
	for i := 0; i < 10; i++ {
		// Send bridges to L2, wait for GER to be injected on L2
		amount := big.NewInt(int64(i) + 1)
		_, err := env.BridgeL1Contract.BridgeAsset(env.AuthL1, env.NetworkIDL2, env.AuthL2.From, amount, common.Address{}, true, nil)
		require.NoError(t, err)
		env.L1Client.Commit()
		time.Sleep(time.Millisecond * 50)
		expectedGER, err := env.GERL1Contract.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		isInjected, err := env.AggOracleSender.IsGERAlreadyInjected(expectedGER)
		require.NoError(t, err)
		require.True(t, isInjected, fmt.Sprintf("iteration %d, GER: %s", i, common.Bytes2Hex(expectedGER[:])))

		// Build MP using bridgeSyncL1 & env.L1InfoTreeSync
		info, err := env.L1InfoTreeSync.GetInfoByIndex(ctx, uint32(i))
		require.NoError(t, err)
		localProof, err := bridgeSyncL1.GetProof(ctx, uint32(i), info.MainnetExitRoot)
		require.NoError(t, err)
		rollupProof, err := env.L1InfoTreeSync.GetRollupExitTreeMerkleProof(ctx, 0, common.Hash{})

		// Request to sponsor claim
		claimer.AddClaimToQueue(ctx, &claimsponsor.Claim{
			LeafType:            0,
			ProofLocalExitRoot:  localProof,
			ProofRollupExitRoot: rollupProof,
			GlobalIndex:         nil, // TODO
			MainnetExitRoot:     info.MainnetExitRoot,
			RollupExitRoot:      info.RollupExitRoot,
			OriginNetwork:       0,
			OriginTokenAddress:  common.Address{},
			DestinationNetwork:  env.NetworkIDL2,
			DestinationAddress:  env.AuthL2.From,
			Amount:              amount,
			Metadata:            nil,
		})

		// TODO: Wait until success

		// Check on contract that is claimed
		isClaimed, err := env.BridgeL2Contract.IsClaimed(&bind.CallOpts{Pending: false}, uint32(i), 0)
		require.NoError(t, err)
		require.True(t, isClaimed)
	}
}
