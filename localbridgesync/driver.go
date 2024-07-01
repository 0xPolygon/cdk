package localbridgesync

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	checkReorgInterval = time.Second * 10
	downloadBufferSize = 100
)

type driver struct {
	p *processor
	d *downloader
}

func newDriver() (*driver, error) {
	return nil, errors.New("not implemented")
}

func (d *driver) Sync(ctx context.Context) {
	for {
		lastProcessedBlock, err := d.p.getLastProcessedBlock(ctx)
		if err != nil {
			// TODO: handle error
			return
		}
		cancellableCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// start downloading
		downloadCh := make(chan block, downloadBufferSize)
		go d.d.download(cancellableCtx, lastProcessedBlock, downloadCh)

		// detect potential reorgs
		reorgCh := make(chan uint64)
		go d.detectReorg(cancellableCtx, reorgCh)

		for {
			select {
			case b := <-downloadCh: // new block from downloader
				err = d.storeBlockToReorgTracker(b.blockHeader)
				if err != nil {
					// TODO: handle error
					return
				}
				err = d.p.storeBridgeEvents(b.Num, b.Events)
				if err != nil {
					// TODO: handle error
					return
				}
			case lastValidBlock := <-reorgCh: // reorg detected
				// stop downloader
				cancel()
				// wait until downloader closes channel
				_, ok := <-downloadCh
				for ok {
					_, ok = <-downloadCh
				}
				// handle reorg
				err = d.p.reorg(lastValidBlock)
				if err != nil {
					// TODO: handle error
					return
				}

				// restart syncing
				break
			}
		}
	}
}

// IMO we could make a package "reorg detector" that could be reused by all the syncers
// each instance should use it's own storage / bucket

func (d *driver) detectReorg(ctx context.Context, reorgCh chan uint64) {
	var (
		expectedHeader blockHeader
		err            error
	)
	for {
		expectedHeader, err = d.getGreatestBlockFromReorgTracker()
		if err != nil {
			// TODO: handle error
			return
		}
		actualHeader, err := d.d.getBlockHeader(ctx, expectedHeader.Num)
		if err != nil {
			// TODO: handle error
			return
		}
		if actualHeader.Hash != expectedHeader.Hash {
			fmt.Printf("reorg detected")
			break
		}
		time.Sleep(checkReorgInterval)
	}
	// Find last valid block
	oldestTrackedBlock, err := d.getSmallestBlockFromReorgTracker()
	if err != nil {
		// TODO: handle error
		return
	}
	lastValidBlock := oldestTrackedBlock.Num
	for i := expectedHeader.Num - 1; i > lastValidBlock; i-- {
		expectedHeader, err = d.getBlockFromReorgTracker(i)
		if err != nil {
			// TODO: handle error
			return
		}
		actualHeader, err := d.d.getBlockHeader(ctx, expectedHeader.Num)
		if err != nil {
			// TODO: handle error
			return
		}
		if actualHeader.Hash == expectedHeader.Hash {
			fmt.Printf("valid block detected")
			reorgCh <- actualHeader.Num
			return
		}
	}
	// this should never happen! But if it happens it should delete ALL the data from the processor
	// or have logic to handle this "special case"
	fmt.Printf("no valid block detected!!!!")
	reorgCh <- oldestTrackedBlock.Num
	return
}

func (d *driver) storeBlockToReorgTracker(b blockHeader) error {
	fmt.Printf("adding block %d with hash %s to the reorg tracker storage\n", b.Num, b.Hash)
	return errors.New("not implemented")
}

func (d *driver) removeBlockFromReorgTracker(blockNum uint64) error {
	fmt.Printf("removing block %d from the reorg tracker storage\n", blockNum)
	return errors.New("not implemented")
}

func (d *driver) getGreatestBlockFromReorgTracker() (blockHeader, error) {
	fmt.Println("getting the block with the greatest block num from the reorg tracker storage")
	return blockHeader{}, errors.New("not implemented")
}

func (d *driver) getBlockFromReorgTracker(blockNum uint64) (blockHeader, error) {
	fmt.Printf("getting the block %d from the reorg tracker storage", blockNum)
	return blockHeader{}, errors.New("not implemented")
}

func (d *driver) getSmallestBlockFromReorgTracker() (blockHeader, error) {
	fmt.Println("getting the block with the smallest block num from the reorg tracker storage")
	return blockHeader{}, errors.New("not implemented")
}
