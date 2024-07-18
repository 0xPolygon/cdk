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

	if err != nil {
		tb.Fatal(err)
	}

	db, err := mdbx.NewMDBX(nil).
		Path(dir).
		WithTableCfg(tableCfgFunc).
		Open()
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			tb.Fatal(err)
		}
	})

	return db
}

func TestBlockMap(t *testing.T) {
	// Create a new block map
	bm := newBlockMap(
		block{Num: 1, Hash: common.HexToHash("0x123")},
		block{Num: 2, Hash: common.HexToHash("0x456")},
		block{Num: 3, Hash: common.HexToHash("0x789")},
	)

	// Test getSorted function
	sortedBlocks := bm.getSorted()
	expectedSortedBlocks := []block{
		{Num: 1, Hash: common.HexToHash("0x123")},
		{Num: 2, Hash: common.HexToHash("0x456")},
		{Num: 3, Hash: common.HexToHash("0x789")},
	}
	if !reflect.DeepEqual(sortedBlocks, expectedSortedBlocks) {
		t.Errorf("getSorted() returned incorrect result, expected: %v, got: %v", expectedSortedBlocks, sortedBlocks)
	}

	// Test getFromBlockSorted function
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

		unfinalisedBlocksMap, exists := rd.trackedBlocks[unfalisedBlocksID]
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

		insertTestData(t, ctx, db, unfinalisedBlocks, unfalisedBlocksID)
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

		unfinalisedBlocksMap, exists := rd.trackedBlocks[unfalisedBlocksID]
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

		insertTestData(t, ctx, db, unfinalisedBlocks, unfalisedBlocksID)
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

		unfinalisedBlocksMap, exists := rd.trackedBlocks[unfalisedBlocksID]
		require.True(t, exists)
		require.Len(t, unfinalisedBlocksMap, len(unfinalisedBlocks)-len(testSubscriberBlocks)) // since all blocks are finalized

		testSubscriberMap, exists := rd.trackedBlocks[testSubscriber]
		require.True(t, exists)
		require.Len(t, testSubscriberMap, 0) // since all blocks are finalized
	})
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
