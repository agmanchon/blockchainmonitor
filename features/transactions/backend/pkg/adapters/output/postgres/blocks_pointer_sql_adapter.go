package postgres

import (
	"context"
	"errors"
	"math/big"

	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBlockPointerStorageAdapter struct {
	db *pgxpool.Pool
}

type PostgresBlockPointerStorageAdapterOptions struct {
	DB *pgxpool.Pool
}

var _ usecases.BlockPointerStorage = (*PostgresBlockPointerStorageAdapter)(nil)

func ProvidePostgresBlockPointerStorageAdapter(options PostgresBlockPointerStorageAdapterOptions) (*PostgresBlockPointerStorageAdapter, error) {
	if options.DB == nil {
		return nil, errors.New("db connection is required")
	}
	return &PostgresBlockPointerStorageAdapter{
		db: options.DB,
	}, nil
}

func (p PostgresBlockPointerStorageAdapter) Add(ctx context.Context, blockPointer usecases.BlockPointer) (*usecases.BlockPointer, error) {
	_, err := p.db.Exec(ctx, `
		INSERT INTO blocks_pointers (id, block_number, chain_id)
		VALUES ($1, $2, $3)
	`, blockPointer.ID, blockPointer.BlockNumber.String(), blockPointer.ChainID)

	if err != nil {
		return nil, err
	}
	return &blockPointer, nil
}

func (p PostgresBlockPointerStorageAdapter) Edit(ctx context.Context, blockPointer usecases.BlockPointer) (*usecases.BlockPointer, error) {
	_, err := p.db.Exec(ctx, `
		UPDATE blocks_pointers
		SET block_number = $1, chain_id = $2
		WHERE id = $3
	`, blockPointer.BlockNumber.String(), blockPointer.ChainID, blockPointer.ID)

	if err != nil {
		return nil, err
	}

	return &blockPointer, nil
}

func (p PostgresBlockPointerStorageAdapter) Get(ctx context.Context, id string) (*usecases.BlockPointer, error) {
	var blockPointer usecases.BlockPointer
	var blockNumberString string
	err := p.db.QueryRow(ctx, `
		SELECT id, block_number, chain_id
		FROM blocks_pointers
		WHERE id = $1
	`, id).Scan(&blockPointer.ID, &blockNumberString, &blockPointer.ChainID)

	if err != nil {
		return nil, err
	}

	blockNumber := new(big.Int)
	blockNumber, success := blockNumber.SetString(blockNumberString, 10)
	if success == false {
		return nil, errors.New("could not parse block number")
	}
	blockPointer.BlockNumber = *blockNumber

	return &blockPointer, nil
}
