package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
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
	rateLimiter *time.Ticker,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	// Call the transaction endpoint of the Algorand Indexer REST API
	var response algorandTransactionResponse
	{
		// Perform the HTTP request
		url := fmt.Sprintf("%s/v2/transactions/%s", baseUrl, txHash)
		body, err := httpGet(ctx, rateLimiter, url)
		if err != nil {
			return nil, fmt.Errorf("HTTP request to Algorand transactions endpoint failed: %w", err)
		}

		// Decode the response
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to decode Algorand transactions response as JSON: %w", err)
		}
	}

	// Populate the result struct and return
	txDetail := TxDetail{
		NativeTxHash: response.Transaction.ID,
		From:         response.Transaction.Sender,
		Timestamp:    time.Unix(int64(response.Transaction.RoundTime), 0),
	}
	return &txDetail, nil
}
