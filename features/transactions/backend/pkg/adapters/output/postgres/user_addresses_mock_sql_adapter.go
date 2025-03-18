package postgres

import (
	"context"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases"
)

type PostgresUserAddressesMockStorageAdapter struct {
}

type PostgresUserAddressesMockStorageAdapterOptions struct {
}

var _ usecases.UserAddressesStorage = (*PostgresUserAddressesMockStorageAdapter)(nil)

func ProvidePostgresUserAddressesStorageAdapter(options PostgresUserAddressesMockStorageAdapterOptions) (*PostgresUserAddressesMockStorageAdapter, error) {

	return &PostgresUserAddressesMockStorageAdapter{}, nil
}

func (p PostgresUserAddressesMockStorageAdapter) All(ctx context.Context, options usecases.UserAddresessListOptions) (*usecases.UserAddressCollection, error) {
	return &usecases.UserAddressCollection{}, nil
}
