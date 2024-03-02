package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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

func fetchCosmosTx(
	ctx context.Context,
	chainID sdk.ChainID,
	rpcPool map[sdk.ChainID]*pool.Pool,
	txHash string,
	logger *zap.Logger,
) (*TxDetail, error) {

	// Get the RPC pool for the chain
	rpcs, err := getRpcPool(rpcPool, chainID)
	if err != nil {
		return nil, err
	}

	// Call the transaction endpoint of the cosmos REST API
	var response *cosmosTxsResponse

	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		// Perform the HTTP request
		uri := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", rpc.Id, txHash)
		body, err := httpGet(ctx, uri)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				logger.Debug("cosmos tx not found", zap.String("txHash", txHash))
				continue
				//return nil, ErrTransactionNotFound
			}
			logger.Debug("failed to query cosmos tx endpoint", zap.Error(err), zap.String("rpc", rpc.Id))
			continue
			//return nil, fmt.Errorf("failed to query cosmos tx endpoint: %w", err)
		}

		// Deserialize response body
		if err := json.Unmarshal(body, &response); err != nil {
			logger.Debug("failed to deserialize cosmos tx response", zap.Error(err), zap.String("rpc", rpc.Id))
			continue
			//return nil, fmt.Errorf("failed to deserialize cosmos tx response: %w", err)
		}
		fmt.Printf("rpc.Id = %s \n", rpc.Id)
		break
	}

	// Check if the transaction was found
	if response == nil {
		return nil, ErrTransactionNotFound
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
