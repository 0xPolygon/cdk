package rpc

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/claimsponsor"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/rpc/types"
	tree "github.com/0xPolygon/cdk/tree/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	// BRIDGE is the namespace of the bridge service
	BRIDGE    = "bridge"
	meterName = "github.com/0xPolygon/cdk/rpc"

	zeroHex              = "0x0"
	binnarySearchDivider = 2
)

var (
	ErrNotOnL1Info = errors.New("this bridge has not been included on the L1 Info Tree yet")
)

// BridgeEndpoints contains implementations for the "bridge" RPC endpoints
type BridgeEndpoints struct {
	logger       *log.Logger
	meter        metric.Meter
	readTimeout  time.Duration
	writeTimeout time.Duration
	networkID    uint32
	sponsor      ClaimSponsorer
	l1InfoTree   L1InfoTreer
	injectedGERs LastGERer
	bridgeL1     Bridger
	bridgeL2     Bridger
}

// NewBridgeEndpoints returns InteropEndpoints
func NewBridgeEndpoints(
	logger *log.Logger,
	writeTimeout time.Duration,
	readTimeout time.Duration,
	networkID uint32,
	sponsor ClaimSponsorer,
	l1InfoTree L1InfoTreer,
	injectedGERs LastGERer,
	bridgeL1 Bridger,
	bridgeL2 Bridger,
) *BridgeEndpoints {
	meter := otel.Meter(meterName)
	return &BridgeEndpoints{
		logger:       logger,
		meter:        meter,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		networkID:    networkID,
		sponsor:      sponsor,
		l1InfoTree:   l1InfoTree,
		injectedGERs: injectedGERs,
		bridgeL1:     bridgeL1,
		bridgeL2:     bridgeL2,
	}
}

// L1InfoTreeIndexForBridge returns the first L1 Info Tree index in which the bridge was included.
// networkID represents the origin network.
// This call needs to be done to a client of the same network were the bridge tx was sent
func (b *BridgeEndpoints) L1InfoTreeIndexForBridge(networkID uint32, depositCount uint32) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.readTimeout)
	defer cancel()

	c, merr := b.meter.Int64Counter("l1_info_tree_index_for_bridge")
	if merr != nil {
		b.logger.Warnf("failed to create l1_info_tree_index_for_bridge counter: %s", merr)
	}
	c.Add(ctx, 1)

	if networkID == 0 {
		l1InfoTreeIndex, err := b.getFirstL1InfoTreeIndexForL1Bridge(ctx, depositCount)
		// TODO: special treatment of the error when not found,
		// as it's expected that it will take some time for the L1 Info tree to be updated
		if err != nil {
			return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf(
				"failed to get l1InfoTreeIndex for networkID %d and deposit count %d, error: %s", networkID, depositCount, err),
			)
		}
		return l1InfoTreeIndex, nil
	}
	if networkID == b.networkID {
		l1InfoTreeIndex, err := b.getFirstL1InfoTreeIndexForL2Bridge(ctx, depositCount)
		// TODO: special treatment of the error when not found,
		// as it's expected that it will take some time for the L1 Info tree to be updated
		if err != nil {
			return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf(
				"failed to get l1InfoTreeIndex for networkID %d and deposit count %d, error: %s", networkID, depositCount, err),
			)
		}
		return l1InfoTreeIndex, nil
	}
	return zeroHex, rpc.NewRPCError(
		rpc.DefaultErrorCode,
		fmt.Sprintf("this client does not support network %d", networkID),
	)
}

// InjectedInfoAfterIndex return the first GER injected onto the network that is linked
// to the given index or greater. This call is useful to understand when a bridge is ready to be claimed
// on its destination network
func (b *BridgeEndpoints) InjectedInfoAfterIndex(networkID uint32, l1InfoTreeIndex uint32) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.readTimeout)
	defer cancel()

	c, merr := b.meter.Int64Counter("injected_info_after_index")
	if merr != nil {
		b.logger.Warnf("failed to create injected_info_after_index counter: %s", merr)
	}
	c.Add(ctx, 1)

	if networkID == 0 {
		info, err := b.l1InfoTree.GetInfoByIndex(ctx, l1InfoTreeIndex)
		if err != nil {
			return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get global exit root, error: %s", err))
		}
		return info, nil
	}
	if networkID == b.networkID {
		e, err := b.injectedGERs.GetFirstGERAfterL1InfoTreeIndex(ctx, l1InfoTreeIndex)
		if err != nil {
			return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get global exit root, error: %s", err))
		}
		info, err := b.l1InfoTree.GetInfoByIndex(ctx, e.L1InfoTreeIndex)
		if err != nil {
			return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get global exit root, error: %s", err))
		}
		return info, nil
	}
	return zeroHex, rpc.NewRPCError(
		rpc.DefaultErrorCode,
		fmt.Sprintf("this client does not support network %d", networkID),
	)
}

// GetProof returns the proofs needed to claim a bridge. NetworkID and depositCount refere to the bridge origin
// while globalExitRoot should be already injected on the destination network.
// This call needs to be done to a client of the same network were the bridge tx was sent
func (b *BridgeEndpoints) GetProof(
	networkID uint32, depositCount uint32, l1InfoTreeIndex uint32,
) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.readTimeout)
	defer cancel()

	c, merr := b.meter.Int64Counter("claim_proof")
	if merr != nil {
		b.logger.Warnf("failed to create claim_proof counter: %s", merr)
	}
	c.Add(ctx, 1)

	info, err := b.l1InfoTree.GetInfoByIndex(ctx, l1InfoTreeIndex)
	if err != nil {
		return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get info from the tree: %s", err))
	}
	proofRollupExitRoot, err := b.l1InfoTree.GetRollupExitTreeMerkleProof(ctx, networkID, info.GlobalExitRoot)
	if err != nil {
		return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get rollup exit proof, error: %s", err))
	}
	var proofLocalExitRoot tree.Proof
	switch {
	case networkID == 0:
		proofLocalExitRoot, err = b.bridgeL1.GetProof(ctx, depositCount, info.MainnetExitRoot)
		if err != nil {
			return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get local exit proof, error: %s", err))
		}

	case networkID == b.networkID:
		localExitRoot, err := b.l1InfoTree.GetLocalExitRoot(ctx, networkID, info.RollupExitRoot)
		if err != nil {
			return zeroHex, rpc.NewRPCError(
				rpc.DefaultErrorCode,
				fmt.Sprintf("failed to get local exit root from rollup exit tree, error: %s", err),
			)
		}
		proofLocalExitRoot, err = b.bridgeL2.GetProof(ctx, depositCount, localExitRoot)
		if err != nil {
			return zeroHex, rpc.NewRPCError(
				rpc.DefaultErrorCode,
				fmt.Sprintf("failed to get local exit proof, error: %s", err),
			)
		}

	default:
		return zeroHex, rpc.NewRPCError(
			rpc.DefaultErrorCode,
			fmt.Sprintf("this client does not support network %d", networkID),
		)
	}
	return types.ClaimProof{
		ProofLocalExitRoot:  proofLocalExitRoot,
		ProofRollupExitRoot: proofRollupExitRoot,
		L1InfoTreeLeaf:      *info,
	}, nil
}

// SponsorClaim sends a claim tx on behalf of the user.
// This call needs to be done to a client of the same network were the claim is going to be sent (bridge destination)
func (b *BridgeEndpoints) SponsorClaim(claim claimsponsor.Claim) (interface{}, rpc.Error) {
	ctx, cancel := context.WithTimeout(context.Background(), b.writeTimeout)
	defer cancel()

	c, merr := b.meter.Int64Counter("sponsor_claim")
	if merr != nil {
		b.logger.Warnf("failed to create sponsor_claim counter: %s", merr)
	}
	c.Add(ctx, 1)

	if b.sponsor == nil {
		return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, "this client does not support claim sponsoring")
	}
	if claim.DestinationNetwork != b.networkID {
		return zeroHex, rpc.NewRPCError(
			rpc.DefaultErrorCode,
			fmt.Sprintf("this client only sponsors claims for network %d", b.networkID),
		)
	}
	if err := b.sponsor.AddClaimToQueue(&claim); err != nil {
		return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("error adding claim to the queue %s", err))
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
		b.logger.Warnf("failed to create get_sponsored_claim_status counter: %s", merr)
	}
	c.Add(ctx, 1)

	if b.sponsor == nil {
		return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, "this client does not support claim sponsoring")
	}
	claim, err := b.sponsor.GetClaim(globalIndex)
	if err != nil {
		return zeroHex, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("failed to get claim status, error: %s", err))
	}
	return claim.Status, nil
}

func (b *BridgeEndpoints) getFirstL1InfoTreeIndexForL1Bridge(ctx context.Context, depositCount uint32) (uint32, error) {
	lastInfo, err := b.l1InfoTree.GetLastInfo()
	if err != nil {
		return 0, err
	}

	root, err := b.bridgeL1.GetRootByLER(ctx, lastInfo.MainnetExitRoot)
	if err != nil {
		return 0, err
	}
	if root.Index < depositCount {
		return 0, ErrNotOnL1Info
	}

	firstInfo, err := b.l1InfoTree.GetFirstInfo()
	if err != nil {
		return 0, err
	}

	// Binary search between the first and last blcoks where L1 info tree was updated.
	// Find the smallest l1 info tree index that is greater than depositCount and matches with
	// a MER that is included on the l1 info tree
	bestResult := lastInfo
	lowerLimit := firstInfo.BlockNumber
	upperLimit := lastInfo.BlockNumber
	for lowerLimit <= upperLimit {
		targetBlock := lowerLimit + ((upperLimit - lowerLimit) / binnarySearchDivider)
		targetInfo, err := b.l1InfoTree.GetFirstInfoAfterBlock(targetBlock)
		if err != nil {
			return 0, err
		}
		root, err := b.bridgeL1.GetRootByLER(ctx, targetInfo.MainnetExitRoot)
		if err != nil {
			return 0, err
		}
		if root.Index < depositCount {
			lowerLimit = targetBlock + 1
		} else if root.Index == depositCount {
			bestResult = targetInfo
			break
		} else {
			bestResult = targetInfo
			upperLimit = targetBlock - 1
		}
	}

	return bestResult.L1InfoTreeIndex, nil
}

func (b *BridgeEndpoints) getFirstL1InfoTreeIndexForL2Bridge(ctx context.Context, depositCount uint32) (uint32, error) {
	// NOTE: this code assumes that all the rollup exit roots
	// (produced by the smart contract call verifyBatches / verifyBatchesTrustedAggregator)
	// are included in the L1 info tree. As per the current implementation (smart contracts) of the protocol
	// this is true. This could change in the future
	lastVerified, err := b.l1InfoTree.GetLastVerifiedBatches(b.networkID - 1)
	if err != nil {
		return 0, err
	}

	root, err := b.bridgeL2.GetRootByLER(ctx, lastVerified.ExitRoot)
	if err != nil {
		return 0, err
	}
	if root.Index < depositCount {
		return 0, ErrNotOnL1Info
	}

	firstVerified, err := b.l1InfoTree.GetFirstVerifiedBatches(b.networkID - 1)
	if err != nil {
		return 0, err
	}

	// Binary search between the first and last blcoks where batches were verified.
	// Find the smallest deposit count that is greater than depositCount and matches with
	// a LER that is verified
	bestResult := lastVerified
	lowerLimit := firstVerified.BlockNumber
	upperLimit := lastVerified.BlockNumber
	for lowerLimit <= upperLimit {
		targetBlock := lowerLimit + ((upperLimit - lowerLimit) / binnarySearchDivider)
		targetVerified, err := b.l1InfoTree.GetFirstVerifiedBatchesAfterBlock(b.networkID-1, targetBlock)
		if err != nil {
			return 0, err
		}
		root, err = b.bridgeL2.GetRootByLER(ctx, targetVerified.ExitRoot)
		if err != nil {
			return 0, err
		}
		if root.Index < depositCount {
			lowerLimit = targetBlock + 1
		} else if root.Index == depositCount {
			bestResult = targetVerified
			break
		} else {
			bestResult = targetVerified
			upperLimit = targetBlock - 1
		}
	}

	info, err := b.l1InfoTree.GetFirstL1InfoWithRollupExitRoot(bestResult.RollupExitRoot)
	if err != nil {
		return 0, err
	}
	return info.L1InfoTreeIndex, nil
}
