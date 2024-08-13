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
		expectedBlocks     []EVMBlock
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
		expectedBlocks: []EVMBlock{},
	}
	testCases = append(testCases, case0)

	// case 1: single block, single event
	logC1, updateC1 := generateEvent(3)
	logsC1 := []types.Log{
		*logC1,
	}
	blocksC1 := []EVMBlock{
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
	blocksC2 := []EVMBlock{
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
	blocksC3 := []EVMBlock{
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
	expectedBlocks := []EVMBlock{}
	dwnldr, _ := NewTestDownloader(t)
	dwnldr.EVMDownloaderInterface = d

	d.On("WaitForNewBlocks", mock.Anything, uint64(0)).
		Return(uint64(1))
	// iteratiion 0:
	// last block is 1, download that block (no events and wait)
	b1 := EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  1,
			Hash: common.HexToHash("01"),
		},
	}
	expectedBlocks = append(expectedBlocks, b1)
	d.On("GetEventsByBlockRange", mock.Anything, uint64(0), uint64(1)).
		Return([]EVMBlock{})
	d.On("GetBlockHeader", mock.Anything, uint64(1)).
		Return(b1.EVMBlockHeader)

	// iteration 1: wait for next block to be created
	d.On("WaitForNewBlocks", mock.Anything, uint64(1)).
		After(time.Millisecond * 100).
		Return(uint64(2)).Once()

	// iteration 2: block 2 has events
	b2 := EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  2,
			Hash: common.HexToHash("02"),
		},
	}
	expectedBlocks = append(expectedBlocks, b2)
	d.On("GetEventsByBlockRange", mock.Anything, uint64(2), uint64(2)).
		Return([]EVMBlock{b2})

	// iteration 3: wait for next block to be created (jump to block 8)
	d.On("WaitForNewBlocks", mock.Anything, uint64(2)).
		After(time.Millisecond * 100).
		Return(uint64(8)).Once()

	// iteration 4: blocks 6 and 7 have events
	b6 := EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  6,
			Hash: common.HexToHash("06"),
		},
		Events: []interface{}{"06"},
	}
	b7 := EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  7,
			Hash: common.HexToHash("07"),
		},
		Events: []interface{}{"07"},
	}
	b8 := EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  8,
			Hash: common.HexToHash("08"),
		},
	}
	expectedBlocks = append(expectedBlocks, b6, b7, b8)
	d.On("GetEventsByBlockRange", mock.Anything, uint64(3), uint64(8)).
		Return([]EVMBlock{b6, b7})
	d.On("GetBlockHeader", mock.Anything, uint64(8)).
		Return(b8.EVMBlockHeader)

	// iteration 5: wait for next block to be created (jump to block 30)
	d.On("WaitForNewBlocks", mock.Anything, uint64(8)).
		After(time.Millisecond * 100).
		Return(uint64(30)).Once()

	// iteration 6: from block 9 to 19, no events
	b19 := EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  19,
			Hash: common.HexToHash("19"),
		},
	}
	expectedBlocks = append(expectedBlocks, b19)
	d.On("GetEventsByBlockRange", mock.Anything, uint64(9), uint64(19)).
		Return([]EVMBlock{})
	d.On("GetBlockHeader", mock.Anything, uint64(19)).
		Return(b19.EVMBlockHeader)

	// iteration 7: from block 20 to 30, events on last block
	b30 := EVMBlock{
		EVMBlockHeader: EVMBlockHeader{
			Num:  30,
			Hash: common.HexToHash("30"),
		},
		Events: []interface{}{testEvent(common.HexToHash("30"))},
	}
	expectedBlocks = append(expectedBlocks, b30)
	d.On("GetEventsByBlockRange", mock.Anything, uint64(20), uint64(30)).
		Return([]EVMBlock{b30})

	// iteration 8: wait for next block to be created (jump to block 35)
	d.On("WaitForNewBlocks", mock.Anything, uint64(30)).
		After(time.Millisecond * 100).
		Return(uint64(35)).Once()

	go dwnldr.Download(ctx1, 0, downloadCh)
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
	actualBlock := d.GetBlockHeader(ctx, blockNum)
	assert.Equal(t, expectedBlock, actualBlock)

	// after error from client
	clientMock.On("HeaderByNumber", ctx, blockNumBig).Return(nil, errors.New("foo")).Once()
	clientMock.On("HeaderByNumber", ctx, blockNumBig).Return(returnedBlock, nil).Once()
	actualBlock = d.GetBlockHeader(ctx, blockNum)
	assert.Equal(t, expectedBlock, actualBlock)
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
	rh := &RetryHandler{
		MaxRetryAttemptsAfterError: 5,
		RetryAfterErrorPeriod:      time.Millisecond * 100,
	}
	clientMock := NewL2Mock(t)
	d, err := NewEVMDownloader(clientMock, syncBlockChunck, etherman.LatestBlock, time.Millisecond, buildAppender(), []common.Address{contractAddr}, rh)
	require.NoError(t, err)
	return d, clientMock
}
