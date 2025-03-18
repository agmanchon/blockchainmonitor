package blockchainnetworkinfra

import (
	"context"
	"errors"
	"fmt"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases"
)

var _ usecases.BlockchainNetworkInfraFactory = (*BlockDaemonNetworkFactory)(nil)

type BlockDaemonNetworkFactory struct {
	NewtworkAdapters map[string]BlockDaemonNetworkAdapter
	ApiKey           string
}

func (f BlockDaemonNetworkFactory) Info() usecases.BlockchainNetworkInfraInfo {
	return usecases.BlockchainNetworkInfraInfo{Name: BlockDaemonInfraType}
}

type BlockDaemonNetworkFactoryOptions struct {
	APIKey string
}

func NewBlockDaemonNetworkFactory(options BlockDaemonNetworkFactoryOptions) (*BlockDaemonNetworkFactory, error) {

	return &BlockDaemonNetworkFactory{
		NewtworkAdapters: make(map[string]BlockDaemonNetworkAdapter),
		ApiKey:           options.APIKey,
	}, nil
}

func (f BlockDaemonNetworkFactory) RegisterBlockDaemonNetwork(k string, a BlockDaemonNetworkAdapter) {
	f.NewtworkAdapters[k] = a
}

func (f BlockDaemonNetworkFactory) NewBlockchainNetworkInfra() usecases.BlockchainNetworkInfraAdapter {

	return &BlockDaemonNetwork{
		blockchainNetworkAdapters: f.NewtworkAdapters,
		apiKey:                    f.ApiKey,
	}
}

var _ usecases.BlockchainNetworkInfraAdapter = (*BlockDaemonNetwork)(nil)

type BlockDaemonNetwork struct {
	blockchainNetworkAdapters map[string]BlockDaemonNetworkAdapter
	apiKey                    string
}

func (b BlockDaemonNetwork) GetMostRecentBlockNumber(ctx context.Context, input usecases.GetMostRecentBlockNumberInput) (*usecases.GetMostRecentBlockNumberOutput, error) {
	blockchainNetworkAdapter, ok := b.blockchainNetworkAdapters[input.BlockchainNetworkType]
	if !ok {
		msgError := fmt.Sprintf("blockchain network adapter [%s] is not registered", input.BlockchainNetworkType)
		return nil, errors.New(msgError)
	}

	getMostRecentBlockNumberInput := GetMostRecentBlockNumberBlockDaemonInput{}
	getMostRecentBlockNumberInput.APIKey = b.apiKey
	getMostRecentBlockNumberBlockDaemonOutput, err := blockchainNetworkAdapter.GetMostRecentBlockNumberBlockDaemon(ctx, getMostRecentBlockNumberInput)
	if err != nil {
		return nil, err
	}
	return &usecases.GetMostRecentBlockNumberOutput{BlockNumber: getMostRecentBlockNumberBlockDaemonOutput.BlockNumber}, nil
}

func (b BlockDaemonNetwork) GetBlockByNumber(ctx context.Context, input usecases.GetBlockByNumberInput) (*usecases.GetBlockByNumberOutput, error) {
	blockchainNetworkAdapter, ok := b.blockchainNetworkAdapters[input.BlockchainNetworkType]
	if !ok {
		msgError := fmt.Sprintf("blockchain network adapter [%s] is not registered", input.BlockchainNetworkType)
		return nil, errors.New(msgError)
	}

	getBlockByNumberBlockDaemonInput := GetBlockByNumberBlockDaemonInput{}
	getBlockByNumberBlockDaemonInput.APIKey = b.apiKey
	getBlockByNumberBlockDaemonInput.BlockNumber = input.BlockNumber
	getBlockByNumberBlockDaemonOutput, err := blockchainNetworkAdapter.GetBlockByNumberBlockDaemon(ctx, getBlockByNumberBlockDaemonInput)
	if err != nil {
		return nil, err
	}
	return &usecases.GetBlockByNumberOutput{TransactionHashes: getBlockByNumberBlockDaemonOutput.TransactionHashes}, nil
}

func (b BlockDaemonNetwork) GetTransaction(ctx context.Context, input usecases.GetTransactionInput) (*usecases.GetTransactionOutput, error) {
	blockchainNetworkAdapter, ok := b.blockchainNetworkAdapters[input.BlockchainNetworkType]
	if !ok {
		msgError := fmt.Sprintf("blockchain network adapter [%s] is not registered", input.BlockchainNetworkType)
		return nil, errors.New(msgError)
	}

	getTransactionBlockDaemonInput := GetTransactionBlockDaemonInput{}
	getTransactionBlockDaemonInput.APIKey = b.apiKey
	getTransactionBlockDaemonInput.BlockNumber = input.BlockNumber
	getTransactionBlockDaemonInput.ID = input.ID
	getTransactionBlockDaemonOutput, err := blockchainNetworkAdapter.GetTransactionBlockDaemon(ctx, getTransactionBlockDaemonInput)
	if err != nil {
		return nil, err
	}
	return &usecases.GetTransactionOutput{
		Source:      getTransactionBlockDaemonOutput.Source,
		Destination: getTransactionBlockDaemonOutput.Destination,
		Amount:      getTransactionBlockDaemonOutput.Amount,
		Fees:        getTransactionBlockDaemonOutput.Fees,
	}, nil
}
