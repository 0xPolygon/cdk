package sync

import (
	"context"
	"errors"
	"fmt"
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
	d, clientMock := NewTestDownloader(t, time.Millisecond*100)

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
	blocksC2 := []*EVMBlock{
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

func TestWaitForNewBlocks(t *testing.T) {
	ctx := context.Background()
	d, clientMock := NewTestDownloader(t, time.Millisecond*100)

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
	d, clientMock := NewTestDownloader(t, time.Millisecond)

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

	// header not found default
	clientMock.On("HeaderByNumber", ctx, blockNumBig).Return(nil, ethereum.NotFound).Once()
	clientMock.On("HeaderByNumber", ctx, blockNumBig).Return(returnedBlock, nil).Once()
	actualBlock, isCanceled = d.GetBlockHeader(ctx, 5)
	assert.Equal(t, expectedBlock, actualBlock)
	assert.False(t, isCanceled)

	// header not found default TO
	d, clientMock = NewTestDownloader(t, 0)
	clientMock.On("HeaderByNumber", ctx, blockNumBig).Return(nil, ethereum.NotFound).Once()
	clientMock.On("HeaderByNumber", ctx, blockNumBig).Return(returnedBlock, nil).Once()
	actualBlock, isCanceled = d.GetBlockHeader(ctx, 5)
	assert.Equal(t, expectedBlock, actualBlock)
	assert.False(t, isCanceled)
}

func TestFilterQueryToString(t *testing.T) {
	addr1 := common.HexToAddress("0xf000")
	addr2 := common.HexToAddress("0xabcd")
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(1000),
		Addresses: []common.Address{addr1, addr2},
		ToBlock:   new(big.Int).SetUint64(1100),
	}

	assert.Equal(t, "FromBlock: 1000, ToBlock: 1100, Addresses: [0x000000000000000000000000000000000000f000 0x000000000000000000000000000000000000ABcD], Topics: []", filterQueryToString(query))

	query = ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(1000),
		Addresses: []common.Address{addr1, addr2},
		ToBlock:   new(big.Int).SetUint64(1100),
		Topics:    [][]common.Hash{{common.HexToHash("0x1234"), common.HexToHash("0x5678")}},
	}
	assert.Equal(t, "FromBlock: 1000, ToBlock: 1100, Addresses: [0x000000000000000000000000000000000000f000 0x000000000000000000000000000000000000ABcD], Topics: [[0x0000000000000000000000000000000000000000000000000000000000001234 0x0000000000000000000000000000000000000000000000000000000000005678]]", filterQueryToString(query))
}

func TestGetLogs(t *testing.T) {
	mockEthClient := NewL2Mock(t)
	sut := EVMDownloaderImplementation{
		ethClient:        mockEthClient,
		adressessToQuery: []common.Address{contractAddr},
		log:              log.WithFields("test", "EVMDownloaderImplementation"),
		rh: &RetryHandler{
			RetryAfterErrorPeriod:      time.Millisecond,
			MaxRetryAttemptsAfterError: 5,
		},
	}
	ctx := context.TODO()
	mockEthClient.EXPECT().FilterLogs(ctx, mock.Anything).Return(nil, errors.New("foo")).Once()
	mockEthClient.EXPECT().FilterLogs(ctx, mock.Anything).Return(nil, nil).Once()
	logs := sut.GetLogs(ctx, 0, 1)
	require.Equal(t, []types.Log{}, logs)
}

func TestDownloadBeforeFinalized(t *testing.T) {
	mockEthDownloader := NewEVMDownloaderMock(t)

	ctx := context.Background()
	ctx1, cancel := context.WithCancel(ctx)
	defer cancel()

	downloader, _ := NewTestDownloader(t, time.Millisecond)
	downloader.EVMDownloaderInterface = mockEthDownloader

	steps := []struct {
		finalizedBlock          uint64
		fromBlock, toBlock      uint64
		eventsReponse           EVMBlocks
		waitForNewBlocks        bool
		waitForNewBlocksRequest uint64
		waitForNewBlockReply    uint64
		getBlockHeader          *EVMBlockHeader
	}{
		{finalizedBlock: 33, fromBlock: 1, toBlock: 11, waitForNewBlocks: true, waitForNewBlocksRequest: 0, waitForNewBlockReply: 35, getBlockHeader: &EVMBlockHeader{Num: 11}},
		{finalizedBlock: 33, fromBlock: 12, toBlock: 22, eventsReponse: EVMBlocks{createEVMBlock(t, 14, true)}, getBlockHeader: &EVMBlockHeader{Num: 22}},
		// It returns the last block of range, so it don't need to create a empty one
		{finalizedBlock: 33, fromBlock: 23, toBlock: 33, eventsReponse: EVMBlocks{createEVMBlock(t, 33, true)}},
		// It reach the top of chain (block 35)
		{finalizedBlock: 33, fromBlock: 34, toBlock: 35},
		// Previous iteration we reach top of chain so we need update the latest block
		{finalizedBlock: 33, fromBlock: 34, toBlock: 54, waitForNewBlocks: true, waitForNewBlocksRequest: 35, waitForNewBlockReply: 60},
		// finalized block is 35, so we can reduce emit an emptyBlock and reduce the range
		{finalizedBlock: 35, fromBlock: 34, toBlock: 60, getBlockHeader: &EVMBlockHeader{Num: 35}},
		{finalizedBlock: 35, fromBlock: 36, toBlock: 46},
		{finalizedBlock: 35, fromBlock: 36, toBlock: 56, eventsReponse: EVMBlocks{createEVMBlock(t, 36, false)}},
		// Block 36 is the new last block,so it reduce the range again to [37-47]
		{finalizedBlock: 35, fromBlock: 37, toBlock: 47},
		{finalizedBlock: 57, fromBlock: 37, toBlock: 57, eventsReponse: EVMBlocks{createEVMBlock(t, 57, false)}},
		{finalizedBlock: 61, fromBlock: 58, toBlock: 60, eventsReponse: EVMBlocks{createEVMBlock(t, 60, false)}},
		{finalizedBlock: 61, fromBlock: 61, toBlock: 61, waitForNewBlocks: true, waitForNewBlocksRequest: 60, waitForNewBlockReply: 61, getBlockHeader: &EVMBlockHeader{Num: 61}},
		{finalizedBlock: 61, fromBlock: 62, toBlock: 62, waitForNewBlocks: true, waitForNewBlocksRequest: 61, waitForNewBlockReply: 62},
	}
	for i := 0; i < len(steps); i++ {
		log.Info("iteration: ", i, "------------------------------------------------")
		downloadCh := make(chan EVMBlock, 100)
		downloader, _ := NewTestDownloader(t, time.Millisecond)
		downloader.EVMDownloaderInterface = mockEthDownloader
		downloader.setStopDownloaderOnIterationN(i + 1)
		expectedBlocks := EVMBlocks{}
		for _, step := range steps[:i+1] {
			mockEthDownloader.On("GetLastFinalizedBlock", mock.Anything).Return(&types.Header{Number: big.NewInt(int64(step.finalizedBlock))}, nil).Once()
			if step.waitForNewBlocks {
				mockEthDownloader.On("WaitForNewBlocks", mock.Anything, step.waitForNewBlocksRequest).Return(step.waitForNewBlockReply).Once()
			}
			mockEthDownloader.On("GetEventsByBlockRange", mock.Anything, step.fromBlock, step.toBlock).
				Return(step.eventsReponse, false).Once()
			for _, eventBlock := range step.eventsReponse {
				expectedBlocks = append(expectedBlocks, eventBlock)
			}
			if step.getBlockHeader != nil {
				log.Infof("iteration:%d : GetBlockHeader(%d) ", i, step.getBlockHeader.Num)
				mockEthDownloader.On("GetBlockHeader", mock.Anything, step.getBlockHeader.Num).Return(*step.getBlockHeader, false).Once()
				expectedBlocks = append(expectedBlocks, &EVMBlock{
					EVMBlockHeader:   *step.getBlockHeader,
					IsFinalizedBlock: step.getBlockHeader.Num <= step.finalizedBlock,
				})
			}
		}
		downloader.Download(ctx1, 1, downloadCh)
		mockEthDownloader.AssertExpectations(t)
		for _, expectedBlock := range expectedBlocks {
			log.Debugf("waiting block %d ", expectedBlock.Num)
			actualBlock := <-downloadCh
			log.Debugf("block %d received!", actualBlock.Num)
			require.Equal(t, *expectedBlock, actualBlock)
		}
	}
}

func buildAppender() LogAppenderMap {
	appender := make(LogAppenderMap)
	appender[eventSignature] = func(b *EVMBlock, l types.Log) error {
		b.Events = append(b.Events, testEvent(l.Topics[1]))
		return nil
	}
	return appender
}

func NewTestDownloader(t *testing.T, retryPeriod time.Duration) (*EVMDownloader, *L2Mock) {
	t.Helper()

	rh := &RetryHandler{
		MaxRetryAttemptsAfterError: 5,
		RetryAfterErrorPeriod:      retryPeriod,
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

func createEVMBlock(t *testing.T, num uint64, isSafeBlock bool) *EVMBlock {
	t.Helper()
	return &EVMBlock{
		IsFinalizedBlock: isSafeBlock,
		EVMBlockHeader: EVMBlockHeader{
			Num:        num,
			Hash:       common.HexToHash(fmt.Sprintf("0x%.2X", num)),
			ParentHash: common.HexToHash(fmt.Sprintf("0x%.2X", num-1)),
			Timestamp:  uint64(time.Now().Unix()),
		},
	}
}
