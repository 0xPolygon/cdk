package rpc

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/claimsponsor"
	"github.com/0xPolygon/cdk/l1infotreesync"
)

type BridgeClientInterface interface {
	L1InfoTreeIndexForBridge(networkID uint32, depositCount uint32) (uint32, error)
	InjectedInfoAfterIndex(networkID uint32, l1InfoTreeIndex uint32) (*l1infotreesync.L1InfoTreeLeaf, error)
	ClaimProof(networkID uint32, depositCount uint32, l1InfoTreeIndex uint32) (*ClaimProof, error)
	SponsorClaim(claim claimsponsor.Claim) error
	GetSponsoredClaimStatus(globalIndex *big.Int) (claimsponsor.ClaimStatus, error)
}

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

func (c *Client) InjectedInfoAfterIndex(networkID uint32, l1InfoTreeIndex uint32) (*l1infotreesync.L1InfoTreeLeaf, error) {
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

func (c *Client) ClaimProof(networkID uint32, depositCount uint32, l1InfoTreeIndex uint32) (*ClaimProof, error) {
	response, err := rpc.JSONRPCCall(c.url, "bridge_claimProof", networkID, depositCount, l1InfoTreeIndex)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}
	var result ClaimProof
	return &result, json.Unmarshal(response.Result, &result)
}

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
