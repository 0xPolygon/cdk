package reorgdetector

import (
	"context"
	big "math/big"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	cdktypes "github.com/0xPolygon/cdk/config/types"
	common "github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func Test_ReorgDetector(t *testing.T) {
	const subID = "test"

	ctx := context.Background()

	// Simulated L1
	clientL1 := simulated.NewBackend(nil, simulated.WithBlockGasLimit(10000000))

	// Create test DB dir
	testDir := path.Join(t.TempDir(), "reorgdetectorTest_ReorgDetector.sqlite")

	reorgDetector, err := New(clientL1.Client(), Config{DBPath: testDir, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)}, L1)
	require.NoError(t, err)

	err = reorgDetector.Start(ctx)
	require.NoError(t, err)

	reorgSub, err := reorgDetector.Subscribe(subID)
	require.NoError(t, err)

	// Block 1
	header1, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit())
	require.NoError(t, err)
	require.Equal(t, uint64(1), header1.Number.Uint64())
	err = reorgDetector.AddBlockToTrack(ctx, subID, header1.Number.Uint64(), header1.Hash()) // Adding block 1
	require.NoError(t, err)

	// Block 2
	header2, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit())
	require.NoError(t, err)
	require.Equal(t, uint64(2), header2.Number.Uint64())
	err = reorgDetector.AddBlockToTrack(ctx, subID, header2.Number.Uint64(), header2.Hash()) // Adding block 1
	require.NoError(t, err)

	// Block 3
	header3Reorged, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit())
	require.NoError(t, err)
	require.Equal(t, uint64(3), header3Reorged.Number.Uint64())
	err = reorgDetector.AddBlockToTrack(ctx, subID, header3Reorged.Number.Uint64(), header3Reorged.Hash()) // Adding block 3
	require.NoError(t, err)

	// Block 4
	header4Reorged, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit())
	require.Equal(t, uint64(4), header4Reorged.Number.Uint64())
	require.NoError(t, err)
	err = reorgDetector.AddBlockToTrack(ctx, subID, header4Reorged.Number.Uint64(), header4Reorged.Hash()) // Adding block 4
	require.NoError(t, err)

	err = clientL1.Fork(header2.Hash()) // Reorg on block 2 (block 2 is still valid)
	require.NoError(t, err)

	// Make sure that the new canonical chain is longer than the previous one so the reorg is visible to the detector
	header3AfterReorg := clientL1.Commit() // Next block 3 after reorg on block 2
	require.NotEqual(t, header3Reorged.Hash(), header3AfterReorg)
	header4AfterReorg := clientL1.Commit() // Block 4
	require.NotEqual(t, header4Reorged.Hash(), header4AfterReorg)
	clientL1.Commit() // Block 5

	// Expect reorg on added blocks 3 -> all further blocks should be removed
	select {
	case firstReorgedBlock := <-reorgSub.ReorgedBlock:
		reorgSub.ReorgProcessed <- true
		require.Equal(t, header3Reorged.Number.Uint64(), firstReorgedBlock)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for reorg")
	}

	// just wait a little for completion
	time.Sleep(time.Second / 5)

	reorgDetector.trackedBlocksLock.Lock()
	headersList, ok := reorgDetector.trackedBlocks[subID]
	reorgDetector.trackedBlocksLock.Unlock()
	require.True(t, ok)
	require.Equal(t, 2, headersList.len()) // Only blocks 1 and 2 left
	actualHeader1, err := headersList.get(1)
	require.NoError(t, err)
	require.Equal(t, header1.Hash(), actualHeader1.Hash)
	actualHeader2, err := headersList.get(2)
	require.NoError(t, err)
	require.Equal(t, header2.Hash(), actualHeader2.Hash)
}

func TestGetTrackedBlocks(t *testing.T) {
	clientL1 := simulated.NewBackend(nil, simulated.WithBlockGasLimit(10000000))
	testDir := path.Join(t.TempDir(), "reorgdetector_TestGetTrackedBlocks.sqlite")
	reorgDetector, err := New(clientL1.Client(), Config{DBPath: testDir, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)}, L1)
	require.NoError(t, err)
	list, err := reorgDetector.getTrackedBlocks()
	require.NoError(t, err)
	require.Equal(t, len(list), 0)

	expectedList := make(map[string]*headersList)
	headersMapFoo := make(map[uint64]header)
	headerFoo2 := header{
		Num:  2,
		Hash: common.HexToHash("foofoo"),
	}
	err = reorgDetector.saveTrackedBlock("foo", headerFoo2)
	require.NoError(t, err)
	headersMapFoo[2] = headerFoo2
	headerFoo3 := header{
		Num:  3,
		Hash: common.HexToHash("foofoofoo"),
	}
	err = reorgDetector.saveTrackedBlock("foo", headerFoo3)
	require.NoError(t, err)
	headersMapFoo[3] = headerFoo3
	expectedList["foo"] = &headersList{
		headers: headersMapFoo,
	}
	list, err = reorgDetector.getTrackedBlocks()
	require.NoError(t, err)
	require.Equal(t, expectedList, list)

	headersMapBar := make(map[uint64]header)
	headerBar2 := header{
		Num:  2,
		Hash: common.HexToHash("BarBar"),
	}
	err = reorgDetector.saveTrackedBlock("Bar", headerBar2)
	require.NoError(t, err)
	headersMapBar[2] = headerBar2
	expectedList["Bar"] = &headersList{
		headers: headersMapBar,
	}
	list, err = reorgDetector.getTrackedBlocks()
	require.NoError(t, err)
	require.Equal(t, expectedList, list)

	require.NoError(t, reorgDetector.loadTrackedHeaders())
	_, ok := reorgDetector.subscriptions["foo"]
	require.True(t, ok)
	_, ok = reorgDetector.subscriptions["Bar"]
	require.True(t, ok)
}

func TestNotSubscribed(t *testing.T) {
	clientL1 := simulated.NewBackend(nil, simulated.WithBlockGasLimit(10000000))
	testDir := path.Join(t.TempDir(), "reorgdetectorTestNotSubscribed.sqlite")
	reorgDetector, err := New(clientL1.Client(), Config{DBPath: testDir, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)}, L1)
	require.NoError(t, err)
	err = reorgDetector.AddBlockToTrack(context.Background(), "foo", 1, common.Hash{})
	require.True(t, strings.Contains(err.Error(), "is not subscribed"))
}

func TestDetectReorgs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	syncerID := "test-syncer"
	trackedBlock := &types.Header{Number: big.NewInt(9)}

	t.Run("Block not finalized", func(t *testing.T) {
		t.Parallel()

		lastFinalizedBlock := &types.Header{Number: big.NewInt(8)}
		client := NewEthClientMock(t)
		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(lastFinalizedBlock, nil)
		client.On("HeaderByNumber", ctx, trackedBlock.Number).Return(trackedBlock, nil)

		testDir := path.Join(t.TempDir(), "reorgdetectorTestDetectReorgs.sqlite")
		reorgDetector, err := New(client, Config{DBPath: testDir, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)}, L1)
		require.NoError(t, err)

		_, err = reorgDetector.Subscribe(syncerID)
		require.NoError(t, err)
		require.NoError(t, reorgDetector.AddBlockToTrack(ctx, syncerID, trackedBlock.Number.Uint64(), trackedBlock.Hash()))

		require.NoError(t, reorgDetector.detectReorgInTrackedList(ctx))

		trackedBlocks, err := reorgDetector.getTrackedBlocks()
		require.NoError(t, err)
		require.Equal(t, 1, len(trackedBlocks))

		syncerTrackedBlocks, ok := trackedBlocks[syncerID]
		require.True(t, ok)
		require.Equal(t, 1, syncerTrackedBlocks.len())
	})

	t.Run("Block finalized", func(t *testing.T) {
		t.Parallel()

		lastFinalizedBlock := trackedBlock
		client := NewEthClientMock(t)
		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(lastFinalizedBlock, nil)

		testDir := path.Join(t.TempDir(), "reorgdetectorTestDetectReorgs.sqlite")
		reorgDetector, err := New(client, Config{DBPath: testDir, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)}, L1)
		require.NoError(t, err)

		_, err = reorgDetector.Subscribe(syncerID)
		require.NoError(t, err)
		require.NoError(t, reorgDetector.AddBlockToTrack(ctx, syncerID, trackedBlock.Number.Uint64(), trackedBlock.Hash()))

		require.NoError(t, reorgDetector.detectReorgInTrackedList(ctx))

		trackedBlocks, err := reorgDetector.getTrackedBlocks()
		require.NoError(t, err)
		require.Equal(t, 0, len(trackedBlocks))
	})

	t.Run("Reorg happened", func(t *testing.T) {
		t.Parallel()

		lastFinalizedBlock := &types.Header{Number: big.NewInt(5)}
		reorgedTrackedBlock := &types.Header{Number: trackedBlock.Number, Extra: []byte("reorged")} // Different hash

		client := NewEthClientMock(t)
		client.On("HeaderByNumber", ctx, big.NewInt(int64(rpc.FinalizedBlockNumber))).Return(lastFinalizedBlock, nil)
		client.On("HeaderByNumber", ctx, trackedBlock.Number).Return(reorgedTrackedBlock, nil)

		testDir := path.Join(t.TempDir(), "reorgdetectorTestDetectReorgs.sqlite")
		reorgDetector, err := New(client, Config{DBPath: testDir, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)}, L1)
		require.NoError(t, err)

		subscription, err := reorgDetector.Subscribe(syncerID)
		require.NoError(t, err)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			<-subscription.ReorgedBlock
			subscription.ReorgProcessed <- true

			wg.Done()
		}()

		require.NoError(t, reorgDetector.AddBlockToTrack(ctx, syncerID, trackedBlock.Number.Uint64(), trackedBlock.Hash()))

		require.NoError(t, reorgDetector.detectReorgInTrackedList(ctx))

		wg.Wait() // we wait here to make sure the reorg is processed

		trackedBlocks, err := reorgDetector.getTrackedBlocks()
		require.NoError(t, err)
		require.Equal(t, 0, len(trackedBlocks)) // shouldn't be any since a reorg happened on that block
	})
}
