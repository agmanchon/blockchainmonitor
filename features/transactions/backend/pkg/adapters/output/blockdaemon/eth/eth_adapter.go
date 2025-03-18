package eth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases/blockchainnetworkinfra"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/rpc"
	"github.com/agmanchon/blockchainmonitor/libs/backend/pkg/utils"
	"log"
	"math/big"
)

var _ blockchainnetworkinfra.BlockDaemonNetworkAdapter = (*DefaultEthAdapter)(nil)

type DefaultEthAdapter struct {
	rpcClient *rpc.RPCClient
	host      string
}

type DefaultEthAdapterOptions struct {
	Host string
}

func ProvideDefaultEthAdapter(options DefaultEthAdapterOptions) (*DefaultEthAdapter, error) {
	if options.Host == "nil" {
		return nil, errors.New("option host is mandatory")
	}

	rpcClient := rpc.NewRPCClient(options.Host)
	return &DefaultEthAdapter{rpcClient: rpcClient, host: options.Host}, nil
}

func (d DefaultEthAdapter) GetMostRecentBlockNumberBlockDaemon(ctx context.Context, input blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonInput) (*blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonOutput, error) {
	response, err := d.rpcClient.Call(ctx, "eth_blockNumber", nil, input.APIKey)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, errors.New("response is nil")
	}
	var hexBlockNumber string
	if err := json.Unmarshal(response, &hexBlockNumber); err != nil {
		return nil, err
	}

	blockNumber, err := utils.ConvertHexToBigInt(hexBlockNumber)
	if err != nil {
		return nil, err
	}
	return &blockchainnetworkinfra.GetMostRecentBlockNumberBlockDaemonOutput{
		BlockNumber: *blockNumber,
	}, nil

}

func (d DefaultEthAdapter) GetBlockByNumberBlockDaemon(ctx context.Context, input blockchainnetworkinfra.GetBlockByNumberBlockDaemonInput) (*blockchainnetworkinfra.GetBlockByNumberBlockDaemonOutput, error) {

	if &input.BlockNumber == nil {
		return nil, errors.New("input block number is nil")
	}
	blockNumberHex := utils.ConvertBigIntToHex(&input.BlockNumber)
	params := []interface{}{blockNumberHex, false}
	response, err := d.rpcClient.Call(ctx, "eth_getBlockByNumber", params, input.APIKey)
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

func (d DefaultEthAdapter) GetTransactionBlockDaemon(ctx context.Context, input blockchainnetworkinfra.GetTransactionBlockDaemonInput) (*blockchainnetworkinfra.GetTransactionBlockDaemonOutput, error) {
	params := []interface{}{input.ID}
	response, err := d.rpcClient.Call(ctx, "eth_getTransactionByHash", params, input.APIKey)
	if err != nil {
		return nil, err
	}

	if response == nil {
		errorMsg := fmt.Sprintf("transaction %s not found or cleaned up by the node", input.ID)
		log.Println(errorMsg)
		return &blockchainnetworkinfra.GetTransactionBlockDaemonOutput{}, nil
		//return nil, errors.New(errorMsg)
	}

	var tx Transaction
	if err := json.Unmarshal(response, &tx); err != nil {
		return nil, err
	}

	source := tx.From
	destination := tx.To
	amountWei, _ := new(big.Int).SetString(tx.Value[2:], 16)
	amountEther := new(big.Float).Quo(new(big.Float).SetInt(amountWei), big.NewFloat(1e18))

	gas := new(big.Int)
	gasPrice := new(big.Int)

	gas.SetString(tx.Gas[2:], 16)           // Strip "0x" and convert to decimal
	gasPrice.SetString(tx.GasPrice[2:], 16) // Strip "0x" and convert to decimal

	feeInWei := new(big.Int).Mul(gas, gasPrice)
	feeWeiInEther := new(big.Float).SetInt(feeInWei)
	feeEtherValue := new(big.Float).Quo(feeWeiInEther, big.NewFloat(1e18))

	fees := fmt.Sprintf("%.18f", feeEtherValue)

	output := &blockchainnetworkinfra.GetTransactionBlockDaemonOutput{}
	output.Source = source
	output.Destination = destination
	output.Amount = amountEther.String()
	output.Fees = fees

	return output, nil
}
