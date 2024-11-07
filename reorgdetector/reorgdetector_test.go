package reorgdetector

import (
	"context"
	"path"
	"strings"
	"testing"
	"time"

	cdktypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/test/helpers"
	common "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func Test_ReorgDetector(t *testing.T) {
	const subID = "test"

	ctx := context.Background()

	// Simulated L1
	clientL1, _ := helpers.SimulatedBackend(t, nil, 0)

	// Create test DB dir
	testDir := path.Join(t.TempDir(), "file::memory:?cache=shared")

	reorgDetector, err := New(clientL1.Client(), Config{DBPath: testDir, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)})
	require.NoError(t, err)

	err = reorgDetector.Start(ctx)
	require.NoError(t, err)

	reorgSub, err := reorgDetector.Subscribe(subID)
	require.NoError(t, err)

	remainingHeader, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit()) // Block 2
	require.NoError(t, err)
	err = reorgDetector.AddBlockToTrack(ctx, subID, remainingHeader.Number.Uint64(), remainingHeader.Hash()) // Adding block 2
	require.NoError(t, err)
	reorgHeader, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit()) // Block 3
	require.NoError(t, err)
	firstHeaderAfterReorg, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit()) // Block 4
	require.NoError(t, err)
	err = reorgDetector.AddBlockToTrack(ctx, subID, firstHeaderAfterReorg.Number.Uint64(), firstHeaderAfterReorg.Hash()) // Adding block 4
	require.NoError(t, err)
	header, err := clientL1.Client().HeaderByHash(ctx, clientL1.Commit()) // Block 5
	require.NoError(t, err)
	err = reorgDetector.AddBlockToTrack(ctx, subID, header.Number.Uint64(), header.Hash()) // Adding block 5
	require.NoError(t, err)
	err = clientL1.Fork(reorgHeader.Hash()) // Reorg on block 3
	require.NoError(t, err)
	clientL1.Commit() // Next block 4 after reorg on block 3
	clientL1.Commit() // Block 5
	clientL1.Commit() // Block 6

	// Expect reorg on added blocks 4 -> all further blocks should be removed
	select {
	case firstReorgedBlock := <-reorgSub.ReorgedBlock:
		reorgSub.ReorgProcessed <- true
		require.Equal(t, firstHeaderAfterReorg.Number.Uint64(), firstReorgedBlock)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for reorg")
	}

	// just wait a little for completion
	time.Sleep(time.Second / 5)

	reorgDetector.trackedBlocksLock.Lock()
	headersList, ok := reorgDetector.trackedBlocks[subID]
	reorgDetector.trackedBlocksLock.Unlock()
	require.True(t, ok)
	require.Equal(t, 1, headersList.len()) // Only block 3 left
	require.Equal(t, remainingHeader.Hash(), headersList.get(4).Hash)
}

func TestGetTrackedBlocks(t *testing.T) {
	clientL1, _ := helpers.SimulatedBackend(t, nil, 0)
	testDir := path.Join(t.TempDir(), "file::memory:?cache=shared")
	reorgDetector, err := New(clientL1.Client(), Config{DBPath: testDir, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)})
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
	clientL1, _ := helpers.SimulatedBackend(t, nil, 0)
	testDir := path.Join(t.TempDir(), "file::memory:?cache=shared")
	reorgDetector, err := New(clientL1.Client(), Config{DBPath: testDir, CheckReorgsInterval: cdktypes.NewDuration(time.Millisecond * 100)})
	require.NoError(t, err)
	err = reorgDetector.AddBlockToTrack(context.Background(), "foo", 1, common.Hash{})
	require.True(t, strings.Contains(err.Error(), "is not subscribed"))
}
