package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/rpc/types"
	"github.com/0xPolygon/cdk/state"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// ErrBusy is returned when the witness server is busy
	ErrBusy = errors.New("witness server is busy")
)

const busyResponse = "busy"

type BatchEndpoints struct {
	url string
}

func NewBatchEndpoints(url string) *BatchEndpoints {
	return &BatchEndpoints{url: url}
}

func (b *BatchEndpoints) GetBatch(batchNumber uint64) (*types.RPCBatch, error) {
	type zkEVMBatch struct {
		AccInputHash   string   `json:"accInputHash"`
		Blocks         []string `json:"blocks"`
		BatchL2Data    string   `json:"batchL2Data"`
		Coinbase       string   `json:"coinbase"`
		GlobalExitRoot string   `json:"globalExitRoot"`
		LocalExitRoot  string   `json:"localExitRoot"`
		StateRoot      string   `json:"stateRoot"`
		Closed         bool     `json:"closed"`
		Timestamp      string   `json:"timestamp"`
	}

	zkEVMBatchData := zkEVMBatch{}

	log.Infof("Getting batch %d from RPC", batchNumber)

	response, err := rpc.JSONRPCCall(b.url, "zkevm_getBatchByNumber", batchNumber)
	if err != nil {
		return nil, err
	}

	// Check if the response is nil
	if response.Result == nil {
		return nil, state.ErrNotFound
	}

	// Check if the response is an error
	if response.Error != nil {
		return nil, fmt.Errorf("error in the response calling zkevm_getBatchByNumber: %v", response.Error)
	}

	// Get the batch number from the response hex string
	err = json.Unmarshal(response.Result, &zkEVMBatchData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling the batch from the response calling zkevm_getBatchByNumber: %w", err)
	}

	rpcBatch := types.NewRPCBatch(batchNumber, common.HexToHash(zkEVMBatchData.AccInputHash), zkEVMBatchData.Blocks,
		common.FromHex(zkEVMBatchData.BatchL2Data), common.HexToHash(zkEVMBatchData.GlobalExitRoot),
		common.HexToHash(zkEVMBatchData.LocalExitRoot), common.HexToHash(zkEVMBatchData.StateRoot),
		common.HexToAddress(zkEVMBatchData.Coinbase), zkEVMBatchData.Closed)

	if len(zkEVMBatchData.Blocks) > 0 {
		lastL2BlockTimestamp, err := b.GetL2BlockTimestamp(zkEVMBatchData.Blocks[len(zkEVMBatchData.Blocks)-1])
		if err != nil {
			return nil, fmt.Errorf("error getting the last l2 block timestamp from the rpc: %w", err)
		}
		rpcBatch.SetLastL2BLockTimestamp(lastL2BlockTimestamp)
	} else {
		log.Infof("No blocks in the batch, setting the last l2 block timestamp from the batch data")
		rpcBatch.SetLastL2BLockTimestamp(new(big.Int).SetBytes(common.FromHex(zkEVMBatchData.Timestamp)).Uint64())
	}

	return rpcBatch, nil
}

func (b *BatchEndpoints) GetL2BlockTimestamp(blockHash string) (uint64, error) {
	type zkeEVML2Block struct {
		Timestamp string `json:"timestamp"`
	}

	log.Infof("Getting l2 block timestamp from RPC. Block hash: %s", blockHash)

	response, err := rpc.JSONRPCCall(b.url, "eth_getBlockByHash", blockHash, false)
	if err != nil {
		return 0, err
	}

	// Check if the response is an error
	if response.Error != nil {
		return 0, fmt.Errorf("error in the response calling eth_getBlockByHash: %v", response.Error)
	}

	//  Get the l2 block from the response
	l2Block := zkeEVML2Block{}
	err = json.Unmarshal(response.Result, &l2Block)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling the l2 block from the response calling eth_getBlockByHash: %w", err)
	}

	return new(big.Int).SetBytes(common.FromHex(l2Block.Timestamp)).Uint64(), nil
}

func (b *BatchEndpoints) GetWitness(batchNumber uint64, fullWitness bool) ([]byte, error) {
	var (
		witness  string
		response rpc.Response
		err      error
	)

	witnessType := "trimmed"
	if fullWitness {
		witnessType = "full"
	}

	log.Infof("Requesting witness for batch %d of type %s", batchNumber, witnessType)

	response, err = rpc.JSONRPCCall(b.url, "zkevm_getBatchWitness", batchNumber, witnessType)
	if err != nil {
		return nil, err
	}

	// Check if the response is an error
	if response.Error != nil {
		if response.Error.Message == busyResponse {
			return nil, ErrBusy
		}

		return nil, fmt.Errorf("error from witness for batch %d: %v", batchNumber, response.Error)
	}

	err = json.Unmarshal(response.Result, &witness)
	if err != nil {
		return nil, err
	}

	return common.FromHex(witness), nil
}
