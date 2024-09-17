package sequencesender

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender/seqsendertypes/rpcbatch"
	"github.com/0xPolygon/cdk/state"
	"github.com/ethereum/go-ethereum/common"
)

func (s *SequenceSender) getBatchFromRPC(batchNumber uint64) (*rpcbatch.RPCBatch, error) {
	type zkEVMBatch struct {
		Blocks         []string `json:"blocks"`
		BatchL2Data    string   `json:"batchL2Data"`
		Coinbase       string   `json:"coinbase"`
		GlobalExitRoot string   `json:"globalExitRoot"`
		Closed         bool     `json:"closed"`
		Timestamp      string   `json:"timestamp"`
	}

	zkEVMBatchData := zkEVMBatch{}

	log.Infof("Getting batch %d from RPC", batchNumber)

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
		return nil, fmt.Errorf("error in the response calling zkevm_getBatchByNumber: %w", response.Error)
	}

	// Get the batch number from the response hex string
	err = json.Unmarshal(response.Result, &zkEVMBatchData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling the batch from the response calling zkevm_getBatchByNumber: %w", err)
	}

	rpcBatch, err := rpcbatch.New(batchNumber, zkEVMBatchData.Blocks, common.Hex2Bytes(zkEVMBatchData.BatchL2Data),
		common.HexToHash(zkEVMBatchData.GlobalExitRoot), common.HexToAddress(zkEVMBatchData.Coinbase), zkEVMBatchData.Closed)
	if err != nil {
		return nil, fmt.Errorf("error creating the rpc batch: %w", err)
	}

	if len(zkEVMBatchData.Blocks) > 0 {
		lastL2BlockTimestamp, err := s.getL2BlockTimestampFromRPC(zkEVMBatchData.Blocks[len(zkEVMBatchData.Blocks)-1])
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

func (s *SequenceSender) getL2BlockTimestampFromRPC(blockHash string) (uint64, error) {
	type zkeEVML2Block struct {
		Timestamp string `json:"timestamp"`
	}

	log.Infof("Getting l2 block timestamp from RPC. Block hash: %s", blockHash)

	response, err := rpc.JSONRPCCall(s.cfg.RPCURL, "eth_getBlockByHash", blockHash, false)
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
