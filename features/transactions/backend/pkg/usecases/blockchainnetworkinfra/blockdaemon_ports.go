package blockchainnetworkinfra

import (
	"context"
	"math/big"
)

type BlockDaemonNetworkAdapter interface {
	GetMostRecentBlockNumberBlockDaemon(context.Context, GetMostRecentBlockNumberBlockDaemonInput) (*GetMostRecentBlockNumberBlockDaemonOutput, error)
	GetBlockByNumberBlockDaemon(context.Context, GetBlockByNumberBlockDaemonInput) (*GetBlockByNumberBlockDaemonOutput, error)
	GetTransactionBlockDaemon(context.Context, GetTransactionBlockDaemonInput) (*GetTransactionBlockDaemonOutput, error)
}

type GetMostRecentBlockNumberBlockDaemonInput struct {
	APIKey string
}

type GetMostRecentBlockNumberBlockDaemonOutput struct {
	BlockNumber big.Int
}

type GetBlockByNumberBlockDaemonInput struct {
	APIKey      string
	BlockNumber big.Int
}

type GetBlockByNumberBlockDaemonOutput struct {
	TransactionHashes []string
}

type GetTransactionBlockDaemonInput struct {
	APIKey      string
	ID          string
	BlockNumber big.Int
}

type GetTransactionBlockDaemonOutput struct {
	Source      string
	Destination string
	Amount      string
	Fees        string
}
