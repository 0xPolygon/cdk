package reorgdetector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	big "math/big"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	"github.com/stretchr/testify/require"
)

const testSubscriber = "testSubscriber"

// newTestDB creates new instance of db used by tests.
func newTestDB(tb testing.TB) kv.RwDB {
	tb.Helper()

	dir := fmt.Sprintf("/tmp/reorgdetector-temp_%v", time.Now().UTC().Format(time.RFC3339Nano))
	err := os.Mkdir(dir, 0775)

	require.NoError(tb, err)

	db, err := mdbx.NewMDBX(nil).
		Path(dir).
		WithTableCfg(tableCfgFunc).
		Open()
	require.NoError(tb, err)

	tb.Cleanup(func() {
		require.NoError(tb, os.RemoveAll(dir))
	})

	return db
}

func TestBlockMap(t *testing.T) {
	t.Parallel()

	// Create a new block map
	bm := newBlockMap(
		block{Num: 1, Hash: common.HexToHash("0x123")},
		block{Num: 2, Hash: common.HexToHash("0x456")},
		block{Num: 3, Hash: common.HexToHash("0x789")},
	)

	t.Run("getSorted", func(t *testing.T) {
		t.Parallel()

		sortedBlocks := bm.getSorted()
		expectedSortedBlocks := []block{
			{Num: 1, Hash: common.HexToHash("0x123")},
			{Num: 2, Hash: common.HexToHash("0x456")},
			{Num: 3, Hash: common.HexToHash("0x789")},
		}
		if !reflect.DeepEqual(sortedBlocks, expectedSortedBlocks) {
			t.Errorf("getSorted() returned incorrect result, expected: %v, got: %v", expectedSortedBlocks, sortedBlocks)
		}
	})

	t.Run("getFromBlockSorted", func(t *testing.T) {
		t.Parallel()

		fromBlockSorted := bm.getFromBlockSorted(2)
		expectedFromBlockSorted := []block{
			{Num: 3, Hash: common.HexToHash("0x789")},
		}
		if !reflect.DeepEqual(fromBlockSorted, expectedFromBlockSorted) {
			t.Errorf("getFromBlockSorted() returned incorrect result, expected: %v, got: %v", expectedFromBlockSorted, fromBlockSorted)
		}

		// Test getFromBlockSorted function when blockNum is greater than the last block
		fromBlockSorted = bm.getFromBlockSorted(4)
		expectedFromBlockSorted = []block{}
		if !reflect.DeepEqual(fromBlockSorted, expectedFromBlockSorted) {
			t.Errorf("getFromBlockSorted() returned incorrect result, expected: %v, got: %v", expectedFromBlockSorted, fromBlockSorted)
		}
	})

	t.Run("getClosestHigherBlock", func(t *testing.T) {
		t.Parallel()

		bm := newBlockMap(
			block{Num: 1, Hash: common.HexToHash("0x123")},
			block{Num: 2, Hash: common.HexToHash("0x456")},
			block{Num: 3, Hash: common.HexToHash("0x789")},
		)

		// Test when the blockNum exists in the block map
		b, exists := bm.getClosestHigherBlock(2)
		require.True(t, exists)
		expectedBlock := block{Num: 2, Hash: common.HexToHash("0x456")}
		if b != expectedBlock {
			t.Errorf("getClosestHigherBlock() returned incorrect result, expected: %v, got: %v", expectedBlock, b)
		}

		// Test when the blockNum does not exist in the block map
		b, exists = bm.getClosestHigherBlock(4)
		require.False(t, exists)
		expectedBlock = block{Num: 0, Hash: common.Hash{}}
		if b != expectedBlock {
			t.Errorf("getClosestHigherBlock() returned incorrect result, expected: %v, got: %v", expectedBlock, b)
		}
	})

	t.Run("removeRange", func(t *testing.T) {
		t.Parallel()

		bm := newBlockMap(
			block{Num: 1, Hash: common.HexToHash("0x123")},
			block{Num: 2, Hash: common.HexToHash("0x456")},
			block{Num: 3, Hash: common.HexToHash("0x789")},
			block{Num: 4, Hash: common.HexToHash("0xabc")},
			block{Num: 5, Hash: common.HexToHash("0xdef")},
		)

		bm.removeRange(3, 5)

		expectedBlocks := []block{
			{Num: 1, Hash: common.HexToHash("0x123")},
			{Num: 2, Hash: common.HexToHash("0x456")},
		}

		sortedBlocks := bm.getSorted()

		if !reflect.DeepEqual(sortedBlocks, expectedBlocks) {
			t.Errorf("removeRange() failed, expected: %v, got: %v", expectedBlocks, sortedBlocks)
		}
	})
}

func TestReorgDetector_New(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("first initialization, no data", func(t *testing.T) {
		t.Parallel()

		client := NewEthClientMock(t)
		db := newTestDB(t)

		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(
			&types.Header{Number: big.NewInt(100)}, nil,
		)

		rd, err := newReorgDetector(ctx, client, db)
		require.NoError(t, err)
		require.Len(t, rd.trackedBlocks, 1)

		unfinalisedBlocksMap, exists := rd.trackedBlocks[unfinalisedBlocksID]
		require.True(t, exists)
		require.Empty(t, unfinalisedBlocksMap)
	})

	t.Run("getting last finalized block failed", func(t *testing.T) {
		t.Parallel()

		client := NewEthClientMock(t)
		db := newTestDB(t)

		expectedErr := errors.New("some error")
		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(nil, expectedErr)

		_, err := newReorgDetector(ctx, client, db)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("have tracked blocks and subscriptions no reorg - all blocks finalized", func(t *testing.T) {
		t.Parallel()

		client := NewEthClientMock(t)
		db := newTestDB(t)

		testBlocks := createTestBlocks(t, 1, 6)
		unfinalisedBlocks := testBlocks[:5]
		testSubscriberBlocks := testBlocks[:3]

		insertTestData(t, ctx, db, unfinalisedBlocks, unfinalisedBlocksID)
		insertTestData(t, ctx, db, testSubscriberBlocks, testSubscriber)

		for _, block := range unfinalisedBlocks {
			client.On("HeaderByNumber", ctx, block.Number).Return(
				block, nil,
			)
		}

		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(
			testBlocks[len(testBlocks)-1], nil,
		)

		rd, err := newReorgDetector(ctx, client, db)
		require.NoError(t, err)
		require.Len(t, rd.trackedBlocks, 2) // testSubscriber and unfinalisedBlocks

		unfinalisedBlocksMap, exists := rd.trackedBlocks[unfinalisedBlocksID]
		require.True(t, exists)
		require.Len(t, unfinalisedBlocksMap, 0) // since all blocks are finalized

		testSubscriberMap, exists := rd.trackedBlocks[testSubscriber]
		require.True(t, exists)
		require.Len(t, testSubscriberMap, 0) // since all blocks are finalized
	})

	t.Run("have tracked blocks and subscriptions no reorg - not all blocks finalized", func(t *testing.T) {
		t.Parallel()

		client := NewEthClientMock(t)
		db := newTestDB(t)

		testBlocks := createTestBlocks(t, 1, 7)
		unfinalisedBlocks := testBlocks[:6]
		testSubscriberBlocks := testBlocks[:4]

		insertTestData(t, ctx, db, unfinalisedBlocks, unfinalisedBlocksID)
		insertTestData(t, ctx, db, testSubscriberBlocks, testSubscriber)

		for _, block := range unfinalisedBlocks {
			client.On("HeaderByNumber", ctx, block.Number).Return(
				block, nil,
			)
		}

		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(
			testSubscriberBlocks[len(testSubscriberBlocks)-1], nil,
		)

		rd, err := newReorgDetector(ctx, client, db)
		require.NoError(t, err)
		require.Len(t, rd.trackedBlocks, 2) // testSubscriber and unfinalisedBlocks

		unfinalisedBlocksMap, exists := rd.trackedBlocks[unfinalisedBlocksID]
		require.True(t, exists)
		require.Len(t, unfinalisedBlocksMap, len(unfinalisedBlocks)-len(testSubscriberBlocks)) // since all blocks are finalized

		testSubscriberMap, exists := rd.trackedBlocks[testSubscriber]
		require.True(t, exists)
		require.Len(t, testSubscriberMap, 0) // since all blocks are finalized
	})

	t.Run("have tracked blocks and subscriptions - reorg happened", func(t *testing.T) {
		t.Parallel()

		client := NewEthClientMock(t)
		db := newTestDB(t)

		trackedBlocks := createTestBlocks(t, 1, 5)
		testSubscriberBlocks := trackedBlocks[:5]

		insertTestData(t, ctx, db, nil, unfinalisedBlocksID) // no unfinalised blocks
		insertTestData(t, ctx, db, testSubscriberBlocks, testSubscriber)

		for _, block := range trackedBlocks[:3] {
			client.On("HeaderByNumber", ctx, block.Number).Return(
				block, nil,
			)
		}

		reorgedBlocks := createTestBlocks(t, 4, 2)            // block 4, and 5 are reorged
		reorgedBlocks[0].ParentHash = trackedBlocks[2].Hash() // block 4 is reorged but his parent is block 3
		reorgedBlocks[1].ParentHash = reorgedBlocks[0].Hash() // block 5 is reorged but his parent is block 4

		client.On("HeaderByNumber", ctx, reorgedBlocks[0].Number).Return(
			reorgedBlocks[0], nil,
		)

		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(
			reorgedBlocks[len(reorgedBlocks)-1], nil,
		)

		rd, err := newReorgDetector(ctx, client, db)
		require.NoError(t, err)
		require.Len(t, rd.trackedBlocks, 2) // testSubscriber and unfinalisedBlocks

		// we wait for the subscriber to be notified about the reorg
		firstReorgedBlock := <-rd.subscriptions[testSubscriber].FirstReorgedBlock
		require.Equal(t, reorgedBlocks[0].Number.Uint64(), firstReorgedBlock)

		// all blocks should be cleaned from the tracked blocks
		// since subscriber had 5 blocks, 3 were finalized, and 2 were reorged but also finalized
		subscriberBlocks := rd.trackedBlocks[testSubscriber]
		require.Len(t, subscriberBlocks, 0)
	})
}

func TestReorgDetector_AddBlockToTrack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("no subscription", func(t *testing.T) {
		t.Parallel()

		client := NewEthClientMock(t)
		db := newTestDB(t)

		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(
			&types.Header{Number: big.NewInt(10)}, nil,
		)

		rd, err := newReorgDetector(ctx, client, db)
		require.NoError(t, err)

		err = rd.AddBlockToTrack(ctx, testSubscriber, 1, common.HexToHash("0x123"))
		require.ErrorIs(t, err, ErrNotSubscribed)
	})

	t.Run("no unfinalised blocks - block not finalised", func(t *testing.T) {
		t.Parallel()

		client := NewEthClientMock(t)
		db := newTestDB(t)

		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(
			&types.Header{Number: big.NewInt(10)}, nil,
		).Once()

		rd, err := newReorgDetector(ctx, client, db)
		require.NoError(t, err)

		_, err = rd.Subscribe(testSubscriber)
		require.NoError(t, err)

		err = rd.AddBlockToTrack(ctx, testSubscriber, 11, common.HexToHash("0x123")) // block not finalized
		require.NoError(t, err)

		subBlocks := rd.trackedBlocks[testSubscriber]
		require.Len(t, subBlocks, 1)
		require.Equal(t, subBlocks[11].Hash, common.HexToHash("0x123"))
	})

	t.Run("have unfinalised blocks - block not finalized", func(t *testing.T) {
		t.Parallel()

		client := NewEthClientMock(t)
		db := newTestDB(t)

		unfinalisedBlocks := createTestBlocks(t, 11, 5)
		insertTestData(t, ctx, db, unfinalisedBlocks, unfinalisedBlocksID)

		for _, block := range unfinalisedBlocks {
			client.On("HeaderByNumber", ctx, block.Number).Return(
				block, nil,
			)
		}

		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(
			&types.Header{Number: big.NewInt(10)}, nil,
		).Once()

		rd, err := newReorgDetector(ctx, client, db)
		require.NoError(t, err)

		_, err = rd.Subscribe(testSubscriber)
		require.NoError(t, err)

		err = rd.AddBlockToTrack(ctx, testSubscriber, 11, unfinalisedBlocks[0].Hash()) // block not finalized
		require.NoError(t, err)

		subBlocks := rd.trackedBlocks[testSubscriber]
		require.Len(t, subBlocks, 1)
		require.Equal(t, subBlocks[11].Hash, unfinalisedBlocks[0].Hash())
	})
}

func TestReorgDetector_removeFinalisedBlocks(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	client := NewEthClientMock(t)
	db := newTestDB(t)

	unfinalisedBlocks := createTestBlocks(t, 1, 10)
	insertTestData(t, ctx, db, unfinalisedBlocks, unfinalisedBlocksID)
	insertTestData(t, ctx, db, unfinalisedBlocks, testSubscriber)

	// call for removeFinalisedBlocks
	client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(
		&types.Header{Number: big.NewInt(5)}, nil,
	)

	rd := &ReorgDetector{
		ethClient:              client,
		db:                     db,
		trackedBlocks:          make(map[string]blockMap),
		waitPeriodBlockRemover: 100 * time.Millisecond,
		waitPeriodBlockAdder:   100 * time.Millisecond,
		subscriptions: map[string]*Subscription{
			testSubscriber: {
				FirstReorgedBlock: make(chan uint64),
				ReorgProcessed:    make(chan bool),
			},
			unfinalisedBlocksID: {
				FirstReorgedBlock: make(chan uint64),
				ReorgProcessed:    make(chan bool),
			},
		},
	}

	trackedBlocks, err := rd.getTrackedBlocks(ctx)
	require.NoError(t, err)
	require.Len(t, trackedBlocks, 2)

	rd.trackedBlocks = trackedBlocks

	// make sure we have all blocks in the tracked blocks before removing finalized blocks
	require.Len(t, rd.trackedBlocks[unfinalisedBlocksID], len(unfinalisedBlocks))
	require.Len(t, rd.trackedBlocks[testSubscriber], len(unfinalisedBlocks))

	// remove finalized blocks
	go rd.removeFinalisedBlocks(ctx)

	time.Sleep(3 * time.Second) // wait for the go routine to remove the finalized blocks
	cancel()

	// make sure all blocks are removed from the tracked blocks
	rd.trackedBlocksLock.RLock()
	defer rd.trackedBlocksLock.RUnlock()

	require.Len(t, rd.trackedBlocks[unfinalisedBlocksID], 5)
	require.Len(t, rd.trackedBlocks[testSubscriber], 5)
}

func createTestBlocks(t *testing.T, startBlock uint64, count uint64) []*types.Header {
	t.Helper()

	blocks := make([]*types.Header, 0, count)
	for i := startBlock; i < startBlock+count; i++ {
		blocks = append(blocks, &types.Header{Number: big.NewInt(int64(i))})
	}

	return blocks
}

func insertTestData(t *testing.T, ctx context.Context, db kv.RwDB, blocks []*types.Header, id string) {
	t.Helper()

	// Insert some test data
	err := db.Update(ctx, func(tx kv.RwTx) error {

		blockMap := newBlockMap()
		for _, b := range blocks {
			blockMap[b.Number.Uint64()] = block{b.Number.Uint64(), b.Hash()}
		}

		raw, err := json.Marshal(blockMap.getSorted())
		if err != nil {
			return err
		}

		return tx.Put(subscriberBlocks, []byte(id), raw)
	})

	require.NoError(t, err)
}
