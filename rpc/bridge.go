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
	// BRIDGE is the namespace of the bridge service
	BRIDGE    = "bridge"
	meterName = "github.com/0xPolygon/cdk/rpc"
)

// BridgeEndpoints contains implementations for the "bridge" RPC endpoints
type BridgeEndpoints struct {
	meter          metric.Meter
	readTimeout    time.Duration
	writeTimeout   time.Duration
	networkID      uint32
	sponsor        *claimsponsor.ClaimSponsor
	l1InfoTree     *l1infotreesync.L1InfoTreeSync
	l1Bridge2Index *l1bridge2infoindexsync.L1Bridge2InfoIndexSync
	injectedGERs   *lastgersync.LastGERSync
	bridgeL1       *bridgesync.BridgeSync
	bridgeL2       *bridgesync.BridgeSync
}

// NewBridgeEndpoints returns InteropEndpoints
func NewBridgeEndpoints(
	writeTimeout time.Duration,
	readTimeout time.Duration,
	networkID uint32,
	sponsor *claimsponsor.ClaimSponsor,
	l1InfoTree *l1infotreesync.L1InfoTreeSync,
	l1Bridge2Index *l1bridge2infoindexsync.L1Bridge2InfoIndexSync,
	injectedGERs *lastgersync.LastGERSync,
	bridgeL1 *bridgesync.BridgeSync,
	bridgeL2 *bridgesync.BridgeSync,
) *BridgeEndpoints {
	meter := otel.Meter(meterName)
	return &BridgeEndpoints{
		meter:          meter,
		readTimeout:    readTimeout,
		writeTimeout:   writeTimeout,
		networkID:      networkID,
		sponsor:        sponsor,
		l1InfoTree:     l1InfoTree,
		l1Bridge2Index: l1Bridge2Index,
		injectedGERs:   injectedGERs,
		bridgeL1:       bridgeL1,
		bridgeL2:       bridgeL2,
	}
}

// L1InfoTreeIndexForBridge returns the first L1 Info Tree index in which the bridge was included.
// This call needs to be done to a client of the same network were the bridge tx was sent
func (b *BridgeEndpoints) L1InfoTreeIndexForBridge(networkID uint32, depositCount uint32) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.readTimeout)
	defer cancel()

	c, merr := b.meter.Int64Counter("l1_info_tree_index_for_bridge")
	if merr != nil {
		log.Warnf("failed to create l1_info_tree_index_for_bridge counter: %s", merr)
	}
	c.Add(ctx, 1)

	if networkID == 0 {
		l1InfoTreeIndex, err := b.l1Bridge2Index.GetL1InfoTreeIndexByDepositCount(ctx, depositCount)
		// TODO: special treatment of the error when not found,
		// as it's expected that it will take some time for the L1 Info tree to be updated
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get l1InfoTreeIndex, error: %s", err))
		}
		return l1InfoTreeIndex, nil
	}
	if networkID == b.networkID {
		// TODO: special treatment of the error when not found,
		// as it's expected that it will take some time for the L1 Info tree to be updated
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("TODO: batchsync / certificatesync missing"))
	}
	return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("this client does not support network %d", networkID))
}

// InjectedInfoAfterIndex return the first GER injected onto the network that is linked
// to the given index or greater. This call is usefull to understand when a bridge is ready to be claimed
// on its destination network
func (b *BridgeEndpoints) InjectedInfoAfterIndex(networkID uint32, l1InfoTreeIndex uint32) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.readTimeout)
	defer cancel()

	c, merr := b.meter.Int64Counter("injected_info_after_index")
	if merr != nil {
		log.Warnf("failed to create injected_info_after_index counter: %s", merr)
	}
	c.Add(ctx, 1)

	if networkID == 0 {
		info, err := b.l1InfoTree.GetInfoByIndex(ctx, l1InfoTreeIndex)
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get global exit root, error: %s", err))
		}
		return info, nil
	}
	if networkID == b.networkID {
		injectedL1InfoTreeIndex, _, err := b.injectedGERs.GetFirstGERAfterL1InfoTreeIndex(ctx, l1InfoTreeIndex)
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get global exit root, error: %s", err))
		}
		info, err := b.l1InfoTree.GetInfoByIndex(ctx, injectedL1InfoTreeIndex)
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get global exit root, error: %s", err))
		}
		return info, nil
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
func (b *BridgeEndpoints) ClaimProof(networkID uint32, depositCount uint32, l1InfoTreeIndex uint32) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.readTimeout)
	defer cancel()

	c, merr := b.meter.Int64Counter("claim_proof")
	if merr != nil {
		log.Warnf("failed to create claim_proof counter: %s", merr)
	}
	c.Add(ctx, 1)

	info, err := b.l1InfoTree.GetInfoByIndex(ctx, l1InfoTreeIndex)
	if err != nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get info from the tree: %s", err))
	}
	proofRollupExitRoot, err := b.l1InfoTree.GetRollupExitTreeMerkleProof(ctx, networkID, info.GlobalExitRoot)
	if err != nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get rollup exit proof, error: %s", err))
	}
	var proofLocalExitRoot [32]common.Hash
	if networkID == 0 {
		proofLocalExitRoot, err = b.bridgeL1.GetProof(ctx, depositCount, info.MainnetExitRoot)
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get local exit proof, error: %s", err))
		}
	} else if networkID == b.networkID {
		localExitRoot, err := b.l1InfoTree.GetLocalExitRoot(ctx, networkID, info.RollupExitRoot)
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get local exit root from rollup exit tree, error: %s", err))
		}
		proofLocalExitRoot, err = b.bridgeL2.GetProof(ctx, depositCount, localExitRoot)
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get local exit proof, error: %s", err))
		}
	} else {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("this client does not support network %d", networkID))
	}
	return ClaimProof{
		ProofLocalExitRoot:  proofLocalExitRoot,
		ProofRollupExitRoot: proofRollupExitRoot,
	}, nil
}

// SponsorClaim sends a claim tx on behalf of the user.
// This call needs to be done to a client of the same network were the claim is going to be sent (bridge destination)
func (b *BridgeEndpoints) SponsorClaim(claim claimsponsor.Claim) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.writeTimeout)
	defer cancel()

	c, merr := b.meter.Int64Counter("sponsor_claim")
	if merr != nil {
		log.Warnf("failed to create sponsor_claim counter: %s", merr)
	}
	c.Add(ctx, 1)

	if b.sponsor == nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("this client does not support claim sponsoring"))
	}
	if claim.DestinationNetwork != b.networkID {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("this client only sponsors claims for network %d", b.networkID))
	}
	if err := b.sponsor.AddClaimToQueue(ctx, &claim); err != nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("error adding claim to the queue %s", err))
	}
	return nil, nil
}

// GetSponsoredClaimStatus returns the status of a claim that has been previously requested to be sponsored.
// This call needs to be done to the same client were it was requested to be sponsored
func (b *BridgeEndpoints) GetSponsoredClaimStatus(globalIndex *big.Int) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.readTimeout)
	defer cancel()

	c, merr := b.meter.Int64Counter("get_sponsored_claim_status")
	if merr != nil {
		log.Warnf("failed to create get_sponsored_claim_status counter: %s", merr)
	}
	c.Add(ctx, 1)

	if b.sponsor == nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("this client does not support claim sponsoring"))
	}
	claim, err := b.sponsor.GetClaim(ctx, globalIndex)
	if err != nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get claim status, error: %s", err))
	}
	return claim.Status, nil
}
