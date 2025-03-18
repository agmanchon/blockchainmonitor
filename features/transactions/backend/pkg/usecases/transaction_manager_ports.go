package usecases

import (
	"context"
	"math/big"
	"time"
)

type TransactionStorage interface {
	// All blockchain events in gatewayEventStorage configurable by BlockStorageListOptions. It returns an error if it fails
	All(ctx context.Context, filter TransactionListOptions) (*TransactionCollection, error)
	// Get a blockchain event in gatewayEventStorage. It returns an error if it fails
	Get(ctx context.Context, id string) (*Transaction, error)
	// Edit a blockchain event in gatewayEventStorage. It returns an error if it fails
	Edit(ctx context.Context, transaction Transaction) (*Transaction, error)
	// CreateInsertBatch provides a Batch to arrange data in sets to add events to blockchainEventStorage
	CreateInsertBatch() TransactionInsertBatchPort
}

type TransactionInsertBatchPort interface {
	// Add inserts an event into the batch's memory
	Add(ctx context.Context, transaction Transaction) (*Transaction, error)
	// Flush persists the events stored in the batch's memory
	Flush(ctx context.Context) error
	// Clear resets the batch's memory
	Clear()
}
type Transaction struct {
	ID           string
	ChainID      string
	State        string
	BlockNumber  big.Int
	CreationDate time.Time
	LastUpdated  time.Time
}
type TransactionCollection struct {
	Items []Transaction
}

type TransactionListOptions struct {
	State      string
	LastUpdate time.Time
}

type TransactionQueue interface {
	// add Inserts the specified item into this queue. true result means transaction was add, while false means it wasnt added because already exists in the queue
	add(item Transaction) bool
	// addDelayed Inserts the specified item into this queue and it will be delivered after now + duration. true result means transaction was add, while false means it wasnt added because already exists in the queue
	addDelayed(item Transaction, deliverIn time.Duration) bool
	// peek retrieves, but does not remove, the head of this queue. Returns nil if this queue is empty.
	peek() *Transaction
	// poll retrieves and removes the head of this queue. Returns nil if this queue is empty.
	poll() *Transaction
	// shutdown shuts down the queue system
	shutdown()
}

type KafkaProducer interface {
	SendEvent(ctx context.Context, event KafkaEvent) error
	Close()
}

type KafkaConsumer interface {
	StartConsumer(ctx context.Context, messageHandler func(KafkaEvent) error) error
	Close()
}

type KafkaEvent struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Amount      string `json:"amount"`
	Fees        string `json:"fees"`
}
