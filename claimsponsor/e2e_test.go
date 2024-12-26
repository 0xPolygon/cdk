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
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestE2EL1toEVML2(t *testing.T) {
	// start other needed components
	ctx := context.Background()
	l1Env, l2Env := helpers.NewL1EnvWithL2EVM(t)

	// start claim sponsor
	dbPathClaimSponsor := path.Join(t.TempDir(), "claimsponsorTestE2EL1toEVML2_cs.sqlite")
	claimer, err := claimsponsor.NewEVMClaimSponsor(
		log.GetDefaultLogger(),
		dbPathClaimSponsor,
		l2Env.SimBackend.Client(),
		l2Env.BridgeAddr,
		l2Env.Auth.From,
		200_000,
		0,
		l2Env.EthTxManagerMock,
		0, 0, time.Millisecond*10, time.Millisecond*10,
	)
	require.NoError(t, err)
	go claimer.Start(ctx)

	// test
	for i := uint32(0); i < 3; i++ {
		// Send bridges to L2, wait for GER to be injected on L2
		amount := new(big.Int).SetUint64(uint64(i) + 1)
		l1Env.Auth.Value = amount
		_, err := l1Env.BridgeContract.BridgeAsset(l1Env.Auth, l2Env.NetworkID, l2Env.Auth.From, amount, common.Address{}, true, nil)
		require.NoError(t, err)
		l1Env.SimBackend.Commit()
		time.Sleep(time.Millisecond * 300)

		expectedGER, err := l1Env.GERContract.GetLastGlobalExitRoot(&bind.CallOpts{Pending: false})
		require.NoError(t, err)
		_, err = l2Env.GERContract.InsertGlobalExitRoot(l2Env.Auth, expectedGER)
		require.NoError(t, err)
		l2Env.SimBackend.Commit()
		gerIndex, err := l2Env.GERContract.GlobalExitRootMap(nil, expectedGER)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(int64(i)+1), gerIndex, fmt.Sprintf("iteration %d, GER: %s is not updated on L2", i, common.Bytes2Hex(expectedGER[:])))

		// Build MP using bridgeSyncL1 & env.InfoTreeSync
		info, err := l1Env.InfoTreeSync.GetInfoByIndex(ctx, i)
		require.NoError(t, err)

		localProof, err := l1Env.BridgeSync.GetProof(ctx, i, info.MainnetExitRoot)
		require.NoError(t, err)

		rollupProof, err := l1Env.InfoTreeSync.GetRollupExitTreeMerkleProof(ctx, 0, common.Hash{})
		require.NoError(t, err)

		// Request to sponsor claim
		globalIndex := bridgesync.GenerateGlobalIndex(true, 0, i)
		err = claimer.AddClaimToQueue(&claimsponsor.Claim{
			LeafType:            claimsponsor.LeafTypeAsset,
			ProofLocalExitRoot:  localProof,
			ProofRollupExitRoot: rollupProof,
			GlobalIndex:         globalIndex,
			MainnetExitRoot:     info.MainnetExitRoot,
			RollupExitRoot:      info.RollupExitRoot,
			OriginNetwork:       0,
			OriginTokenAddress:  common.Address{},
			DestinationNetwork:  l2Env.NetworkID,
			DestinationAddress:  l2Env.Auth.From,
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
		isClaimed, err := l2Env.BridgeContract.IsClaimed(&bind.CallOpts{Pending: false}, i, 0)
		require.NoError(t, err)
		require.True(t, isClaimed)
	}
}
