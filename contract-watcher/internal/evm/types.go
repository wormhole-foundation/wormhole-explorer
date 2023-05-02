package evm

type getLatestBlockResponse struct {
	Result string `json:"result"`
}

type Transaction struct {
	BlockHash            string `json:"blockHash"`
	BlockNumber          string `json:"blockNumber"`
	From                 string `json:"from"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
	Hash                 string `json:"hash"`
	Input                string `json:"input"`
	Nonce                string `json:"nonce"`
	To                   string `json:"to"`
	TransactionIndex     string `json:"transactionIndex"`
	Value                string `json:"value"`
	Type                 string `json:"type"`
	AccessList           any    `json:"accessList"`
	ChainID              string `json:"chainId"`
	V                    string `json:"v"`
	R                    string `json:"r"`
	S                    string `json:"s"`
}

type GetBlockResult struct {
	Hash         string        `json:"hash"`
	Number       string        `json:"number"`
	Timestamp    string        `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
}

type getBlockResponse struct {
	Result GetBlockResult `json:"result"`
}

type getTransactionReceiptResponse struct {
	Result TransactionReceiptResult `json:"result"`
}

type TransactionReceiptResult struct {
	BlockHash         string  `json:"blockHash"`
	BlockNumber       string  `json:"blockNumber"`
	ContractAddress   *string `json:"contractAddress"`
	CumulativeGasUsed string  `json:"cumulativeGasUsed"`
	EffectiveGasPrice string  `json:"effectiveGasPrice"`
	From              string  `json:"from"`
	GasUsed           string  `json:"gasUsed"`
	Logs              []any   `json:"logs"`
	LogsBloom         string  `json:"logsBloom"`
	Status            string  `json:"status"`
	To                *string `json:"to"`
	TransactionHash   *string `json:"transactionHash"`
	TransactionIndex  *string `json:"transactionIndex"`
	Type              string  `json:"type"`
}

type EvmRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      int    `json:"id"`
}
