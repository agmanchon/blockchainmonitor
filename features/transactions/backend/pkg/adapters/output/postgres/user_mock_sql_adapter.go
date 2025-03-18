package postgres

import (
	"context"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases"
)

type PostgresUserMockStorageAdapter struct {
}

type PostgresUserMockStorageAdapterOptions struct {
}

var _ usecases.UserStorage = (*PostgresUserMockStorageAdapter)(nil)

func ProvidePostgresUserMockStorageAdapter(options PostgresUserMockStorageAdapterOptions) (*PostgresUserMockStorageAdapter, error) {

	return &PostgresUserMockStorageAdapter{}, nil
}

func (p PostgresUserMockStorageAdapter) All(ctx context.Context, options usecases.UserListOptions) (*usecases.UserCollection, error) {
	return &usecases.UserCollection{}, nil
}
