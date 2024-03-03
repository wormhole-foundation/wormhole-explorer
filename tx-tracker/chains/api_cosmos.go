package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"go.uber.org/zap"
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

func FetchCosmosTx(
	ctx context.Context,
	pool *pool.Pool,
	txHash string,
	logger *zap.Logger,
) (*TxDetail, error) {

	// get rpc sorted by score and priority.
	rpcs := pool.GetItems()
	if len(rpcs) == 0 {
		return nil, ErrChainNotSupported
	}

	var txDetail *TxDetail
	var err error
	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		txDetail, err = fetchCosmosTx(ctx, rpc.Id, txHash)
		if err != nil {
			logger.Debug("Failed to fetch transaction from cosmos node", zap.String("url", rpc.Id), zap.Error(err))
			continue
		}
		break
	}
	return txDetail, err
}

func fetchCosmosTx(
	ctx context.Context,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	// Call the transaction endpoint of the cosmos REST API
	var response cosmosTxsResponse
	{
		// Perform the HTTP request
		uri := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", baseUrl, txHash)
		body, err := httpGet(ctx, uri)
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
