package eth

type BlockTransactions struct {
	TransactionsHashes []string `json:"transactions"`
}

type Transaction struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
}
