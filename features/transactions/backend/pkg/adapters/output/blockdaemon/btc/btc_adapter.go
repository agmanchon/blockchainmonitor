package btc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases/blockchainnetworkinfra"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/rpc"
	"math"
)

var _ blockchainnetworkinfra.BlockDaemonNetworkAdapter = (*DefaultBtcAdapter)(nil)

type DefaultBtcAdapter struct {
	rpcClient *rpc.RPCClient
	host      string
}

type DefaultBtcAdapterOptions struct {
	Host string
}

func ProvideDefaultBtcAdapter(options DefaultBtcAdapterOptions) (*DefaultBtcAdapter, error) {
	if options.Host == "nil" {
		return nil, errors.New("option host is mandatory")
	}

	rpcClient := rpc.NewRPCClient(options.Host)
	return &DefaultBtcAdapter{rpcClient: rpcClient, host: options.Host}, nil
}

func (d DefaultBtcAdapter) GetMostRecentBlockNumberBlockDaemon(ctx context.Context, input blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonInput) (*blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonOutput, error) {
	/*response, err := d.rpcClient.Call(ctx, "getblockcount", nil, input.APIKey)
	if err != nil {
		return nil, err
	}
	var blockNumber big.Int
	if err := json.Unmarshal(response, &blockNumber); err != nil {
		return nil, err
	}*/

	response, err := d.rpcClient.Call(ctx, "getbestblockhash", nil, input.APIKey)
	if err != nil {
		return nil, err
	}
	var blockHash string
	if err := json.Unmarshal(response, &blockHash); err != nil {
		return nil, err
	}

	verbosity := 1
	params := []interface{}{blockHash, verbosity}
	response, err = d.rpcClient.Call(ctx, "getblock", params, input.APIKey)
	if err != nil {
		return nil, err
	}

	if response == nil {
		errorMsg := fmt.Sprintf("block hash not found %d", blockHash)
		return nil, errors.New(errorMsg)
	}

	var blockHeight BlockHeight
	if err := json.Unmarshal(response, &blockHeight); err != nil {
		return nil, err
	}

	return &blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonOutput{
		BlockNumber: blockHeight.Height,
	}, nil

}

func (d DefaultBtcAdapter) GetBlockByNumberBlockDaemon(ctx context.Context, input blockchainnetworkinfra.GetBlockByNumberBlockDaemonInput) (*blockchainnetworkinfra.GetBlockByNumberBlockDaemonOutput, error) {

	paramsGetBlockHash := []interface{}{input.BlockNumber.Int64()}
	response, err := d.rpcClient.Call(ctx, "getblockhash", paramsGetBlockHash, input.APIKey)
	if err != nil {
		return nil, err
	}
	var blockHash string
	if err := json.Unmarshal(response, &blockHash); err != nil {
		return nil, err
	}

	if response == nil {
		errorMsg := fmt.Sprintf("block not found %d", input.BlockNumber)
		return nil, errors.New(errorMsg)
	}

	verbosity := 1
	params := []interface{}{blockHash, verbosity}
	response, err = d.rpcClient.Call(ctx, "getblock", params, input.APIKey)
	if err != nil {
		return nil, err
	}

	if response == nil {
		errorMsg := fmt.Sprintf("block hash not found %d", blockHash)
		return nil, errors.New(errorMsg)
	}

	var blockTransactions BlockTransactions
	if err := json.Unmarshal(response, &blockTransactions); err != nil {
		return nil, err
	}

	return &blockchainnetworkinfra.GetBlockByNumberBlockDaemonOutput{TransactionHashes: blockTransactions.TransactionsHashes}, nil
}

func (d DefaultBtcAdapter) GetTransactionBlockDaemon(ctx context.Context, input blockchainnetworkinfra.GetTransactionBlockDaemonInput) (*blockchainnetworkinfra.GetTransactionBlockDaemonOutput, error) {
	// BTC RPC call to get the raw transaction
	params := []interface{}{input.ID, true} // true to get verbose mode
	response, err := d.rpcClient.Call(ctx, "getrawtransaction", params, input.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}

	if response == nil {
		return nil, fmt.Errorf("transaction %s not found or pruned by the node", input.ID)
	}

	// Parse the response
	var tx Transaction
	if err := json.Unmarshal(response, &tx); err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	// Collect source and destination addresses
	var source string
	var destination string

	// For simplicity, consider the first input's address as the source
	if len(tx.Vin) > 0 {
		// Check if it's a coinbase transaction (Txid is empty)
		if tx.Vin[0].Txid == "" {
			// Coinbase transaction: No source from a previous transaction
			source = "coinbase" // Mark it as a coinbase transaction
		} else {
			source = tx.Vin[0].Txid // This can be adjusted depending on the actual source of the transaction
		}
	}

	// For simplicity, consider the first output's address as the destination
	if len(tx.Vout) > 0 {
		// Convert float64 to string with 8 decimal places
		destination = fmt.Sprintf("%.8f", tx.Vout[0].Value)
	}

	// Calculate total input (satoshis)
	var totalInput float64
	for _, vin := range tx.Vin {
		// If it's a coinbase transaction, skip fetching previous transaction
		if vin.Txid == "" {
			continue // Skip for coinbase
		}

		// Retrieve the previous transaction's output value using vin.Txid
		params := []interface{}{vin.Txid, true} // Get decoded previous transaction
		rawResponse, err := d.rpcClient.Call(ctx, "getrawtransaction", params, input.APIKey)
		if err != nil {
			return nil, err
		}

		var prevTx Transaction
		if err := json.Unmarshal(rawResponse, &prevTx); err != nil {
			return nil, err
		}

		// Add value of the previous transaction's output (vin.Vout) to totalInput
		if len(prevTx.Vout) > vin.Vout {
			totalInput += prevTx.Vout[vin.Vout].Value
		}
	}

	// Calculate total output (satoshis)
	var totalOutput float64
	for _, vout := range tx.Vout {
		totalOutput += vout.Value
	}

	// Calculate fees (if it's a non-coinbase transaction)
	var fees float64
	if totalInput > 0 {
		fees = totalInput - totalOutput
	} else {
		// Coinbase transaction: Calculate the fees as totalOutput minus the block reward
		blockReward := 6.25 // Example: current Bitcoin block reward in BTC
		fees = totalOutput - blockReward
	}

	// Convert fees to BTC (1 BTC = 100,000,000 satoshis)
	feeBtc := fees
	feesFormatted := fmt.Sprintf("%.8f", feeBtc)

	// Convert amount to BTC (1 BTC = 100,000,000 satoshis)
	amount := totalOutput
	amountFormatted := fmt.Sprintf("%.8f", amount)

	// Prepare output
	output := &blockchainnetworkinfra.GetTransactionBlockDaemonOutput{
		Source:      source,
		Destination: destination,
		Amount:      amountFormatted,
		Fees:        feesFormatted,
	}

	return output, nil
}

func getBlockReward(blockHeight int64) float64 {
	// Number of halvings
	halvings := blockHeight / 210000
	// Calculate block reward
	blockReward := 50.0 / math.Pow(2, float64(halvings))
	return blockReward
}
