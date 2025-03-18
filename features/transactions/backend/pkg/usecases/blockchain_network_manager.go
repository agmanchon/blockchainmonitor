package usecases

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
)

type BlockchainNetworkInfraAdapter interface {
	GetMostRecentBlockNumber(context.Context, GetMostRecentBlockNumberInput) (*GetMostRecentBlockNumberOutput, error)
	GetBlockByNumber(context.Context, GetBlockByNumberInput) (*GetBlockByNumberOutput, error)
	GetTransaction(context.Context, GetTransactionInput) (*GetTransactionOutput, error)
}

type GetMostRecentBlockNumberInput struct {
	BlockchainNetworkType string
}

type GetMostRecentBlockNumberOutput struct {
	BlockNumber big.Int
}

type GetBlockByNumberInput struct {
	BlockNumber           big.Int
	BlockchainNetworkType string
}

type GetBlockByNumberOutput struct {
	TransactionHashes []string
}

type GetTransactionInput struct {
	ID                    string
	BlockNumber           big.Int
	BlockchainNetworkType string
}

type GetTransactionOutput struct {
	Source      string
	Destination string
	Amount      string
	Fees        string
}
type BlockchainNetworkInfraFactory interface {
	Info() BlockchainNetworkInfraInfo
	NewBlockchainNetworkInfra() BlockchainNetworkInfraAdapter
}
type BlockchainNetworkInfraInfo struct {
	Name string
}

type BlockchainNetworkInfraAdapterManager interface {
	RegisterBlockchainNetworkFactory(factory BlockchainNetworkInfraFactory) error
	Use(ctx context.Context, blockchainNetworkInfraType string) (BlockchainNetworkInfraAdapter, error)
	ListBlockchainNetworkInfraAdapters(ctx context.Context) (*ListBlockchainNetworkInfraAdaptersOutput, error)
}

type ListBlockchainNetworkInfraAdaptersOutput struct {
	Items []string
}

type DefaultBlockchainNetworkInfraAdapterManager struct {
	blockchainNetworkInfraAdapters map[string]BlockchainNetworkInfraFactory
	mutex                          *sync.RWMutex
}

func ProvideDefaultBlockchainNetworkInfraAdapterManager() (*DefaultBlockchainNetworkInfraAdapterManager, error) {
	return &DefaultBlockchainNetworkInfraAdapterManager{
		blockchainNetworkInfraAdapters: make(map[string]BlockchainNetworkInfraFactory),
		mutex:                          &sync.RWMutex{},
	}, nil
}

func (d DefaultBlockchainNetworkInfraAdapterManager) RegisterBlockchainNetworkFactory(factory BlockchainNetworkInfraFactory) error {
	blockchainNetworkInfo := factory.Info()
	d.mutex.Lock()
	defer d.mutex.Unlock()
	_, ok := d.blockchainNetworkInfraAdapters[blockchainNetworkInfo.Name]
	if ok {
		return fmt.Errorf("cannot register blockchain network adapter with name [%v]", blockchainNetworkInfo.Name)
	}

	d.blockchainNetworkInfraAdapters[blockchainNetworkInfo.Name] = factory

	return nil
}

func (d DefaultBlockchainNetworkInfraAdapterManager) Use(ctx context.Context, blockchainNetworkInfraType string) (BlockchainNetworkInfraAdapter, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	factory, ok := d.blockchainNetworkInfraAdapters[blockchainNetworkInfraType]
	if !ok {
		errorMsg := fmt.Sprintf("cannot use formatter with type [%v]", blockchainNetworkInfraType)
		return nil, errors.New(errorMsg)
	}
	return factory.NewBlockchainNetworkInfra(), nil
}

func (d DefaultBlockchainNetworkInfraAdapterManager) ListBlockchainNetworkInfraAdapters(ctx context.Context) (*ListBlockchainNetworkInfraAdaptersOutput, error) {
	blockchainNetworkInfraAdapters := make([]string, 0)
	for _, adapter := range d.blockchainNetworkInfraAdapters {
		blockchainNetworkInfraAdapters = append(blockchainNetworkInfraAdapters, adapter.Info().Name)
	}
	return &ListBlockchainNetworkInfraAdaptersOutput{Items: blockchainNetworkInfraAdapters}, nil
}

var _ BlockchainNetworkInfraAdapterManager = (*DefaultBlockchainNetworkInfraAdapterManager)(nil)
