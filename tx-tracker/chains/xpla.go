package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
)

const (
	cosmosMsgExecuteContract = "/cosmwasm.wasm.v1.MsgExecuteContract"
)

// cosmosTxsResponse models the response body from `GET /cosmos/tx/v1beta1/txs/{hash}`
type cosmosTxsResponse struct {
	TxResponse struct {
		Tx struct {
			Body struct {
				Messages []struct {
					Type_  string `json:"@type"`
					Sender string `json:"sender"`
				} `json:"messages"`
			} `json:"body"`
		} `json:"tx"`
		Timestamp string `json:"timestamp"`
		TxHash    string `json:"txhash"`
	} `json:"tx_response"`
}

func fetchXplaTx(
	ctx context.Context,
	cfg *config.RpcProviderSettings,
	txHash string,
) (*TxDetail, error) {

	// Query the Cosmos transaction endpoint
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", cfg.XplaBaseUrl, txHash)
	body, err := httpGet(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to query cosmos tx endpoint: %w", err)
	}

	// Deserialize response body
	var response cosmosTxsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to deserialize cosmos tx response: %w", err)
	}

	// Find the sender address
	var sender string
	for i := range response.TxResponse.Tx.Body.Messages {

		msg := &response.TxResponse.Tx.Body.Messages[i]
		if msg.Type_ == cosmosMsgExecuteContract {
			sender = msg.Sender
			break
		}
	}
	if sender == "" {
		return nil, fmt.Errorf("failed to find sender address in cosmos tx response")
	}

	// Parse the timestamp
	timestamp, err := time.Parse("2006-01-02T15:04:05Z", response.TxResponse.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tx timestamp from cosmos tx response: %w", err)
	}

	// Build the result object and return
	TxDetail := &TxDetail{
		From:         sender,
		Timestamp:    timestamp,
		NativeTxHash: response.TxResponse.TxHash,
	}
	return TxDetail, nil
}
