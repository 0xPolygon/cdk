package agglayer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk-rpc/types"
	"github.com/ethereum/go-ethereum/common"
)

const errCodeAgglayerRateLimitExceeded int = -10007

var (
	ErrAgglayerRateLimitExceeded = fmt.Errorf("agglayer rate limit exceeded")
	jSONRPCCall                  = rpc.JSONRPCCall
)

type AggLayerClientGetEpochConfiguration interface {
	GetEpochConfiguration() (*ClockConfiguration, error)
}

// AgglayerClientInterface is the interface that defines the methods that the AggLayerClient will implement
type AgglayerClientInterface interface {
	SendTx(signedTx SignedTx) (common.Hash, error)
	WaitTxToBeMined(hash common.Hash, ctx context.Context) error
	SendCertificate(certificate *SignedCertificate) (common.Hash, error)
	GetCertificateHeader(certificateHash common.Hash) (*CertificateHeader, error)
	AggLayerClientGetEpochConfiguration
}

// AggLayerClient is the client that will be used to interact with the AggLayer
type AggLayerClient struct {
	url string
}

// NewAggLayerClient returns a client ready to be used
func NewAggLayerClient(url string) *AggLayerClient {
	return &AggLayerClient{
		url: url,
	}
}

// SendTx sends a signed transaction to the AggLayer
func (c *AggLayerClient) SendTx(signedTx SignedTx) (common.Hash, error) {
	response, err := rpc.JSONRPCCall(c.url, "interop_sendTx", signedTx)
	if err != nil {
		return common.Hash{}, err
	}

	if response.Error != nil {
		if response.Error.Code == errCodeAgglayerRateLimitExceeded {
			return common.Hash{}, ErrAgglayerRateLimitExceeded
		}
		return common.Hash{}, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result types.ArgHash
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return common.Hash{}, err
	}

	return result.Hash(), nil
}

// WaitTxToBeMined waits for a transaction to be mined
func (c *AggLayerClient) WaitTxToBeMined(hash common.Hash, ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return errors.New("context finished before tx was mined")
		case <-ticker.C:
			response, err := rpc.JSONRPCCall(c.url, "interop_getTxStatus", hash)
			if err != nil {
				return err
			}

			if response.Error != nil {
				return fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
			}

			var result string
			err = json.Unmarshal(response.Result, &result)
			if err != nil {
				return err
			}
			if strings.ToLower(result) == "done" {
				return nil
			}
		}
	}
}

// SendCertificate sends a certificate to the AggLayer
func (c *AggLayerClient) SendCertificate(certificate *SignedCertificate) (common.Hash, error) {
	certificateToSend := certificate.CopyWithDefaulting()

	response, err := rpc.JSONRPCCall(c.url, "interop_sendCertificate", certificateToSend)
	if err != nil {
		return common.Hash{}, err
	}

	if response.Error != nil {
		return common.Hash{}, fmt.Errorf("%d %s", response.Error.Code, response.Error.Message)
	}

	var result types.ArgHash
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return common.Hash{}, err
	}

	return result.Hash(), nil
}

// GetCertificateHeader returns the certificate header associated to the hash
func (c *AggLayerClient) GetCertificateHeader(certificateHash common.Hash) (*CertificateHeader, error) {
	response, err := rpc.JSONRPCCall(c.url, "interop_getCertificateHeader", certificateHash)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%d %s", response.Error.Code, response.Error.Message)
	}

	var result *CertificateHeader
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetEpochConfiguration returns the clock configuration of AggLayer
func (c *AggLayerClient) GetEpochConfiguration() (*ClockConfiguration, error) {
	response, err := jSONRPCCall(c.url, "interop_getEpochConfiguration")
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("GetEpochConfiguration code=%d msg=%s", response.Error.Code, response.Error.Message)
	}

	var result *ClockConfiguration
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
