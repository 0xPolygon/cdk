package sequencesender

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes/rpcbatch"
	"github.com/0xPolygon/cdk/state"
	"github.com/ethereum/go-ethereum/common"
)

func (s *SequenceSender) getBatchFromRPC(batchNumber uint64) (*rpcbatch.RPCBatch, error) {
	type zkEVMBatch struct {
		Blocks         []string `mapstructure:"blocks"`
		BatchL2Data    string   `mapstructure:"batchL2Data"`
		Coinbase       string   `mapstructure:"coinbase"`
		GlobalExitRoot string   `mapstructure:"globalExitRoot"`
		Closed         bool     `mapstructure:"closed"`
	}

	zkEVMBatchData := zkEVMBatch{}

	response, err := rpc.JSONRPCCall(s.cfg.RPCURL, "zkevm_getBatchByNumber", batchNumber)
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
		return nil, fmt.Errorf("error unmarshalling the batch from the response calling zkevm_getBatchByNumber: %v", err)
	}

	rpcBatch, err := rpcbatch.New(batchNumber, zkEVMBatchData.Blocks, common.Hex2Bytes(zkEVMBatchData.BatchL2Data), common.HexToHash(zkEVMBatchData.GlobalExitRoot), common.HexToAddress(zkEVMBatchData.Coinbase), zkEVMBatchData.Closed)
	if err != nil {
		return nil, fmt.Errorf("error creating the rpc batch: %v", err)
	}

	lastL2BlockTimestamp, err := s.getL2BlockTimestampFromRPC(zkEVMBatchData.Blocks[len(zkEVMBatchData.Blocks)-1])
	if err != nil {
		return nil, fmt.Errorf("error getting the last l2 block timestamp from the rpc: %v", err)
	}

	rpcBatch.SetLastL2BLockTimestamp(lastL2BlockTimestamp)

	return rpcBatch, nil
}

func (s *SequenceSender) getL2BlockTimestampFromRPC(blockHash string) (uint64, error) {
	type zkeEVML2Block struct {
		timestamp string `mapstructure:"timestamp"`
	}

	l2Block := zkeEVML2Block{}

	response, err := rpc.JSONRPCCall(s.cfg.RPCURL, "eth_getBlockByHash", blockHash, false)
	if err != nil {
		return 0, err
	}

	// Check if the response is an error
	if response.Error != nil {
		return 0, fmt.Errorf("error in the response calling eth_getBlockByHash: %v", response.Error)
	}

	// Get the batch number from the response hex string
	err = json.Unmarshal(response.Result, &l2Block)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling the l2 block from the response calling eth_getBlockByHash: %v", err)
	}

	return new(big.Int).SetBytes(common.FromHex(l2Block.timestamp)).Uint64(), nil
}
