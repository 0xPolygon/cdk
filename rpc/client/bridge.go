package rpc

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/claimsponsor"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/rpc/types"
)

type BridgeClientInterface interface {
	L1InfoTreeIndexForBridge(networkID uint32, depositCount uint32) (uint32, error)
	InjectedInfoAfterIndex(networkID uint32, l1InfoTreeIndex uint32) (*l1infotreesync.L1InfoTreeLeaf, error)
	ClaimProof(networkID uint32, depositCount uint32, l1InfoTreeIndex uint32) (*types.ClaimProof, error)
	SponsorClaim(claim claimsponsor.Claim) error
	GetSponsoredClaimStatus(globalIndex *big.Int) (claimsponsor.ClaimStatus, error)
}

// L1InfoTreeIndexForBridge returns the first L1 Info Tree index in which the bridge was included.
// networkID represents the origin network.
// This call needs to be done to a client of the same network were the bridge tx was sent
func (c *Client) L1InfoTreeIndexForBridge(networkID uint32, depositCount uint32) (uint32, error) {
	response, err := rpc.JSONRPCCall(c.url, "bridge_l1InfoTreeIndexForBridge", networkID, depositCount)
	if err != nil {
		return 0, err
	}
	if response.Error != nil {
		return 0, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}
	var result uint32
	return result, json.Unmarshal(response.Result, &result)
}

// InjectedInfoAfterIndex return the first GER injected onto the network that is linked
// to the given index or greater. This call is useful to understand when a bridge is ready to be claimed
// on its destination network
func (c *Client) InjectedInfoAfterIndex(
	networkID uint32, l1InfoTreeIndex uint32,
) (*l1infotreesync.L1InfoTreeLeaf, error) {
	response, err := rpc.JSONRPCCall(c.url, "bridge_injectedInfoAfterIndex", networkID, l1InfoTreeIndex)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}
	var result l1infotreesync.L1InfoTreeLeaf
	return &result, json.Unmarshal(response.Result, &result)
}

// ClaimProof returns the proofs needed to claim a bridge. NetworkID and depositCount refere to the bridge origin
// while globalExitRoot should be already injected on the destination network.
// This call needs to be done to a client of the same network were the bridge tx was sent
func (c *Client) ClaimProof(networkID uint32, depositCount uint32, l1InfoTreeIndex uint32) (*types.ClaimProof, error) {
	response, err := rpc.JSONRPCCall(c.url, "bridge_claimProof", networkID, depositCount, l1InfoTreeIndex)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}
	var result types.ClaimProof
	return &result, json.Unmarshal(response.Result, &result)
}

// SponsorClaim sends a claim tx on behalf of the user.
// This call needs to be done to a client of the same network were the claim is going to be sent (bridge destination)
func (c *Client) SponsorClaim(claim claimsponsor.Claim) error {
	response, err := rpc.JSONRPCCall(c.url, "bridge_sponsorClaim", claim)
	if err != nil {
		return err
	}
	if response.Error != nil {
		return fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}
	return nil
}

// GetSponsoredClaimStatus returns the status of a claim that has been previously requested to be sponsored.
// This call needs to be done to the same client were it was requested to be sponsored
func (c *Client) GetSponsoredClaimStatus(globalIndex *big.Int) (claimsponsor.ClaimStatus, error) {
	response, err := rpc.JSONRPCCall(c.url, "bridge_getSponsoredClaimStatus", globalIndex)
	if err != nil {
		return "", err
	}
	if response.Error != nil {
		return "", fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}
	var result claimsponsor.ClaimStatus
	return result, json.Unmarshal(response.Result, &result)
}
