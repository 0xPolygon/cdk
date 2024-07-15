package localbridgesync

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	retryAfterErrorPeriod = time.Millisecond * 100
	rdm := NewReorgDetectorMock(t)
	pm := NewProcessorMock(t)
	dm := NewDownloaderMock(t)
	firstReorgedBlock := make(chan uint64)
	reorgProcessed := make(chan bool)
	rdm.On("Subscribe", reorgDetectorID).Return(&reorgdetector.Subscription{
		FirstReorgedBlock: firstReorgedBlock,
		ReorgProcessed:    reorgProcessed,
	})
	driver, err := newDriver(rdm, pm, dm)
	require.NoError(t, err)
	ctx := context.Background()
	expectedBlock1 := block{
		blockHeader: blockHeader{
			Num:  3,
			Hash: common.HexToHash("03"),
		},
	}
	expectedBlock2 := block{
		blockHeader: blockHeader{
			Num:  9,
			Hash: common.HexToHash("09"),
		},
	}
	type reorgSemaphore struct {
		mu    sync.Mutex
		green bool
	}
	reorg1Completed := reorgSemaphore{}

	mockDownload := func(
		ctx context.Context,
		d downloaderInterface,
		fromBlock uint64,
		downloadedCh chan block,
	) {
		log.Info("entering mock loop")
		for {
			select {
			case <-ctx.Done():
				log.Info("closing channel")
				close(downloadedCh)
				return
			default:
			}
			reorg1Completed.mu.Lock()
			green := reorg1Completed.green
			reorg1Completed.mu.Unlock()
			if green {
				downloadedCh <- expectedBlock2
			} else {
				downloadedCh <- expectedBlock1
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Mocking this actions, the driver should "store" all the blocks from the downloader
	pm.On("getLastProcessedBlock", ctx).
		Return(uint64(3), nil)
	rdm.On("AddBlockToTrack", ctx, reorgDetectorID, expectedBlock1.Num, expectedBlock1.Hash).
		Return(nil)
	pm.On("storeBridgeEvents", expectedBlock1.Num, expectedBlock1.Events).
		Return(nil)
	rdm.On("AddBlockToTrack", ctx, reorgDetectorID, expectedBlock2.Num, expectedBlock2.Hash).
		Return(nil)
	pm.On("storeBridgeEvents", expectedBlock2.Num, expectedBlock2.Events).
		Return(nil)
	go driver.Sync(ctx, mockDownload)
	time.Sleep(time.Millisecond * 200) // time to download expectedBlock1

	// Trigger reorg 1
	reorgedBlock1 := uint64(5)
	pm.On("reorg", reorgedBlock1).Return(nil)
	firstReorgedBlock <- reorgedBlock1
	ok := <-reorgProcessed
	require.True(t, ok)
	reorg1Completed.mu.Lock()
	reorg1Completed.green = true
	reorg1Completed.mu.Unlock()
	time.Sleep(time.Millisecond * 200) // time to download expectedBlock2

	// Trigger reorg 2: syncer restarts the porcess
	reorgedBlock2 := uint64(7)
	pm.On("reorg", reorgedBlock2).Return(nil)
	firstReorgedBlock <- reorgedBlock2
	ok = <-reorgProcessed
	require.True(t, ok)
}

func TestHandleNewBlock(t *testing.T) {
	retryAfterErrorPeriod = time.Millisecond * 100
	rdm := NewReorgDetectorMock(t)
	pm := NewProcessorMock(t)
	dm := NewDownloaderMock(t)
	rdm.On("Subscribe", reorgDetectorID).Return(&reorgdetector.Subscription{})
	driver, err := newDriver(rdm, pm, dm)
	require.NoError(t, err)
	ctx := context.Background()

	// happy path
	b1 := block{
		blockHeader: blockHeader{
			Num:  1,
			Hash: common.HexToHash("f00"),
		},
	}
	rdm.
		On("AddBlockToTrack", ctx, reorgDetectorID, b1.Num, b1.Hash).
		Return(nil)
	pm.On("storeBridgeEvents", b1.Num, b1.Events).
		Return(nil)
	driver.handleNewBlock(ctx, b1)

	// reorg deteector fails once
	b2 := block{
		blockHeader: blockHeader{
			Num:  2,
			Hash: common.HexToHash("f00"),
		},
	}
	rdm.
		On("AddBlockToTrack", ctx, reorgDetectorID, b2.Num, b2.Hash).
		Return(errors.New("foo")).Once()
	rdm.
		On("AddBlockToTrack", ctx, reorgDetectorID, b2.Num, b2.Hash).
		Return(nil).Once()
	pm.On("storeBridgeEvents", b2.Num, b2.Events).
		Return(nil)
	driver.handleNewBlock(ctx, b2)

	// processor fails once
	b3 := block{
		blockHeader: blockHeader{
			Num:  3,
			Hash: common.HexToHash("f00"),
		},
	}
	rdm.
		On("AddBlockToTrack", ctx, reorgDetectorID, b3.Num, b3.Hash).
		Return(nil)
	pm.On("storeBridgeEvents", b3.Num, b3.Events).
		Return(errors.New("foo")).Once()
	pm.On("storeBridgeEvents", b3.Num, b3.Events).
		Return(nil).Once()
	driver.handleNewBlock(ctx, b3)

}

func TestHandleReorg(t *testing.T) {
	retryAfterErrorPeriod = time.Millisecond * 100
	rdm := NewReorgDetectorMock(t)
	pm := NewProcessorMock(t)
	dm := NewDownloaderMock(t)
	reorgProcessed := make(chan bool)
	rdm.On("Subscribe", reorgDetectorID).Return(&reorgdetector.Subscription{
		ReorgProcessed: reorgProcessed,
	})
	driver, err := newDriver(rdm, pm, dm)
	require.NoError(t, err)
	ctx := context.Background()

	// happy path
	_, cancel := context.WithCancel(ctx)
	downloadCh := make(chan block)
	firstReorgedBlock := uint64(5)
	pm.On("reorg", firstReorgedBlock).Return(nil)
	go driver.handleReorg(cancel, downloadCh, firstReorgedBlock)
	close(downloadCh)
	done := <-reorgProcessed
	require.True(t, done)

	// download ch sends some garbage
	_, cancel = context.WithCancel(ctx)
	downloadCh = make(chan block)
	firstReorgedBlock = uint64(6)
	pm.On("reorg", firstReorgedBlock).Return(nil)
	go driver.handleReorg(cancel, downloadCh, firstReorgedBlock)
	downloadCh <- block{}
	downloadCh <- block{}
	downloadCh <- block{}
	close(downloadCh)
	done = <-reorgProcessed
	require.True(t, done)

	// processor fails 2 times
	_, cancel = context.WithCancel(ctx)
	downloadCh = make(chan block)
	firstReorgedBlock = uint64(7)
	pm.On("reorg", firstReorgedBlock).Return(errors.New("foo")).Once()
	pm.On("reorg", firstReorgedBlock).Return(errors.New("foo")).Once()
	pm.On("reorg", firstReorgedBlock).Return(nil).Once()
	go driver.handleReorg(cancel, downloadCh, firstReorgedBlock)
	close(downloadCh)
	done = <-reorgProcessed
	require.True(t, done)
}
