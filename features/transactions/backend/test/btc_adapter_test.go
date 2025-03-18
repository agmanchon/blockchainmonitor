package test

import (
	"context"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/adapters/output/blockdaemon/btc"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases/blockchainnetworkinfra"
	"github.com/stretchr/testify/require"
	"testing"
)

var btcEndpoint = "https://svc.blockdaemon.com/bitcoin/mainnet/native"
var btcApiKey = "zpka_1b2cf3711bc5465fad8bebd618e0e060_2e4bceae"

func TestBtcAdapterOk(t *testing.T) {
	ctx := context.TODO()

	options := btc.DefaultBtcAdapterOptions{}
	options.Host = btcEndpoint
	btcAdapter, err := btc.ProvideDefaultBtcAdapter(options)
	require.Nil(t, err)

	getMostRecentBlockNumberInput := blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonInput{}
	getMostRecentBlockNumberInput.APIKey = btcApiKey
	currentBlockNumber, err := btcAdapter.GetMostRecentBlockNumberBlockDaemon(ctx, getMostRecentBlockNumberInput)
	require.Nil(t, err)
	require.NotNil(t, currentBlockNumber)

	getBlockByNumberInput := blockchainnetworkinfra.GetBlockByNumberBlockDaemonInput{}
	getBlockByNumberInput.APIKey = btcApiKey
	getBlockByNumberInput.BlockNumber = currentBlockNumber.BlockNumber
	block, err := btcAdapter.GetBlockByNumberBlockDaemon(ctx, getBlockByNumberInput)
	require.Nil(t, err)
	require.NotNil(t, block)

	for _, transactionHash := range block.TransactionHashes {
		getTransactionByHashInput := blockchainnetworkinfra.GetTransactionBlockDaemonInput{}
		getTransactionByHashInput.APIKey = btcApiKey
		getTransactionByHashInput.ID = transactionHash
		transaction, err := btcAdapter.GetTransactionBlockDaemon(ctx, getTransactionByHashInput)
		require.Nil(t, err)
		require.NotNil(t, transaction)
	}
}
