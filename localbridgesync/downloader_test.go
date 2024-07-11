package localbridgesync

import (
	"context"
	"math/big"
	"testing"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridge"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridgev2"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	contractAddr = common.HexToAddress("1234567890")
)

const (
	syncBlockChunck = 10
)

func TestGetEventsByBlockRange(t *testing.T) {
	type testCase struct {
		inputLogs          []types.Log
		fromBlock, toBlock uint64
		expectedBlocks     []block
	}
	testCases := []testCase{}
	clientMock := NewL2Mock(t)
	ctx := context.Background()
	d, err := newDownloader(contractAddr, syncBlockChunck, clientMock)
	require.NoError(t, err)

	// case 1
	log, bridge := generateBridge(t, 1)
	logs := []types.Log{
		*log,
	}
	blocks := []block{
		{
			blockHeader: blockHeader{
				Num:  log.BlockNumber,
				Hash: log.BlockHash,
			},
			Events: bridgeEvents{
				Bridges: []Bridge{bridge},
				Claims:  []Claim{},
			},
		},
	}
	case1 := testCase{
		inputLogs:      logs,
		fromBlock:      0,
		toBlock:        1,
		expectedBlocks: blocks,
	}
	testCases = append(testCases, case1)

	for _, tc := range testCases {
		clientMock.
			On("FilterLogs", mock.Anything, mock.Anything).
			Return(tc.inputLogs, nil)

		actualBlocks := d.getEventsByBlockRange(ctx, 0, 1)
		assert.Equal(t, tc.expectedBlocks, actualBlocks)
	}
}

func generateBridge(t *testing.T, blockNum uint32) (*types.Log, Bridge) {
	b := Bridge{
		LeafType:           1,
		OriginNetwork:      blockNum,
		OriginAddress:      contractAddr,
		DestinationNetwork: blockNum,
		DestinationAddress: contractAddr,
		Amount:             big.NewInt(int64(blockNum)),
		Metadata:           common.Hex2Bytes("01"),
		DepositCount:       blockNum,
	}
	abi, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	require.NoError(t, err)
	event, err := abi.EventByID(bridgeEventSignature)
	require.NoError(t, err)
	data, err := event.Inputs.Pack(
		b.LeafType,
		b.OriginNetwork,
		b.OriginAddress,
		b.DestinationNetwork,
		b.DestinationAddress,
		b.Amount,
		b.Metadata,
		b.DepositCount,
	)
	require.NoError(t, err)
	log := &types.Log{
		Address:     contractAddr,
		BlockNumber: uint64(blockNum),
		BlockHash:   common.BytesToHash(blockNum2Bytes(uint64(blockNum))),
		Topics:      []common.Hash{bridgeEventSignature},
		Data:        data,
	}
	return log, b
}

func generateClaimV1(t *testing.T, blockNum uint32) (*types.Log, Claim) {
	abi, err := polygonzkevmbridge.PolygonzkevmbridgeMetaData.GetAbi()
	require.NoError(t, err)
	event, err := abi.EventByID(claimEventSignaturePreEtrog)
	require.NoError(t, err)
	return generateClaim(t, blockNum, event)
}

func generateClaimV2(t *testing.T, blockNum uint32) (*types.Log, Claim) {
	abi, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	require.NoError(t, err)
	event, err := abi.EventByID(claimEventSignature)
	require.NoError(t, err)
	return generateClaim(t, blockNum, event)
}

func generateClaim(t *testing.T, blockNum uint32, event *abi.Event) (*types.Log, Claim) {
	c := Claim{
		GlobalIndex:        big.NewInt(int64(blockNum)),
		OriginNetwork:      blockNum,
		OriginAddress:      contractAddr,
		DestinationAddress: contractAddr,
		Amount:             big.NewInt(int64(blockNum)),
	}
	data, err := event.Inputs.Pack(
		c.GlobalIndex,
		c.OriginNetwork,
		c.OriginAddress,
		c.DestinationAddress,
		c.Amount,
	)
	require.NoError(t, err)
	log := &types.Log{
		Address:     contractAddr,
		BlockNumber: uint64(blockNum),
		BlockHash:   common.BytesToHash(blockNum2Bytes(uint64(blockNum))),
		Topics:      []common.Hash{bridgeEventSignature},
		Data:        data,
	}
	return log, c
}
