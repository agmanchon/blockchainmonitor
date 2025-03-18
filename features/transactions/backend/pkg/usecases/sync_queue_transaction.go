package usecases

import (
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/syncqueue"
	"time"
)

var _ TransactionQueue = (*gatewayTxBlockingSyncQueue)(nil)

func NewTransactionsBlockingSyncQueue(queueName string) *gatewayTxBlockingSyncQueue {

	queue := syncqueue.NewGenericBlockingSyncQueue[Transaction](queueName)
	return &gatewayTxBlockingSyncQueue{innerQueue: queue}
}

type gatewayTxBlockingSyncQueue struct {
	innerQueue syncqueue.SyncQueue[Transaction]
}

func (g gatewayTxBlockingSyncQueue) shutdown() {
	g.innerQueue.Shutdown()
}

func (g gatewayTxBlockingSyncQueue) add(item Transaction) bool {
	return g.innerQueue.Add(syncqueue.QueueItem[Transaction]{
		ID:      item.ID,
		Content: &item,
	})
}

func (g gatewayTxBlockingSyncQueue) addDelayed(item Transaction, deliverIn time.Duration) bool {
	return g.innerQueue.AddDelayed(syncqueue.QueueItem[Transaction]{
		ID:      item.ID,
		Content: &item,
	}, deliverIn)
}

func (g gatewayTxBlockingSyncQueue) peek() *Transaction {
	item := g.innerQueue.Peek()
	if item == nil {
		return nil
	}

	return item.Content
}

func (g gatewayTxBlockingSyncQueue) poll() *Transaction {
	item := g.innerQueue.Poll()
	if item == nil {
		return nil
	}

	return item.Content
}
