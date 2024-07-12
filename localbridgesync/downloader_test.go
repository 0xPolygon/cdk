package localbridgesync

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridge"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonzkevmbridgev2"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
		description        string
		inputLogs          []types.Log
		fromBlock, toBlock uint64
		expectedBlocks     []block
	}
	testCases := []testCase{}
	clientMock := NewL2Mock(t)
	ctx := context.Background()
	d, err := newDownloader(contractAddr, clientMock)
	require.NoError(t, err)

	// case 0: single block, no events
	case0 := testCase{
		description:    "case 0: single block, no events",
		inputLogs:      []types.Log{},
		fromBlock:      1,
		toBlock:        3,
		expectedBlocks: []block{},
	}
	testCases = append(testCases, case0)

	// case 1: single block, single event
	logC1, bridgeC1 := generateBridge(t, 3)
	logsC1 := []types.Log{
		*logC1,
	}
	blocksC1 := []block{
		{
			blockHeader: blockHeader{
				Num:  logC1.BlockNumber,
				Hash: logC1.BlockHash,
			},
			Events: bridgeEvents{
				Bridges: []Bridge{bridgeC1},
				Claims:  []Claim{},
			},
		},
	}
	case1 := testCase{
		description:    "case 1: single block, single event",
		inputLogs:      logsC1,
		fromBlock:      3,
		toBlock:        3,
		expectedBlocks: blocksC1,
	}
	testCases = append(testCases, case1)

	// case 2: single block, multiple events
	logC2_1, bridgeC2_1 := generateBridge(t, 5)
	logC2_2, bridgeC2_2 := generateBridge(t, 5)
	logC2_3, claimC2_1 := generateClaimV1(t, 5)
	logC2_4, claimC2_2 := generateClaimV2(t, 5)
	logsC2 := []types.Log{
		*logC2_1,
		*logC2_2,
		*logC2_3,
		*logC2_4,
	}
	blocksC2 := []block{
		{
			blockHeader: blockHeader{
				Num:  logC2_1.BlockNumber,
				Hash: logC2_1.BlockHash,
			},
			Events: bridgeEvents{
				Bridges: []Bridge{bridgeC2_1, bridgeC2_2},
				Claims:  []Claim{claimC2_1, claimC2_2},
			},
		},
	}
	case2 := testCase{
		description:    "case 2: single block, multiple events",
		inputLogs:      logsC2,
		fromBlock:      5,
		toBlock:        5,
		expectedBlocks: blocksC2,
	}
	testCases = append(testCases, case2)

	// case 3: multiple blocks, some events
	logC3_1, bridgeC3_1 := generateBridge(t, 7)
	logC3_2, bridgeC3_2 := generateBridge(t, 7)
	logC3_3, claimC3_1 := generateClaimV1(t, 8)
	logC3_4, claimC3_2 := generateClaimV2(t, 8)
	logsC3 := []types.Log{
		*logC3_1,
		*logC3_2,
		*logC3_3,
		*logC3_4,
	}
	blocksC3 := []block{
		{
			blockHeader: blockHeader{
				Num:  logC3_1.BlockNumber,
				Hash: logC3_1.BlockHash,
			},
			Events: bridgeEvents{
				Bridges: []Bridge{bridgeC3_1, bridgeC3_2},
				Claims:  []Claim{},
			},
		},
		{
			blockHeader: blockHeader{
				Num:  logC3_3.BlockNumber,
				Hash: logC3_3.BlockHash,
			},
			Events: bridgeEvents{
				Bridges: []Bridge{},
				Claims:  []Claim{claimC3_1, claimC3_2},
			},
		},
	}
	case3 := testCase{
		description:    "case 3: multiple blocks, some events",
		inputLogs:      logsC3,
		fromBlock:      7,
		toBlock:        8,
		expectedBlocks: blocksC3,
	}
	testCases = append(testCases, case3)

	for _, tc := range testCases {
		query := ethereum.FilterQuery{
			FromBlock: new(big.Int).SetUint64(tc.fromBlock),
			Addresses: []common.Address{d.bridgeAddr},
			Topics: [][]common.Hash{
				{bridgeEventSignature},
				{claimEventSignature},
				{claimEventSignaturePreEtrog},
			},
			ToBlock: new(big.Int).SetUint64(tc.toBlock),
		}
		clientMock.
			On("FilterLogs", mock.Anything, query).
			Return(tc.inputLogs, nil)

		actualBlocks := d.getEventsByBlockRange(ctx, tc.fromBlock, tc.toBlock)
		require.Equal(t, tc.expectedBlocks, actualBlocks, tc.description)
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
	return generateClaim(t, blockNum, event, true)
}

func generateClaimV2(t *testing.T, blockNum uint32) (*types.Log, Claim) {
	abi, err := polygonzkevmbridgev2.Polygonzkevmbridgev2MetaData.GetAbi()
	require.NoError(t, err)
	event, err := abi.EventByID(claimEventSignature)
	require.NoError(t, err)
	return generateClaim(t, blockNum, event, false)
}

func generateClaim(t *testing.T, blockNum uint32, event *abi.Event, isV1 bool) (*types.Log, Claim) {
	c := Claim{
		GlobalIndex:        big.NewInt(int64(blockNum)),
		OriginNetwork:      blockNum,
		OriginAddress:      contractAddr,
		DestinationAddress: contractAddr,
		Amount:             big.NewInt(int64(blockNum)),
	}
	var (
		data      []byte
		err       error
		signature common.Hash
	)
	if isV1 {
		data, err = event.Inputs.Pack(
			uint32(c.GlobalIndex.Uint64()),
			c.OriginNetwork,
			c.OriginAddress,
			c.DestinationAddress,
			c.Amount,
		)
		signature = claimEventSignaturePreEtrog
	} else {
		data, err = event.Inputs.Pack(
			c.GlobalIndex,
			c.OriginNetwork,
			c.OriginAddress,
			c.DestinationAddress,
			c.Amount,
		)
		signature = claimEventSignature
	}
	require.NoError(t, err)
	log := &types.Log{
		Address:     contractAddr,
		BlockNumber: uint64(blockNum),
		BlockHash:   common.BytesToHash(blockNum2Bytes(uint64(blockNum))),
		Topics:      []common.Hash{signature},
		Data:        data,
	}
	return log, c
}

func TestDownload(t *testing.T) {
	/*
		NOTE: due to the concurrent nature of this test (the function being tested runs through a goroutine)
		if the mock doesn't match, the goroutine will get stuck and the test will timeout
	*/
	d := NewDownloaderMock(t)
	downloadCh := make(chan block, 1)
	ctx := context.Background()
	ctx1, cancel := context.WithCancel(ctx)
	expectedBlocks := []block{}

	d.On("waitForNewBlocks", mock.Anything, uint64(0)).
		Return(uint64(1))
	// iteratiion 0:
	// last block is 1, download that block (no events and wait)
	b1 := block{
		blockHeader: blockHeader{
			Num:  1,
			Hash: common.HexToHash("01"),
		},
	}
	expectedBlocks = append(expectedBlocks, b1)
	d.On("getEventsByBlockRange", mock.Anything, uint64(0), uint64(1)).
		Return([]block{})
	d.On("getBlockHeader", mock.Anything, uint64(1)).
		Return(b1.blockHeader)

	// iteration 1: wait for next block to be created
	d.On("waitForNewBlocks", mock.Anything, uint64(1)).
		After(time.Millisecond * 100).
		Return(uint64(2)).Once()

	// iteration 2: block 2 has events
	b2 := block{
		blockHeader: blockHeader{
			Num:  2,
			Hash: common.HexToHash("02"),
		},
	}
	expectedBlocks = append(expectedBlocks, b2)
	d.On("getEventsByBlockRange", mock.Anything, uint64(2), uint64(2)).
		Return([]block{b2})

	// iteration 3: wait for next block to be created (jump to block 8)
	d.On("waitForNewBlocks", mock.Anything, uint64(2)).
		After(time.Millisecond * 100).
		Return(uint64(8)).Once()

	// iteration 4: blocks 6 and 7 have events
	b6 := block{
		blockHeader: blockHeader{
			Num:  6,
			Hash: common.HexToHash("06"),
		},
		Events: bridgeEvents{
			Claims: []Claim{
				{OriginNetwork: 6},
			},
			Bridges: []Bridge{},
		},
	}
	b7 := block{
		blockHeader: blockHeader{
			Num:  7,
			Hash: common.HexToHash("07"),
		},
		Events: bridgeEvents{
			Claims: []Claim{},
			Bridges: []Bridge{
				{DestinationNetwork: 7},
			},
		},
	}
	b8 := block{
		blockHeader: blockHeader{
			Num:  8,
			Hash: common.HexToHash("08"),
		},
	}
	expectedBlocks = append(expectedBlocks, b6, b7, b8)
	d.On("getEventsByBlockRange", mock.Anything, uint64(3), uint64(8)).
		Return([]block{b6, b7})
	d.On("getBlockHeader", mock.Anything, uint64(8)).
		Return(b8.blockHeader)

	// iteration 5: wait for next block to be created (jump to block 30)
	d.On("waitForNewBlocks", mock.Anything, uint64(8)).
		After(time.Millisecond * 100).
		Return(uint64(30)).Once()

	// iteration 6: from block 9 to 19, no events
	b19 := block{
		blockHeader: blockHeader{
			Num:  19,
			Hash: common.HexToHash("19"),
		},
	}
	expectedBlocks = append(expectedBlocks, b19)
	d.On("getEventsByBlockRange", mock.Anything, uint64(9), uint64(19)).
		Return([]block{})
	d.On("getBlockHeader", mock.Anything, uint64(19)).
		Return(b19.blockHeader)

	// iteration 7: from block 20 to 30, events on last block
	b30 := block{
		blockHeader: blockHeader{
			Num:  30,
			Hash: common.HexToHash("30"),
		},
		Events: bridgeEvents{
			Claims: []Claim{},
			Bridges: []Bridge{
				{DestinationNetwork: 30},
			},
		},
	}
	expectedBlocks = append(expectedBlocks, b30)
	d.On("getEventsByBlockRange", mock.Anything, uint64(20), uint64(30)).
		Return([]block{b30})

	// iteration 8: wait for next block to be created (jump to block 35)
	d.On("waitForNewBlocks", mock.Anything, uint64(30)).
		After(time.Millisecond * 100).
		Return(uint64(35)).Once()

	go download(ctx1, d, 0, syncBlockChunck, downloadCh)
	for _, expectedBlock := range expectedBlocks {
		actualBlock := <-downloadCh
		log.Debugf("block %d received!", actualBlock.Num)
		require.Equal(t, expectedBlock, actualBlock)
	}
	log.Debug("canceling")
	cancel()
	_, ok := <-downloadCh
	require.False(t, ok)
}

// func TestWaitForNewBlocks(t *testing.T) {
// 	clientMock := NewL2Mock(t)
// 	clientMock.Mock.
// }

// func TestGetBlockHeader(t *testing.T) {
// 	clientMock := NewL2Mock(t)
// 	clientMock.Mock.
// }
