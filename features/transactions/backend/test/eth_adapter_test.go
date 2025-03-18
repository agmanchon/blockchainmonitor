package test

import (
	"context"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/adapters/output/blockdaemon/eth"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases/blockchainnetworkinfra"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var ethEndpoint = "https://svc.blockdaemon.com/ethereum/mainnet/native"
var ethApiKey = "zpka_1b2cf3711bc5465fad8bebd618e0e060_2e4bceae"

func TestEthAdapterOk(t *testing.T) {
	ctx := context.TODO()

	options := eth.DefaultEthAdapterOptions{}
	options.Host = ethEndpoint
	ethAdapter, err := eth.ProvideDefaultEthAdapter(options)
	require.Nil(t, err)

	getMostRecentBlockNumberInput := blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonInput{}
	getMostRecentBlockNumberInput.APIKey = ethApiKey
	currentBlockNumber, err := ethAdapter.GetMostRecentBlockNumberBlockDaemon(ctx, getMostRecentBlockNumberInput)
	require.Nil(t, err)
	require.NotNil(t, currentBlockNumber)

	getBlockByNumberInput := blockchainnetworkinfra.GetBlockByNumberBlockDaemonInput{}
	getBlockByNumberInput.APIKey = ethApiKey
	getBlockByNumberInput.BlockNumber = currentBlockNumber.BlockNumber
	block, err := ethAdapter.GetBlockByNumberBlockDaemon(ctx, getBlockByNumberInput)
	require.Nil(t, err)
	require.NotNil(t, block)

	for _, transactionHash := range block.TransactionHashes {
		getTransactionByHashInput := blockchainnetworkinfra.GetTransactionBlockDaemonInput{}
		getTransactionByHashInput.APIKey = ethApiKey
		getTransactionByHashInput.ID = transactionHash
		transaction, err := ethAdapter.GetTransactionBlockDaemon(ctx, getTransactionByHashInput)
		require.Nil(t, err)
		require.NotNil(t, transaction)
		time.Sleep(100 * time.Millisecond)
	}
}
