package reorgdetector

import (
	"context"
	"testing"
	"time"

	cdktypes "github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/test/helpers"
	"github.com/stretchr/testify/require"
)

func Test_ReorgDetector(t *testing.T) {
	const subID = "test"

	ctx := context.Background()

	// Simulated L1
	clientL1, _ := helpers.SimulatedBackend(t, nil, 0)

	// Create test DB dir
	testDir := t.TempDir()

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
