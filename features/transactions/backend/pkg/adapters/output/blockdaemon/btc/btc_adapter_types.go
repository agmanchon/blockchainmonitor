package btc

import "math/big"

type BlockTransactions struct {
	TransactionsHashes []string `json:"tx"`
}

type Transaction struct {
	Txid string `json:"txid"`
	Vin  []Vin  `json:"vin"`
	Vout []Vout `json:"vout"`
}

type Vin struct {
	Txid  string  `json:"txid"`
	Vout  int     `json:"vout"`
	Addr  string  `json:"address,omitempty"` // Input address (optional, depends on node verbosity)
	Value float64 `json:"value,omitempty"`   // Input value (optional, depends on node verbosity)
}

type Vout struct {
	Value        float64      `json:"value"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Address string `json:"address"`
}

type BlockHeight struct {
	Height big.Int `json:"height"`
}
