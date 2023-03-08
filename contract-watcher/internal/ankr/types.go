package ankr

type MaultichainOption func(*TransactionsByAddressRequest)

type TransactionsByAddressRequest struct {
	ID           int64        `json:"id"`
	Jsonrpc      string       `json:"jsonrpc"`
	Method       string       `json:"method"`
	RquestParams RquestParams `json:"params"`
}

func WithBlochchain(blockchain string) MaultichainOption {
	return func(h *TransactionsByAddressRequest) {
		h.RquestParams.Blockchain = blockchain
	}
}

func WithContract(address string) MaultichainOption {
	return func(h *TransactionsByAddressRequest) {
		h.RquestParams.Address = address
	}
}

func WithBlocks(fromBlock int64, toBlock int64) MaultichainOption {
	return func(h *TransactionsByAddressRequest) {
		h.RquestParams.FromBlock = fromBlock
		h.RquestParams.ToBlock = toBlock
	}
}

func NewTransactionsByAddressRequest(opts ...MaultichainOption) *TransactionsByAddressRequest {
	const (
		defaultMethod = "ankr_getTransactionsByAddress"
	)

	h := &TransactionsByAddressRequest{
		ID:      1,
		Jsonrpc: "2.0",
		Method:  defaultMethod,
		RquestParams: RquestParams{
			DescOrder: true,
		},
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

type RquestParams struct {
	Address       string `json:"address"`
	Blockchain    string `json:"blockchain"`
	FromBlock     int64  `json:"fromBlock,omitempty"`
	ToBlock       int64  `json:"toBlock,omitempty"`
	FromTimestamp int64  `json:"fromTimestamp,omitempty"`
	ToTimestamp   int64  `json:"toTimestamp,omitempty"`
	IncludeLogs   bool   `json:"includeLogs,omitempty"`
	DescOrder     bool   `json:"descOrder,omitempty"`
	PageSize      int64  `json:"pageSize,omitempty"`
	PageToken     string `json:"pageToken,omitempty"`
}

type TransactionsByAddressResponse struct {
	ID      int64  `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		NextPageToken string `json:"nextPageToken"`
		Transactions  []struct {
			BlockHash         string `json:"blockHash"`
			BlockNumber       string `json:"blockNumber"`
			Blockchain        string `json:"blockchain"`
			CumulativeGasUsed string `json:"cumulativeGasUsed"`
			From              string `json:"from"`
			Gas               string `json:"gas"`
			GasPrice          string `json:"gasPrice"`
			GasUsed           string `json:"gasUsed"`
			Hash              string `json:"hash"`
			Input             string `json:"input"`
			Logs              []struct {
				Address          string   `json:"address"`
				BlockHash        string   `json:"blockHash"`
				BlockNumber      string   `json:"blockNumber"`
				Blockchain       string   `json:"blockchain"`
				Data             string   `json:"data"`
				LogIndex         string   `json:"logIndex"`
				Removed          bool     `json:"removed"`
				Topics           []string `json:"topics"`
				TransactionHash  string   `json:"transactionHash"`
				TransactionIndex string   `json:"transactionIndex"`
			} `json:"logs"`
			Nonce            string `json:"nonce"`
			R                string `json:"r"`
			S                string `json:"s"`
			Status           string `json:"status"`
			Timestamp        string `json:"timestamp"`
			To               string `json:"to"`
			TransactionIndex string `json:"transactionIndex"`
			Type             string `json:"type"`
			V                string `json:"v"`
			Value            string `json:"value"`
		} `json:"transactions"`
	} `json:"result"`
}

type BlockchainStatsResponse struct {
	ID      int64  `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Stats []struct {
			BlockTimeMs            int64  `json:"blockTimeMs"`
			Blockchain             string `json:"blockchain"`
			LatestBlockNumber      int64  `json:"latestBlockNumber"`
			NativeCoinUsdPrice     string `json:"nativeCoinUsdPrice"`
			TotalEventsCount       int64  `json:"totalEventsCount"`
			TotalTransactionsCount int64  `json:"totalTransactionsCount"`
		} `json:"stats"`
	} `json:"result"`
}
