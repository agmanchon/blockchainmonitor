package test

import (
	"context"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/adapters/output/blockdaemon/sol"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases/blockchainnetworkinfra"
	"github.com/stretchr/testify/require"
	"testing"
)

var solEndpoint = "https://svc.blockdaemon.com/solana/mainnet/native"
var solApiKey = "zpka_1b2cf3711bc5465fad8bebd618e0e060_2e4bceae"

func TestSolAdapterOk(t *testing.T) {
	ctx := context.TODO()

	options := sol.DefaultSolAdapterOptions{}
	options.Host = solEndpoint
	solAdapter, err := sol.ProvideDefaultSolAdapter(options)
	require.Nil(t, err)

	getMostRecentBlockNumberInput := blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonInput{}
	getMostRecentBlockNumberInput.APIKey = solApiKey
	currentBlockNumber, err := solAdapter.GetMostRecentBlockNumberBlockDaemon(ctx, getMostRecentBlockNumberInput)
	require.Nil(t, err)
	require.NotNil(t, currentBlockNumber)

	getBlockByNumberInput := blockchainnetworkinfra.GetBlockByNumberBlockDaemonInput{}
	getBlockByNumberInput.APIKey = solApiKey
	getBlockByNumberInput.BlockNumber = currentBlockNumber.BlockNumber
	block, err := solAdapter.GetBlockByNumberBlockDaemon(ctx, getBlockByNumberInput)
	require.Nil(t, err)
	require.NotNil(t, block)

	for _, transactionHash := range block.TransactionHashes {
		getTransactionByHashInput := blockchainnetworkinfra.GetTransactionBlockDaemonInput{}
		getTransactionByHashInput.APIKey = solApiKey
		getTransactionByHashInput.ID = transactionHash
		transaction, err := solAdapter.GetTransactionBlockDaemon(ctx, getTransactionByHashInput)
		require.Nil(t, err)
		require.NotNil(t, transaction)
	}
}
