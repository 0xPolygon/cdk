package reorgdetector

// Subscription is a subscription to reorg events
type Subscription struct {
	ReorgedBlock   chan uint64
	ReorgProcessed chan bool
}

// Subscribe subscribes to reorg events
func (rd *ReorgDetector) Subscribe(id string) (*Subscription, error) {
	rd.subscriptionsLock.Lock()
	defer rd.subscriptionsLock.Unlock()

	if sub, ok := rd.subscriptions[id]; ok {
		return sub, nil
	}

	// Create a new subscription
	sub := &Subscription{
		ReorgedBlock:   make(chan uint64),
		ReorgProcessed: make(chan bool),
	}
	rd.subscriptions[id] = sub

	// Create a new tracked blocks list for the subscriber
	rd.trackedBlocksLock.Lock()
	rd.trackedBlocks[id] = newHeadersList()
	rd.trackedBlocksLock.Unlock()

	// Create a new notified reorgs list for the subscriber
	rd.notifiedReorgsLock.Lock()
	rd.notifiedReorgs[id] = make(map[uint64]struct{})
	rd.notifiedReorgsLock.Unlock()

	return sub, nil
}

// notifySubscriber notifies the subscriber with the block of the reorg
func (rd *ReorgDetector) notifySubscriber(id string, startingBlock header) {
	// Check if the given reorg was already notified to the given subscriber
	rd.notifiedReorgsLock.RLock()
	if _, ok := rd.notifiedReorgs[id][startingBlock.Num]; ok {
		rd.notifiedReorgsLock.RUnlock()
		return
	}
	rd.notifiedReorgsLock.RUnlock()

	// Notify subscriber about this particular reorg
	rd.subscriptionsLock.RLock()
	if sub, ok := rd.subscriptions[id]; ok {
		sub.ReorgedBlock <- startingBlock.Num
		<-sub.ReorgProcessed
	}
	rd.subscriptionsLock.RUnlock()

	// Mark the reorg as notified
	rd.notifiedReorgsLock.Lock()
	rd.notifiedReorgs[id][startingBlock.Num] = struct{}{}
	rd.notifiedReorgsLock.Unlock()
}
