package rpcclient

import (
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/aggsender/types"
)

var jSONRPCCall = rpc.JSONRPCCall

// Client wraps all the available endpoints of the data abailability committee node server
type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

func (c *Client) GetStatus() (*types.AggsenderInfo, error) {
	response, err := jSONRPCCall(c.url, "aggsender_status")
	if err != nil {
		return nil, err
	}

	// Check if the response is an error
	if response.Error != nil {
		return nil, fmt.Errorf("error in the response calling aggsender_status: %w", response.Error)
	}
	result := types.AggsenderInfo{}
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetCertificateHeaderPerHeight(height *uint64) (*types.CertificateInfo, error) {
	response, err := jSONRPCCall(c.url, "aggsender_getCertificateHeaderPerHeight", height)
	if err != nil {
		return nil, err
	}

	// Check if the response is an error
	if response.Error != nil {
		return nil, fmt.Errorf("error in the response calling aggsender_getCertificateHeaderPerHeight: %w", response.Error)
	}
	cert := types.CertificateInfo{}
	err = json.Unmarshal(response.Result, &cert)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}
