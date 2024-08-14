package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/hex"
	"github.com/0xPolygonHermez/zkevm-synchronizer-l1/jsonrpcclient/types"
	"github.com/ethereum/go-ethereum/common"
)

// ZKEVMClientInterface is the interface that defines the implementation of all the zkevm endpoints
type ZKEVMClientInterface interface {
	BatchNumber(ctx context.Context) (uint64, error)
	BatchNumberByBlockNumber(ctx context.Context, blockNum uint64) (uint64, error)
	BatchByNumber(ctx context.Context, number *big.Int) (*types.Batch, error)
	BatchesByNumbers(ctx context.Context, numbers []*big.Int) ([]*types.BatchData, error)
	ForcedBatchesByNumbers(ctx context.Context, numbers []*big.Int) ([]*types.BatchData, error)
	batchesByNumbers(_ context.Context, numbers []*big.Int, method string) ([]*types.BatchData, error)
	ExitRootsByGER(ctx context.Context, globalExitRoot common.Hash) (*types.ExitRoots, error)
	GetLatestGlobalExitRoot(ctx context.Context) (common.Hash, error)
}

// ZKEVMClientFactoryInterface interface for the client factory
type ZKEVMClientFactoryInterface interface {
	NewZKEVMClient(url string) ZKEVMClientInterface
}

// ZKEVMClientFactory is the implementation of the data committee client factory
type ZKEVMClientFactory struct{}

// NewZKEVMClient returns an implementation of the data committee node client
func (f *ZKEVMClientFactory) NewZKEVMClient(url string) ZKEVMClientInterface {
	return NewZKEVMClient(url)
}

// ZKEVMClient wraps all the available endpoints of the data abailability committee node server
type ZKEVMClient struct {
	url string
}

// NewZKEVMClient returns a client ready to be used
func NewZKEVMClient(url string) *ZKEVMClient {
	return &ZKEVMClient{
		url: url,
	}
}

// BatchNumber returns the latest batch number
func (c *ZKEVMClient) BatchNumber(ctx context.Context) (uint64, error) {
	response, err := rpc.JSONRPCCall(c.url, "zkevm_batchNumber")
	if err != nil {
		return 0, err
	}

	if response.Error != nil {
		return 0, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result string
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return 0, err
	}

	bigBatchNumber := hex.DecodeBig(result)
	batchNumber := bigBatchNumber.Uint64()

	return batchNumber, nil
}

// BatchByNumber returns a batch from the current canonical chain. If number is nil, the
// latest known batch is returned.
func (c *ZKEVMClient) BatchByNumber(ctx context.Context, number *big.Int) (*types.Batch, error) {
	bn := types.LatestBatchNumber
	if number != nil {
		bn = types.BatchNumber(number.Int64())
	}
	response, err := rpc.JSONRPCCall(c.url, "zkevm_getBatchByNumber", bn.StringOrHex(), true)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result *types.Batch
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// BatchNumberByBlockNumber returns the batch number in which the block is includded
func (c *ZKEVMClient) BatchNumberByBlockNumber(ctx context.Context, blockNum uint64) (uint64, error) {
	bn := types.BatchNumber(int64(blockNum))
	response, err := rpc.JSONRPCCall(c.url, "zkevm_getBatchByNumber", bn.StringOrHex(), true)
	if err != nil {
		return 0, err
	}

	if response.Error != nil {
		return 0, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result *types.BatchNumber
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return 0, err
	}

	return uint64(*result), nil
}

// BatchesByNumbers returns batches from the current canonical chain by batch numbers. If the list is empty, the last
// known batch is returned as a list.
func (c *ZKEVMClient) BatchesByNumbers(ctx context.Context, numbers []*big.Int) ([]*types.BatchData, error) {
	return c.batchesByNumbers(ctx, numbers, "zkevm_getBatchDataByNumbers")
}

// ForcedBatchesByNumbers returns forced batches data.
func (c *ZKEVMClient) ForcedBatchesByNumbers(ctx context.Context, numbers []*big.Int) ([]*types.BatchData, error) {
	return c.batchesByNumbers(ctx, numbers, "zkevm_getForcedBatchDataByNumbers")
}

// BatchesByNumbers returns batches from the current canonical chain by batch numbers. If the list is empty, the last
// known batch is returned as a list.
func (c *ZKEVMClient) batchesByNumbers(_ context.Context, numbers []*big.Int, method string) ([]*types.BatchData, error) {
	batchNumbers := make([]types.BatchNumber, 0, len(numbers))
	for _, n := range numbers {
		batchNumbers = append(batchNumbers, types.BatchNumber(n.Int64()))
	}
	if len(batchNumbers) == 0 {
		batchNumbers = append(batchNumbers, types.LatestBatchNumber)
	}

	response, err := rpc.JSONRPCCall(c.url, method, &types.BatchFilter{Numbers: batchNumbers})
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result *types.BatchDataResult
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// ExitRootsByGER returns the exit roots accordingly to the provided Global Exit Root
func (c *ZKEVMClient) ExitRootsByGER(ctx context.Context, globalExitRoot common.Hash) (*types.ExitRoots, error) {
	response, err := rpc.JSONRPCCall(c.url, "zkevm_getExitRootsByGER", globalExitRoot.String())
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result *types.ExitRoots
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetLatestGlobalExitRoot returns the latest global exit root
func (c *ZKEVMClient) GetLatestGlobalExitRoot(ctx context.Context) (common.Hash, error) {
	response, err := rpc.JSONRPCCall(c.url, "zkevm_getLatestGlobalExitRoot")
	if err != nil {
		return common.Hash{}, err
	}

	if response.Error != nil {
		return common.Hash{}, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result string
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return common.Hash{}, err
	}

	return common.HexToHash(result), nil
}
