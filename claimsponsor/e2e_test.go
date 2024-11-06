package claimsponsor_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"path"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/claimsponsor"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/test/aggoraclehelpers"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestE2EL1toEVML2(t *testing.T) {
	// start other needed components
	ctx := context.Background()
	env := aggoraclehelpers.SetupAggoracleWithEVMChain(t)
	dbPathBridgeSyncL1 := path.Join(t.TempDir(), "file::memory:?cache=shared")
	testClient := helpers.TestClient{ClientRenamed: env.L1Client.Client()}
	bridgeSyncL1, err := bridgesync.NewL1(ctx, dbPathBridgeSyncL1, env.BridgeL1Addr, 10, etherman.LatestBlock, env.ReorgDetector, testClient, 0, time.Millisecond*10, 0, 0, 1)
	require.NoError(t, err)
	go bridgeSyncL1.Start(ctx)

	// start claim sponsor
	dbPathClaimSponsor := path.Join(t.TempDir(), "file::memory:?cache=shared")
	claimer, err := claimsponsor.NewEVMClaimSponsor(
		log.GetDefaultLogger(),
		dbPathClaimSponsor,
		env.L2Client.Client(),
		env.BridgeL2Addr,
		env.AuthL2.From,
		200_000,
		0,
		env.EthTxManMockL2,
		0, 0, time.Millisecond*10, time.Millisecond*10,
	)
	require.NoError(t, err)
	go claimer.Start(ctx)

	// test
	for i := 0; i < 3; i++ {
		// Send bridges to L2, wait for GER to be injected on L2
		amount := big.NewInt(int64(i) + 1)
		env.AuthL1.Value = amount
		_, err := env.BridgeL1Contract.BridgeAsset(env.AuthL1, env.NetworkIDL2, env.AuthL2.From, amount, common.Address{}, true, nil)
		require.NoError(t, err)
		env.L1Client.Commit()
		time.Sleep(time.Millisecond * 300)
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
		require.NoError(t, err)

		// Request to sponsor claim
		globalIndex := bridgesync.GenerateGlobalIndex(true, 0, uint32(i))
		err = claimer.AddClaimToQueue(&claimsponsor.Claim{
			LeafType:            0,
			ProofLocalExitRoot:  localProof,
			ProofRollupExitRoot: rollupProof,
			GlobalIndex:         globalIndex,
			MainnetExitRoot:     info.MainnetExitRoot,
			RollupExitRoot:      info.RollupExitRoot,
			OriginNetwork:       0,
			OriginTokenAddress:  common.Address{},
			DestinationNetwork:  env.NetworkIDL2,
			DestinationAddress:  env.AuthL2.From,
			Amount:              amount,
			Metadata:            nil,
		})
		require.NoError(t, err)

		// Wait until success
		succeed := false
		for i := 0; i < 10; i++ {
			claim, err := claimer.GetClaim(globalIndex)
			require.NoError(t, err)
			if claim.Status == claimsponsor.FailedClaimStatus {
				require.NoError(t, errors.New("claim failed"))
			} else if claim.Status == claimsponsor.SuccessClaimStatus {
				succeed = true

				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		require.True(t, succeed)

		// Check on contract that is claimed
		isClaimed, err := env.BridgeL2Contract.IsClaimed(&bind.CallOpts{Pending: false}, uint32(i), 0)
		require.NoError(t, err)
		require.True(t, isClaimed)
	}
}
