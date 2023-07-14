package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	cosmosMsgExecuteContract    = "/cosmwasm.wasm.v1.MsgExecuteContract"
	injectiveMsgExecuteContract = "/injective.wasmx.v1.MsgExecuteContractCompat"
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

func fetchCosmosTx(
	ctx context.Context,
	rateLimiter *time.Ticker,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	// Call the transaction endpoint of the cosmos REST API
	var response cosmosTxsResponse
	{
		// Perform the HTTP request
		uri := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", baseUrl, txHash)
		body, err := httpGet(ctx, rateLimiter, uri)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				return nil, ErrTransactionNotFound
			}
			return nil, fmt.Errorf("failed to query cosmos tx endpoint: %w", err)
		}

		// Deserialize response body
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to deserialize cosmos tx response: %w", err)
		}
	}

	// Find the sender address
	var sender string
	for i := range response.TxResponse.Tx.Body.Messages {
		msg := &response.TxResponse.Tx.Body.Messages[i]

		if msg.Type_ == cosmosMsgExecuteContract || msg.Type_ == injectiveMsgExecuteContract {
			sender = msg.Sender
			break
		}
	}
	if sender == "" {
		return nil, fmt.Errorf("failed to find sender address in cosmos tx response")
	}

	// Build the result object and return
	TxDetail := &TxDetail{
		From:         sender,
		NativeTxHash: response.TxResponse.TxHash,
	}
	return TxDetail, nil
}
