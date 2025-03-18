package usecases

import (
	"context"
	"math/big"
)

type BlockPointerStorage interface {
	Add(ctx context.Context, block BlockPointer) (*BlockPointer, error)
	Get(ctx context.Context, id string) (*BlockPointer, error)
	Edit(ctx context.Context, blockPointer BlockPointer) (*BlockPointer, error)
}

type BlockPointer struct {
	ID          string
	ChainID     string
	BlockNumber big.Int
}

type BlockStorage interface {
	Add(ctx context.Context, block Block) (*Block, error)
	AllByDescOrder(ctx context.Context, options BlockStorageListOptions) (*BlockCollection, error)
	AllByAscFromBlockNumber(ctx context.Context, options BlockStorageListByAscFromBlockNumberOptions) (*BlockCollection, error)
	Get(ctx context.Context, id string) (*Block, error)
	Edit(ctx context.Context, block Block) (*Block, error)
}

type Block struct {
	ID          string
	BlockNumber big.Int
	State       string
	ChainID     string
}

type BlockCollection struct {
	Items []Block
}

type BlockStorageListOptions struct {
	ChainID string
	State   string
}
type BlockStorageListByAscFromBlockNumberOptions struct {
	ChainID     string
	BlockNumber big.Int
}

type UserAddressesStorage interface {
	All(ctx context.Context, options UserAddresessListOptions) (*UserAddressCollection, error)
}
type UserAddresessListOptions struct {
	UserID string
}

type UserAddressCollection struct {
	Items []string
}

type UserStorage interface {
	All(ctx context.Context, options UserListOptions) (*UserCollection, error)
}
type UserListOptions struct {
}

type UserCollection struct {
	Items []string
}
