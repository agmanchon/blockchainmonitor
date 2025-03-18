package syncqueue

import (
	"time"
)

type SyncQueue[K any] interface {
	Add(item QueueItem[K]) bool
	AddDelayed(item QueueItem[K], deliverIn time.Duration) bool
	Peek() *QueueItem[K]
	Poll() *QueueItem[K]
	Shutdown()
}

type QueueItem[K any] struct {
	ID      string
	Content *K
}

func NewQueueItem[K any](id string, content *K) QueueItem[K] {
	return QueueItem[K]{
		ID:      id,
		Content: content,
	}
}
