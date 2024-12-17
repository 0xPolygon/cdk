package rpcclient

import (
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/aggsender/types"
)

// Client wraps all the available endpoints of the data abailability committee node server
type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}

func (c *Client) GetCertificateHeaderPerHeight(height *uint64) (*types.CertificateInfo, error) {
	response, err := rpc.JSONRPCCall(c.url, "aggsender_getCertificateHeaderPerHeight", height)
	if err != nil {
		return nil, err
	}

	// Check if the response is an error
	if response.Error != nil {
		return nil, fmt.Errorf("error in the response calling aggsender_getCertificateHeaderPerHeight: %v", response.Error)
	}
	cert := types.CertificateInfo{}
	// Get the batch number from the response hex string
	err = json.Unmarshal(response.Result, &cert)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}
