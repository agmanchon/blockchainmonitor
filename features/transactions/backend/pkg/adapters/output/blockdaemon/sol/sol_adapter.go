package sol

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases/blockchainnetworkinfra"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/rpc"
	"math/big"
)

var _ blockchainnetworkinfra.BlockDaemonNetworkAdapter = (*DefaultSolAdapter)(nil)

type DefaultSolAdapter struct {
	rpcClient *rpc.RPCClient
	host      string
}

type DefaultSolAdapterOptions struct {
	Host string
}

func ProvideDefaultSolAdapter(options DefaultSolAdapterOptions) (*DefaultSolAdapter, error) {
	if options.Host == "nil" {
		return nil, errors.New("option host is mandatory")
	}

	rpcClient := rpc.NewRPCClient(options.Host)
	return &DefaultSolAdapter{rpcClient: rpcClient, host: options.Host}, nil
}

func (d DefaultSolAdapter) GetMostRecentBlockNumberBlockDaemon(ctx context.Context, input blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonInput) (*blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonOutput, error) {
	response, err := d.rpcClient.Call(ctx, "getSlot", nil, input.APIKey)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("response is nil")
	}
	var blockNumber big.Int
	if err := json.Unmarshal(response, &blockNumber); err != nil {
		return nil, err
	}

	/*blockNumber := new(big.Int)
	if _, ok := blockNumber.SetString(blockNumberStr, 10); !ok {
		return nil, errors.New("failed to parse block number to big.Int")
	}*/

	return &blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonOutput{
		BlockNumber: blockNumber,
	}, nil

}

func (d DefaultSolAdapter) GetBlockByNumberBlockDaemon(ctx context.Context, input blockchainnetworkinfra.GetBlockByNumberBlockDaemonInput) (*blockchainnetworkinfra.GetBlockByNumberBlockDaemonOutput, error) {

	requestParams := requestGetBlockData{}
	requestParams.Encoding = "json"
	requestParams.TransactionDetails = "signatures"
	requestParams.MaxSupportedTransactionVersion = 0
	params := []interface{}{input.BlockNumber.Uint64(), requestParams}
	response, err := d.rpcClient.Call(ctx, "getBlock", params, input.APIKey)
	if err != nil {
		return nil, err
	}
	if response == nil {
		errorMsg := fmt.Sprintf("block not found %d", input.BlockNumber)
		return nil, errors.New(errorMsg)
	}

	var blockTransactions BlockTransactions
	if err := json.Unmarshal(response, &blockTransactions); err != nil {
		return nil, err
	}

	return &blockchainnetworkinfra.GetBlockByNumberBlockDaemonOutput{TransactionHashes: blockTransactions.TransactionsHashes}, nil
}

func (d DefaultSolAdapter) GetTransactionBlockDaemon(ctx context.Context, input blockchainnetworkinfra.GetTransactionBlockDaemonInput) (*blockchainnetworkinfra.GetTransactionBlockDaemonOutput, error) {
	requestParams := requestGetBlockData{}
	requestParams.Encoding = "json"
	requestParams.MaxSupportedTransactionVersion = 0
	params := []interface{}{input.ID, requestParams}
	response, err := d.rpcClient.Call(ctx, "getTransaction", params, input.APIKey)
	if err != nil {
		return nil, err
	}
	if response == nil {
		errorMsg := fmt.Sprintf("transaction %s not found or cleaned up by the node", input.ID)
		return nil, errors.New(errorMsg)
	}

	var tx Transaction
	if err := json.Unmarshal(response, &tx); err != nil {
		return nil, err
	}

	// Source is typically the first account in the accountKeys array
	source := tx.Transaction.Message.AccountKeys[0]
	// Destination is typically the second account in the accountKeys array
	destination := tx.Transaction.Message.AccountKeys[1]

	// Calculate the amount as the difference between the pre- and post-balances for the source account
	amount := float64(tx.Meta.PreBalances[0]-tx.Meta.PostBalances[0]) / 1e9 // Assuming SOL amount is in lamports, divide by 1e9

	// Calculate the fee (fees are given directly in the meta data)
	fee := float64(tx.Meta.Fee) / 1e9 // Convert fee from lamports to SOL

	// Prepare output
	output := &blockchainnetworkinfra.GetTransactionBlockDaemonOutput{}
	output.Source = source
	output.Destination = destination
	output.Amount = fmt.Sprintf("%.9f", amount) // Format amount with 9 decimals for SOL
	output.Fees = fmt.Sprintf("%.9f", fee)      // Format fee with 9 decimals for SOL

	return output, nil
}
