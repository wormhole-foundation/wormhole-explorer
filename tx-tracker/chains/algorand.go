package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
)

type algorandTransactionResponse struct {
	Transaction struct {
		ID        string `json:"id"`
		Sender    string `json:"sender"`
		RoundTime int    `json:"round-time"`
	} `json:"transaction"`
}

func fetchAlgorandTx(
	ctx context.Context,
	cfg *config.RpcProviderSettings,
	txHash string,
) (*TxDetail, error) {

	// Fetch tx data from the Algorand Indexer API
	url := fmt.Sprintf("%s/v2/transactions/%s", cfg.AlgorandBaseUrl, txHash)
	fmt.Println(url)
	body, err := httpGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to Algorand transactions endpoint failed: %w", err)
	}

	// Decode the response
	var response algorandTransactionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode Algorand transactions response as JSON: %w", err)
	}

	// Populate the result struct and return
	txDetail := TxDetail{
		NativeTxHash: response.Transaction.ID,
		From:         response.Transaction.Sender,
		Timestamp:    time.Unix(int64(response.Transaction.RoundTime), 0),
	}
	return &txDetail, nil
}
