package sync

import (
	"context"
	"errors"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	contractAddr   = common.HexToAddress("f00")
	eventSignature = crypto.Keccak256Hash([]byte("foo"))
)

const (
	syncBlockChunck = uint64(10)
)

type testEvent common.Hash

func TestGetEventsByBlockRange(t *testing.T) {
	type testCase struct {
		description        string
		inputLogs          []types.Log
		fromBlock, toBlock uint64
		expectedBlocks     EVMBlocks
	}
	testCases := []testCase{}
	ctx := context.Background()
	d, clientMock := NewTestDownloader(t)

	// case 0: single block, no events
	case0 := testCase{
		description:    "case 0: single block, no events",
		inputLogs:      []types.Log{},
		fromBlock:      1,
		toBlock:        3,
		expectedBlocks: EVMBlocks{},
	}
	testCases = append(testCases, case0)

	// case 1: single block, single event
	logC1, updateC1 := generateEvent(3)
	logsC1 := []types.Log{
		*logC1,
	}
	blocksC1 := EVMBlocks{
		{
			EVMBlockHeader: EVMBlockHeader{
				Num:        logC1.BlockNumber,
				Hash:       logC1.BlockHash,
				ParentHash: common.HexToHash("foo"),
			},
			Events: []interface{}{updateC1},
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
	logC2_1, updateC2_1 := generateEvent(5)
	logC2_2, updateC2_2 := generateEvent(5)
	logC2_3, updateC2_3 := generateEvent(5)
	logC2_4, updateC2_4 := generateEvent(5)
	logsC2 := []types.Log{
		*logC2_1,
		*logC2_2,
		*logC2_3,
		*logC2_4,
	}
	blocksC2 := EVMBlocks{
		{
			EVMBlockHeader: EVMBlockHeader{
				Num:        logC2_1.BlockNumber,
				Hash:       logC2_1.BlockHash,
				ParentHash: common.HexToHash("foo"),
			},
			Events: []interface{}{
				updateC2_1,
				updateC2_2,
				updateC2_3,
				updateC2_4,
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
	logC3_1, updateC3_1 := generateEvent(7)
	logC3_2, updateC3_2 := generateEvent(7)
	logC3_3, updateC3_3 := generateEvent(8)
	logC3_4, updateC3_4 := generateEvent(8)
	logsC3 := []types.Log{
		*logC3_1,
		*logC3_2,
		*logC3_3,
		*logC3_4,
	}
	blocksC3 := EVMBlocks{
		{
			EVMBlockHeader: EVMBlockHeader{
				Num:        logC3_1.BlockNumber,
				Hash:       logC3_1.BlockHash,
				ParentHash: common.HexToHash("foo"),
			},
			Events: []interface{}{
				updateC3_1,
				updateC3_2,
			},
		},
		{
			EVMBlockHeader: EVMBlockHeader{
				Num:        logC3_3.BlockNumber,
				Hash:       logC3_3.BlockHash,
				ParentHash: common.HexToHash("foo"),
			},
			Events: []interface{}{
				updateC3_3,
				updateC3_4,
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
			Addresses: []common.Address{contractAddr},
			ToBlock:   new(big.Int).SetUint64(tc.toBlock),
		}
		clientMock.
			On("FilterLogs", mock.Anything, query).
			Return(tc.inputLogs, nil)
		for _, b := range tc.expectedBlocks {
			clientMock.
				On("HeaderByNumber", mock.Anything, big.NewInt(int64(b.Num))).
				Return(&types.Header{
					Number:     big.NewInt(int64(b.Num)),
					ParentHash: common.HexToHash("foo"),
				}, nil)
		}

		actualBlocks := d.GetEventsByBlockRange(ctx, tc.fromBlock, tc.toBlock)
		require.Equal(t, tc.expectedBlocks, actualBlocks, tc.description)
	}
}

func generateEvent(blockNum uint32) (*types.Log, testEvent) {
	h := common.HexToHash(strconv.Itoa(int(blockNum)))
	header := types.Header{
		Number:     big.NewInt(int64(blockNum)),
		ParentHash: common.HexToHash("foo"),
	}
	blockHash := header.Hash()
	log := &types.Log{
		Address:     contractAddr,
		BlockNumber: uint64(blockNum),
		Topics: []common.Hash{
			eventSignature,
			h,
		},
		BlockHash: blockHash,
		Data:      nil,
	}
	return log, testEvent(h)
}

func TestDownload(t *testing.T) {
	/*
		NOTE: due to the concurrent nature of this test (the function being tested runs through a goroutine)
		if the mock doesn't match, the goroutine will get stuck and the test will timeout
	*/
	d := NewEVMDownloaderMock(t)
	downloadCh := make(chan EVMBlock, 1)
	ctx := context.Background()
	ctx1, cancel := context.WithCancel(ctx)
	expectedBlocks := EVMBlocks{}
	dwnldr, _ := NewTestDownloader(t)
	dwnldr.EVMDownloaderInterface = d

	d.On("WaitForNewBlocks", mock.Anything, uint64(0)).
		Return(uint64(1))

	lastFinalizedBlock := &types.Header{Number: big.NewInt(1)}
	createEVMBlockFn := func(header *types.Header, isSafeBlock bool) *EVMBlock {
		return &EVMBlock{
			IsSafeBlock: isSafeBlock,
			EVMBlockHeader: EVMBlockHeader{
				Num:        header.Number.Uint64(),
				Hash:       header.Hash(),
				ParentHash: header.ParentHash,
				Timestamp:  header.Time,
			},
		}
	}

	// iteration 0:
	// last block is 1, download that block (no events and wait)
	b0 := createEVMBlockFn(lastFinalizedBlock, true)
	expectedBlocks = append(expectedBlocks, b0)
	d.On("GetLastFinalizedBlock", mock.Anything).Return(lastFinalizedBlock, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(0), uint64(1)).
		Return(EVMBlocks{}, false).Once()
	d.On("GetBlockHeader", mock.Anything, uint64(1)).Return(b0.EVMBlockHeader, false).Once()

	// iteration 1: we have a new block, so increase to block (no events)
	lastFinalizedBlock = &types.Header{Number: big.NewInt(2)}
	b2 := createEVMBlockFn(lastFinalizedBlock, true)
	expectedBlocks = append(expectedBlocks, b2)
	d.On("WaitForNewBlocks", mock.Anything, uint64(1)).
		Return(uint64(2))
	d.On("GetLastFinalizedBlock", mock.Anything).Return(lastFinalizedBlock, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(2), uint64(2)).
		Return(EVMBlocks{}, false).Once()
	d.On("GetBlockHeader", mock.Anything, uint64(2)).Return(b2.EVMBlockHeader, false).Once()

	// iteration 2: wait for next block to be created (jump to block 8)
	d.On("WaitForNewBlocks", mock.Anything, uint64(2)).
		After(time.Millisecond * 100).
		Return(uint64(8)).Once()

	// iteration 3: blocks 6 and 7 have events, last finalized block is 5
	lastFinalizedBlock = &types.Header{Number: big.NewInt(5)}
	b6 := &EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  6,
			Hash: common.HexToHash("06"),
		},
		Events: []interface{}{"06"},
	}
	b7 := &EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  7,
			Hash: common.HexToHash("07"),
		},
		Events: []interface{}{"07"},
	}
	expectedBlocks = append(expectedBlocks, b6, b7)
	d.On("GetLastFinalizedBlock", mock.Anything).Return(lastFinalizedBlock, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(3), uint64(8)).
		Return(EVMBlocks{b6, b7}, false)

	// iteration 4: finalized block is now block 8, report the finalized block
	lastFinalizedBlock = &types.Header{Number: big.NewInt(8)}
	b8 := createEVMBlockFn(lastFinalizedBlock, true)
	expectedBlocks = append(expectedBlocks, b8)
	d.On("GetLastFinalizedBlock", mock.Anything).Return(lastFinalizedBlock, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(8), uint64(8)).
		Return(EVMBlocks{}, false)
	d.On("GetBlockHeader", mock.Anything, uint64(8)).Return(b8.EVMBlockHeader, false).Once()

	// iteration 5: from block 9 to 19, no events
	lastFinalizedBlock = &types.Header{Number: big.NewInt(15)}
	d.On("WaitForNewBlocks", mock.Anything, uint64(8)).
		After(time.Millisecond * 100).
		Return(uint64(19)).Once()
	d.On("GetLastFinalizedBlock", mock.Anything).Return(lastFinalizedBlock, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(9), uint64(19)).
		Return(EVMBlocks{}, false)

	// iteration 6: last finalized block is now 20, no events, report empty block
	d.On("GetLastFinalizedBlock", mock.Anything).Return(&types.Header{Number: big.NewInt(20)}, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(9), uint64(19)).
		Return(EVMBlocks{}, false)

	d.On("WaitForNewBlocks", mock.Anything, uint64(19)).
		After(time.Millisecond * 100).
		Return(uint64(20)).Once()
	b19 := createEVMBlockFn(&types.Header{Number: big.NewInt(19)}, true)
	expectedBlocks = append(expectedBlocks, b19)
	d.On("GetBlockHeader", mock.Anything, uint64(19)).Return(b19.EVMBlockHeader, false) // reporting empty finalized to block

	// iteration 8: last finalized block is 21, no events
	b20 := createEVMBlockFn(&types.Header{Number: big.NewInt(20)}, true)
	expectedBlocks = append(expectedBlocks, b20)
	d.On("GetLastFinalizedBlock", mock.Anything).Return(&types.Header{Number: big.NewInt(21)}, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(20), uint64(20)).
		Return(EVMBlocks{}, false)
	d.On("GetBlockHeader", mock.Anything, uint64(20)).Return(b20.EVMBlockHeader, false) // reporting empty finalized to block

	// iteration 9: last finalized block is 22, no events
	d.On("WaitForNewBlocks", mock.Anything, uint64(20)).
		After(time.Millisecond * 100).
		Return(uint64(21)).Once()
	b21 := createEVMBlockFn(&types.Header{Number: big.NewInt(21)}, true)
	expectedBlocks = append(expectedBlocks, b21)
	d.On("GetLastFinalizedBlock", mock.Anything).Return(&types.Header{Number: big.NewInt(22)}, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(21), uint64(21)).
		Return(EVMBlocks{}, false)
	d.On("GetBlockHeader", mock.Anything, uint64(21)).Return(b21.EVMBlockHeader, false) // reporting empty finalized to block

	// iteration 10: last finalized block is 23, no events
	d.On("WaitForNewBlocks", mock.Anything, uint64(21)).
		After(time.Millisecond * 100).
		Return(uint64(22)).Once()
	b22 := createEVMBlockFn(&types.Header{Number: big.NewInt(22)}, true)
	expectedBlocks = append(expectedBlocks, b22)
	d.On("GetLastFinalizedBlock", mock.Anything).Return(&types.Header{Number: big.NewInt(23)}, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(22), uint64(22)).
		Return(EVMBlocks{}, false)
	d.On("GetBlockHeader", mock.Anything, uint64(22)).Return(b22.EVMBlockHeader, false) // reporting empty finalized to block

	// iteration 11: last finalized block is still 23, no events
	d.On("WaitForNewBlocks", mock.Anything, uint64(22)).
		After(time.Millisecond * 100).
		Return(uint64(23)).Once()
	b23 := createEVMBlockFn(&types.Header{Number: big.NewInt(23)}, true)
	expectedBlocks = append(expectedBlocks, b23)
	d.On("GetLastFinalizedBlock", mock.Anything).Return(&types.Header{Number: big.NewInt(23)}, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(23), uint64(23)).
		Return(EVMBlocks{}, false)
	d.On("GetBlockHeader", mock.Anything, uint64(23)).Return(b23.EVMBlockHeader, false) // reporting empty finalized to block

	// iteration 12: finalized block is 24, has events
	d.On("WaitForNewBlocks", mock.Anything, uint64(23)).
		After(time.Millisecond * 100).
		Return(uint64(24)).Once()
	b24 := &EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  24,
			Hash: common.HexToHash("24"),
		},
		Events: []interface{}{testEvent(common.HexToHash("24"))},
	}
	expectedBlocks = append(expectedBlocks, b24)
	d.On("GetLastFinalizedBlock", mock.Anything).Return(&types.Header{Number: big.NewInt(24)}, nil).Once()
	d.On("GetEventsByBlockRange", mock.Anything, uint64(24), uint64(24)).
		Return(EVMBlocks{b24}, false)

	// iteration 13: closing the downloader
	d.On("WaitForNewBlocks", mock.Anything, uint64(24)).Return(uint64(25)).After(time.Millisecond * 100).Once()

	go dwnldr.Download(ctx1, 0, downloadCh)
	for _, expectedBlock := range expectedBlocks {
		actualBlock := <-downloadCh
		log.Debugf("block %d received!", actualBlock.Num)
		require.Equal(t, *expectedBlock, actualBlock)
	}
	log.Debug("canceling")
	cancel()
	_, ok := <-downloadCh
	require.False(t, ok)
}

func TestWaitForNewBlocks(t *testing.T) {
	ctx := context.Background()
	d, clientMock := NewTestDownloader(t)

	// at first attempt
	currentBlock := uint64(5)
	expectedBlock := uint64(6)
	clientMock.On("HeaderByNumber", ctx, mock.Anything).Return(&types.Header{
		Number: big.NewInt(6),
	}, nil).Once()
	actualBlock := d.WaitForNewBlocks(ctx, currentBlock)
	assert.Equal(t, expectedBlock, actualBlock)

	// 2 iterations
	clientMock.On("HeaderByNumber", ctx, mock.Anything).Return(&types.Header{
		Number: big.NewInt(5),
	}, nil).Once()
	clientMock.On("HeaderByNumber", ctx, mock.Anything).Return(&types.Header{
		Number: big.NewInt(6),
	}, nil).Once()
	actualBlock = d.WaitForNewBlocks(ctx, currentBlock)
	assert.Equal(t, expectedBlock, actualBlock)

	// after error from client
	clientMock.On("HeaderByNumber", ctx, mock.Anything).Return(nil, errors.New("foo")).Once()
	clientMock.On("HeaderByNumber", ctx, mock.Anything).Return(&types.Header{
		Number: big.NewInt(6),
	}, nil).Once()
	actualBlock = d.WaitForNewBlocks(ctx, currentBlock)
	assert.Equal(t, expectedBlock, actualBlock)
}

func TestGetBlockHeader(t *testing.T) {
	ctx := context.Background()
	d, clientMock := NewTestDownloader(t)

	blockNum := uint64(5)
	blockNumBig := big.NewInt(5)
	returnedBlock := &types.Header{
		Number: blockNumBig,
	}
	expectedBlock := EVMBlockHeader{
		Num:  5,
		Hash: returnedBlock.Hash(),
	}

	// at first attempt
	clientMock.On("HeaderByNumber", ctx, blockNumBig).Return(returnedBlock, nil).Once()
	actualBlock, isCanceled := d.GetBlockHeader(ctx, blockNum)
	assert.Equal(t, expectedBlock, actualBlock)
	assert.False(t, isCanceled)

	// after error from client
	clientMock.On("HeaderByNumber", ctx, blockNumBig).Return(nil, errors.New("foo")).Once()
	clientMock.On("HeaderByNumber", ctx, blockNumBig).Return(returnedBlock, nil).Once()
	actualBlock, isCanceled = d.GetBlockHeader(ctx, blockNum)
	assert.Equal(t, expectedBlock, actualBlock)
	assert.False(t, isCanceled)
}

func buildAppender() LogAppenderMap {
	appender := make(LogAppenderMap)
	appender[eventSignature] = func(b *EVMBlock, l types.Log) error {
		b.Events = append(b.Events, testEvent(l.Topics[1]))
		return nil
	}
	return appender
}

func NewTestDownloader(t *testing.T) (*EVMDownloader, *L2Mock) {
	t.Helper()

	rh := &RetryHandler{
		MaxRetryAttemptsAfterError: 5,
		RetryAfterErrorPeriod:      time.Millisecond * 100,
	}
	clientMock := NewL2Mock(t)
	d, err := NewEVMDownloader("test",
		clientMock, syncBlockChunck, etherman.LatestBlock, time.Millisecond,
		buildAppender(), []common.Address{contractAddr}, rh,
		etherman.FinalizedBlock,
	)
	require.NoError(t, err)
	return d, clientMock
}
