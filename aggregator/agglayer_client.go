package aggregator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	aggTx "github.com/0xPolygon/agglayer/tx"
	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk-rpc/types"
	"github.com/ethereum/go-ethereum/common"
)

// AgglayerClientInterface is the interface that defines the methods that the AggLayerClient will implement
type AgglayerClientInterface interface {
	SendTx(signedTx aggTx.SignedTx) (common.Hash, error)
	WaitTxToBeMined(hash common.Hash, ctx context.Context) error
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
func (c *AggLayerClient) SendTx(signedTx aggTx.SignedTx) (common.Hash, error) {
	response, err := rpc.JSONRPCCall(c.url, "interop_sendTx", signedTx)
	if err != nil {
		return common.Hash{}, err
	}

	if response.Error != nil {
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
