package sol

type BlockTransactions struct {
	TransactionsHashes []string `json:"signatures"`
}

type Transaction struct {
	Meta struct {
		Fee          int64   `json:"fee"`
		PreBalances  []int64 `json:"preBalances"`
		PostBalances []int64 `json:"postBalances"`
	} `json:"meta"`
	Transaction struct {
		Message struct {
			AccountKeys []string `json:"accountKeys"`
		} `json:"message"`
	} `json:"transaction"`
}

type requestGetBlockData struct {
	Encoding                       string `json:"encoding"`
	MaxSupportedTransactionVersion int    `json:"maxSupportedTransactionVersion"`
	TransactionDetails             string `json:"transactionDetails"`
	Rewards                        bool   `json:"rewards"`
}
