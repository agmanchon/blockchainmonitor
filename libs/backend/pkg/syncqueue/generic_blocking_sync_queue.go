package syncqueue

import (
	"container/list"
	"sync"
	"time"
)

var _ SyncQueue[any] = (*GenericBlockingSyncQueue[any])(nil)

func NewGenericBlockingSyncQueue[K any](queueName string) *GenericBlockingSyncQueue[K] {
	return &GenericBlockingSyncQueue[K]{
		innerList:    list.New(),
		itemsInQueue: make(map[string]bool),
		queueName:    queueName,
		cond:         sync.NewCond(new(sync.Mutex)),
		active:       true,
	}
}

// GenericBlockingSyncQueue SyncQueue implementation blocking consumers if the queue is empty
type GenericBlockingSyncQueue[K any] struct {
	innerList    *list.List
	cond         *sync.Cond
	itemsInQueue map[string]bool
	queueName    string
	active       bool
}

// Add Inserts the specified item into this queue
func (queue *GenericBlockingSyncQueue[K]) Add(item QueueItem[K]) bool {
	return queue.doAdd(item, false)
}

// doAdd Inserts the specified item into this queue
func (queue *GenericBlockingSyncQueue[K]) doAdd(item QueueItem[K], delayed bool) bool {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()

	if !delayed {
		if _, ok := queue.itemsInQueue[item.ID]; ok {
			return false
		}
	}

	wrappedItem := itemWrapper[K]{
		item:         item,
		delayed:      false,
		deliveryTime: time.Time{},
	}

	if queue.innerList.Len() == 0 {
		queue.innerList.PushFront(wrappedItem)
	} else {
		lastElement := queue.innerList.Back()
		queue.innerList.InsertAfter(wrappedItem, lastElement)
	}

	queue.itemsInQueue[item.ID] = true
	queue.cond.Signal()
	return true
}

// AddDelayed Inserts the specified item into this queue and it will be delivered after now + deliverIn
func (queue *GenericBlockingSyncQueue[K]) AddDelayed(item QueueItem[K], deliverIn time.Duration) bool {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()

	if _, ok := queue.itemsInQueue[item.ID]; ok {
		return false
	}

	queue.itemsInQueue[item.ID] = true

	timer := time.NewTimer(deliverIn)

	go func() {
		<-timer.C
		queue.doAdd(item, true)

	}()

	return true
}

// Peek retrieves, but does not remove, the head of this queue. Returns nil if this queue is empty.
func (queue *GenericBlockingSyncQueue[K]) Peek() *QueueItem[K] {
	// Note: peek function won't be exposed as metric because it is a readonly operation
	currentElement := queue.innerList.Front()
	for currentElement != nil {
		wrappedItem, _ := currentElement.Value.(itemWrapper[K])
		if !wrappedItem.delayed {
			return &wrappedItem.item
		}
		if wrappedItem.deliveryTime.Before(time.Now()) {
			return &wrappedItem.item
		}
		currentElement = currentElement.Next()
	}

	return nil
}

// Poll retrieves and removes the head of this queue. Returns nil if this queue is marked as inactive (shutdown process)
func (queue *GenericBlockingSyncQueue[K]) Poll() *QueueItem[K] {
	queue.cond.L.Lock()
	for queue.active && queue.innerList.Len() == 0 {
		queue.cond.Wait()
	}
	defer queue.cond.L.Unlock()
	if !queue.active {
		return nil
	}
	currentElement := queue.innerList.Front()

	wrappedItem, _ := currentElement.Value.(itemWrapper[K])
	queue.innerList.Remove(currentElement)
	delete(queue.itemsInQueue, wrappedItem.item.ID)
	return &wrappedItem.item
}

// Shutdown shuts down the queue. Data can be lost depending on the implementation of the queue
func (queue *GenericBlockingSyncQueue[K]) Shutdown() {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()
	queue.active = false
	queue.cond.Broadcast()
}

type itemWrapper[K any] struct {
	item         QueueItem[K]
	delayed      bool
	deliveryTime time.Time
}
