package rpc

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/claimsponsor"
	"github.com/0xPolygon/cdk/l1bridge2infoindexsync"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/lastgersync"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	// CDK is the namespace of the cdk service
	CDK       = "cdk"
	meterName = "github.com/0xPolygon/cdk/rpc"
)

// CDKEndpoints contains implementations for the "cdk" RPC endpoints
type CDKEndpoints struct {
	meter       metric.Meter
	readTimeout time.Duration
	networkID   uint32
	// TODO: Add syncers
	sponsor        *claimsponsor.ClaimSponsor
	l1InfoTree     *l1infotreesync.L1InfoTreeSync
	l1Bridge2Index *l1bridge2infoindexsync.L1Bridge2InfoIndexSync
	injectedGERs   *lastgersync.LastGERSync
	bridgeL1       *bridgesync.BridgeSync
	bridgeL2       *bridgesync.BridgeSync
}

// NewInteropEndpoints returns InteropEndpoints
func NewInteropEndpoints(
// TODO: Add syncers
) *CDKEndpoints {
	meter := otel.Meter(meterName)

	return &CDKEndpoints{
		// TODO: Add syncers
		meter: meter,
	}
}

// L1InfoTreeIndexForBridge returns the first L1 Info Tree index in which the bridge was included.
// This call needs to be done to a client of the same network were the bridge tx was sent
func (cdk *CDKEndpoints) L1InfoTreeIndexForBridge(networkID uint32, depositCount uint32) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), cdk.readTimeout)
	defer cancel()

	c, merr := cdk.meter.Int64Counter("l1_info_tree_index_for_bridge")
	if merr != nil {
		log.Warnf("failed to create l1_info_tree_index_for_bridge counter: %s", merr)
	}
	c.Add(ctx, 1)

	if networkID == 0 {
		l1InfoTreeIndex, err := cdk.l1Bridge2Index.GetL1InfoTreeIndexByDepositCount(ctx, depositCount)
		// TODO: special treatment of the error when not found,
		// as it's expected that it will take some time for the L1 Info tree to be updated
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get l1InfoTreeIndex, error: %s", err))
		}
		return l1InfoTreeIndex, nil
	}
	if networkID == cdk.networkID {
		// TODO: special treatment of the error when not found,
		// as it's expected that it will take some time for the L1 Info tree to be updated
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("TODO: batchsync / certificatesync missing"))
	}
	return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("this client does not support network %d", networkID))
}

// InjectedGlobalExitRootAfterIndex return the first GER injected onto the network that is linked
// to the given index or greater. This call is usefull to understand when a bridge is ready to be claimed
// on its destination network
func (cdk *CDKEndpoints) InjectedGlobalExitRootAfterIndex(networkID uint32, l1InfoTreeIndex uint32) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), cdk.readTimeout)
	defer cancel()

	c, merr := cdk.meter.Int64Counter("injected_ger_after_index")
	if merr != nil {
		log.Warnf("failed to create injected_ger_after_index counter: %s", merr)
	}
	c.Add(ctx, 1)

	if networkID == 0 {
		ger, err := cdk.l1InfoTree.GetInfoByIndex(ctx, l1InfoTreeIndex)
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get global exit root, error: %s", err))
		}
		return ger, nil
	}
	if networkID == cdk.networkID {
		ger, err := cdk.injectedGERs.GetFirstGERAfterL1InfoTreeIndex(ctx, l1InfoTreeIndex)
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get global exit root, error: %s", err))
		}
		return ger, nil
	}
	return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("this client does not support network %d", networkID))
}

type ClaimProof struct {
	ProofLocalExitRoot  [32]common.Hash
	ProofRollupExitRoot [32]common.Hash
}

// ClaimProof returns the proofs needed to claim a bridge. NetworkID and depositCount refere to the bridge origin
// while globalExitRoot should be already injected on the destination network.
// This call needs to be done to a client of the same network were the bridge tx was sent
func (cdk *CDKEndpoints) ClaimProof(networkID uint32, depositCount uint32, globalExitRoot common.Hash) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), cdk.readTimeout)
	defer cancel()

	c, merr := cdk.meter.Int64Counter("claim_proof")
	if merr != nil {
		log.Warnf("failed to create claim_proof counter: %s", merr)
	}
	c.Add(ctx, 1)

	proofRollupExitRoot, err := cdk.l1InfoTree.GetRollupExitTreeMerkleProof(ctx, networkID, globalExitRoot)
	if err != nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get rollup exit proof, error: %s", err))
	}
	var proofLocalExitRoot [32]common.Hash
	if networkID == 0 {
		proofLocalExitRoot, err := cdk.bridgeL1.GetProof(ctx, depositCount)
	} else if networkID == cdk.networkID {
		proofLocalExitRoot, err := cdk.bridgeL2.GetProof(ctx, depositCount)
	} else {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("this client does not support network %d", networkID))
	}
	return ProofClaim{
		ProofLocalExitRoot:  proofLocalExitRoot,
		ProofRollupExitRoot: proofRollupExitRoot,
	}, nil
}

// SponsorClaim sends a claim tx on behalf of the user.
// This call needs to be done to a client of the same network were the claim is going to be sent (bridge destination)
func (cdk *CDKEndpoints) SponsorClaim() (interface{}, rpc.Error) {
	return nil, nil
}

// GetSponsoredClaimStatus returns the status of a claim that has been previously requested to be sponsored.
// This call needs to be done to the same client were it was requested to be sponsored
func (cdk *CDKEndpoints) GetSponsoredClaimStatus(globalIndex *big.Int) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), cdk.readTimeout)
	defer cancel()

	c, merr := cdk.meter.Int64Counter("get_sponsored_claim_status")
	if merr != nil {
		log.Warnf("failed to create get_sponsored_claim_status counter: %s", merr)
	}
	c.Add(ctx, 1)

	claim, err := cdk.sponsor.GetClaim(ctx, globalIndex)
	if err != nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get claim status, error: %s", err))
	}
	return claim.Status, nil
}
