package reorgdetector

import "sync"

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
	rd.subscriptionsLock.RLock()
	if sub, ok := rd.subscriptions[id]; ok {
		sub.pendingReorgsToBeProcessed.Add(1)
		sub.ReorgedBlock <- startingBlock.Num
		<-sub.ReorgProcessed
		sub.pendingReorgsToBeProcessed.Done()
	}
	rd.subscriptionsLock.RUnlock()
}
