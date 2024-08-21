package reorgdetector

import "sync"

type Subscription struct {
	FirstReorgedBlock          chan uint64
	ReorgProcessed             chan bool
	pendingReorgsToBeProcessed sync.WaitGroup
}

func (rd *ReorgDetector) Subscribe(id string) (*Subscription, error) {
	rd.subscriptionsLock.Lock()
	defer rd.subscriptionsLock.Unlock()

	if sub, ok := rd.subscriptions[id]; ok {
		return sub, nil
	}

	sub := &Subscription{
		FirstReorgedBlock: make(chan uint64),
		ReorgProcessed:    make(chan bool),
	}
	rd.subscriptions[id] = sub

	return sub, nil
}

func (rd *ReorgDetector) notifySubscribers(startingBlock header) {
	rd.subscriptionsLock.RLock()
	for _, sub := range rd.subscriptions {
		sub.pendingReorgsToBeProcessed.Add(1)
		go func(sub *Subscription) {
			sub.FirstReorgedBlock <- startingBlock.Num
			<-sub.ReorgProcessed
			sub.pendingReorgsToBeProcessed.Done()
		}(sub)
	}
	rd.subscriptionsLock.RUnlock()
}

func (rd *ReorgDetector) notifySubscriber(id string, startingBlock header) {
	rd.subscriptionsLock.RLock()
	sub, ok := rd.subscriptions[id]
	if ok {
		sub.pendingReorgsToBeProcessed.Add(1)
		sub.FirstReorgedBlock <- startingBlock.Num
		<-sub.ReorgProcessed
		sub.pendingReorgsToBeProcessed.Done()
	}
	rd.subscriptionsLock.RUnlock()
}
