package postgres

import (
	"context"
	"errors"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/big"
	"time"
)

type PostgresTransactionStorageAdapter struct {
	db *pgxpool.Pool
}

type PostgresTransactionStorageAdapterOptions struct {
	DB *pgxpool.Pool
}

var _ usecases.TransactionStorage = (*PostgresTransactionStorageAdapter)(nil)

func ProvidePostgresTransactionStorageAdapter(options PostgresTransactionStorageAdapterOptions) (*PostgresTransactionStorageAdapter, error) {
	if options.DB == nil {
		return nil, errors.New("db connection is required")
	}
	return &PostgresTransactionStorageAdapter{
		db: options.DB,
	}, nil
}

func (p PostgresTransactionStorageAdapter) All(ctx context.Context, options usecases.TransactionListOptions) (*usecases.TransactionCollection, error) {
	query := `
		SELECT id, block_number, chain_id, state, creation_date, last_updated 
		FROM transactions 
		WHERE state=$1 AND last_updated < $2
	`
	rows, err := p.db.Query(ctx, query, options.State, options.LastUpdate.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []usecases.Transaction
	for rows.Next() {
		var t usecases.Transaction
		var blockNumberString string
		var creationDateUnix, lastUpdatedUnix int64
		if err := rows.Scan(&t.ID, &blockNumberString, &t.ChainID, &t.State, &creationDateUnix, &lastUpdatedUnix); err != nil {
			return nil, err
		}

		blockNumber := new(big.Int)
		blockNumber, success := blockNumber.SetString(blockNumberString, 10)
		if !success {
			return nil, errors.New("could not parse block number")
		}
		t.BlockNumber = *blockNumber

		t.CreationDate = time.Unix(creationDateUnix, 0)
		t.LastUpdated = time.Unix(lastUpdatedUnix, 0)

		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &usecases.TransactionCollection{Items: transactions}, nil
}

func (p PostgresTransactionStorageAdapter) Get(ctx context.Context, id string) (*usecases.Transaction, error) {
	var transaction usecases.Transaction
	var blockNumberString string
	var creationDateUnix, lastUpdatedUnix int64
	err := p.db.QueryRow(ctx, `
		SELECT id, block_number, chain_id, state, creation_date, last_updated
		FROM transactions
		WHERE id = $1
	`, id).Scan(&transaction.ID, &blockNumberString, &transaction.ChainID, &transaction.State, &creationDateUnix, &lastUpdatedUnix)

	if err != nil {
		return nil, err
	}

	blockNumber := new(big.Int)
	blockNumber, success := blockNumber.SetString(blockNumberString, 10)
	if !success {
		return nil, errors.New("could not parse block number")
	}
	transaction.BlockNumber = *blockNumber

	transaction.CreationDate = time.Unix(creationDateUnix, 0)
	transaction.LastUpdated = time.Unix(lastUpdatedUnix, 0)

	return &transaction, nil
}

// Edit - Update a transaction
func (p PostgresTransactionStorageAdapter) Edit(ctx context.Context, transaction usecases.Transaction) (*usecases.Transaction, error) {
	_, err := p.db.Exec(ctx, `
		UPDATE transactions
		SET block_number = $2, state = $3, chain_id = $4, creation_date = $5, last_updated = $6
		WHERE id = $1
	`, transaction.ID, transaction.BlockNumber.String(), transaction.State, transaction.ChainID, transaction.CreationDate.Unix(), transaction.LastUpdated.Unix())

	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// CreateInsertBatch - Returns a batch handler for transactions
func (p PostgresTransactionStorageAdapter) CreateInsertBatch() usecases.TransactionInsertBatchPort {
	return &PostgresTransactionBatch{
		db:               p.db,
		transactionBatch: make([]usecases.Transaction, 0),
	}
}

// PostgresTransactionBatch - Implements TransactionInsertBatchPort
type PostgresTransactionBatch struct {
	db               *pgxpool.Pool
	transactionBatch []usecases.Transaction
}

// Add - Adds a transaction to the batch
func (b *PostgresTransactionBatch) Add(ctx context.Context, transaction usecases.Transaction) (*usecases.Transaction, error) {
	b.transactionBatch = append(b.transactionBatch, transaction)
	return &transaction, nil
}

// Flush - Persists all transactions in the batch
func (b *PostgresTransactionBatch) Flush(ctx context.Context) error {
	// Prepare the insert query
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert each transaction into the database
	for _, t := range b.transactionBatch {
		_, err := tx.Exec(ctx, `
			INSERT INTO transactions (id, block_number, chain_id, state, creation_date, last_updated)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, t.ID, t.BlockNumber.String(), t.ChainID, t.State, t.CreationDate.Unix(), t.LastUpdated.Unix())

		if err != nil {
			return err
		}
	}

	// Commit the transaction batch
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// Clear the batch after successful flush
	b.Clear()
	return nil
}

// Clear - Resets the batch memory
func (b *PostgresTransactionBatch) Clear() {
	b.transactionBatch = make([]usecases.Transaction, 0)
}
