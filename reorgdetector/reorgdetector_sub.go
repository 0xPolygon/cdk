package reorgdetector

import (
	"fmt"
	"sync"
)

type Subscription struct {
	ReorgedBlock   chan uint64
	ReorgProcessed chan bool

	pendingReorgsToBeProcessed sync.WaitGroup
}

func (rd *ReorgDetector) Subscribe(id string) (*Subscription, error) {
	rd.subscriptionsLock.Lock()
	defer rd.subscriptionsLock.Unlock()

	if sub, ok := rd.subscriptions[id]; ok {
		return sub, nil
	}

	sub := &Subscription{
		ReorgedBlock:   make(chan uint64),
		ReorgProcessed: make(chan bool),
	}
	rd.subscriptions[id] = sub

	rd.trackedBlocksLock.Lock()
	rd.trackedBlocks[id] = newHeadersList()
	rd.trackedBlocksLock.Unlock()

	return sub, nil
}

// notifySubscriber notifies the subscriber with the block of the reorg
func (rd *ReorgDetector) notifySubscriber(id string, startingBlock header) {
	// Check if the given reorg was already notified to the given subscriber
	reorgKey := fmt.Sprintf("%s_%d", id, startingBlock.Num)
	rd.notifiedReorgsLock.RLock()
	if _, ok := rd.notifiedReorgs[reorgKey]; ok {
		rd.notifiedReorgsLock.RUnlock()
		return
	}
	rd.notifiedReorgsLock.RUnlock()

	// Notify subscriber about this particular reorg
	rd.subscriptionsLock.RLock()
	if sub, ok := rd.subscriptions[id]; ok {
		sub.pendingReorgsToBeProcessed.Add(1)
		sub.ReorgedBlock <- startingBlock.Num
		<-sub.ReorgProcessed
		sub.pendingReorgsToBeProcessed.Done()
	}
	rd.subscriptionsLock.RUnlock()

	// Mark the reorg as notified
	rd.notifiedReorgsLock.RLock()
	rd.notifiedReorgs[reorgKey] = struct{}{}
	rd.notifiedReorgsLock.RUnlock()
}
