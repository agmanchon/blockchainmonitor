package postgres

import (
	"context"
	"errors"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/big"
)

type PostgresBlockStorageAdapter struct {
	db *pgxpool.Pool
}

type PostgresBlockStorageAdapterOptions struct {
	DB *pgxpool.Pool
}

var _ usecases.BlockStorage = (*PostgresBlockStorageAdapter)(nil)

func ProvidePostgresBlockStorageAdapter(options PostgresBlockStorageAdapterOptions) (*PostgresBlockStorageAdapter, error) {
	if options.DB == nil {
		return nil, errors.New("db connection is required")
	}
	return &PostgresBlockStorageAdapter{
		db: options.DB,
	}, nil
}

func (p PostgresBlockStorageAdapter) Add(ctx context.Context, block usecases.Block) (*usecases.Block, error) {
	_, err := p.db.Exec(ctx, `
		INSERT INTO blocks (id, block_number, state, chain_id)
		VALUES ($1, $2, $3, $4)
	`, block.ID, block.BlockNumber.String(), block.State, block.ChainID)

	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (p PostgresBlockStorageAdapter) AllByDescOrder(ctx context.Context, options usecases.BlockStorageListOptions) (*usecases.BlockCollection, error) {
	query := `
		SELECT id, block_number, state, chain_id
		FROM blocks
		WHERE chain_id = $1 AND state = $2
		ORDER BY block_number::NUMERIC DESC`

	rows, err := p.db.Query(ctx, query, options.ChainID, options.State)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []usecases.Block
	for rows.Next() {
		var b usecases.Block
		var blockNumberString string
		if err := rows.Scan(&b.ID, &blockNumberString, &b.State, &b.ChainID); err != nil {
			return nil, err
		}

		blockNumber := new(big.Int)
		if _, success := blockNumber.SetString(blockNumberString, 10); !success {
			return nil, errors.New("could not parse block number")
		}
		b.BlockNumber = *blockNumber
		blocks = append(blocks, b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &usecases.BlockCollection{Items: blocks}, nil
}

func (p PostgresBlockStorageAdapter) AllByAscFromBlockNumber(ctx context.Context, options usecases.BlockStorageListByAscFromBlockNumberOptions) (*usecases.BlockCollection, error) {
	rows, err := p.db.Query(ctx, `
		SELECT id, block_number, state, chain_id
		FROM blocks
		WHERE chain_id = $1 AND block_number::numeric > $2
		ORDER BY block_number::numeric ASC
	`, options.ChainID, options.BlockNumber.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []usecases.Block
	for rows.Next() {
		var b usecases.Block
		var blockNumberString string
		if err := rows.Scan(&b.ID, &blockNumberString, &b.State, &b.ChainID); err != nil {
			return nil, err
		}
		blockNumber := new(big.Int)
		if _, success := blockNumber.SetString(blockNumberString, 10); !success {
			return nil, errors.New("could not parse block number")
		}
		b.BlockNumber = *blockNumber
		blocks = append(blocks, b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &usecases.BlockCollection{Items: blocks}, nil
}

func (p PostgresBlockStorageAdapter) Get(ctx context.Context, id string) (*usecases.Block, error) {
	var block usecases.Block
	var blockNumberString string
	err := p.db.QueryRow(ctx, `
		SELECT id, block_number, state, chain_id
		FROM blocks
		WHERE id = $1
	`, id).Scan(&block.ID, &blockNumberString, &block.State, &block.ChainID)

	if err != nil {
		return nil, err
	}

	blockNumber := new(big.Int)
	blockNumber, success := blockNumber.SetString(blockNumberString, 10)
	if success == false {
		return nil, errors.New("could not parse block number")
	}
	block.BlockNumber = *blockNumber

	return &block, nil
}

func (p PostgresBlockStorageAdapter) Edit(ctx context.Context, block usecases.Block) (*usecases.Block, error) {
	_, err := p.db.Exec(ctx, `
		UPDATE blocks
		SET block_number = $2, state = $3, chain_id = $4
		WHERE id = $1
	`, block.ID, block.BlockNumber.String(), block.State, block.ChainID)

	if err != nil {
		return nil, err
	}

	return &block, nil
}
