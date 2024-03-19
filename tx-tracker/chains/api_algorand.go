package chains

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type algorandTransactionResponse struct {
	Transaction struct {
		ID        string `json:"id"`
		Sender    string `json:"sender"`
		RoundTime int    `json:"round-time"`
	} `json:"transaction"`
}

func FetchAlgorandTx(
	ctx context.Context,
	pool *pool.Pool,
	txHash string,
	metrics metrics.Metrics,
	logger *zap.Logger,
) (*TxDetail, error) {

	// get rpc sorted by score and priority.
	rpcs := pool.GetItems()
	if len(rpcs) == 0 {
		return nil, ErrChainNotSupported
	}

	// Call the transaction endpoint of the Algorand Indexer REST API
	var txDetail *TxDetail
	var err error
	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		txDetail, err = fetchAlgorandTx(ctx, rpc.Id, txHash)
		if txDetail != nil {
			metrics.IncCallRpcSuccess(uint16(sdk.ChainIDAlgorand), rpc.Description)
			break
		}
		if err != nil {
			metrics.IncCallRpcError(uint16(sdk.ChainIDAlgorand), rpc.Description)
			logger.Debug("Failed to fetch transaction from Algorand indexer", zap.String("url", rpc.Id), zap.Error(err))
		}
	}

	return txDetail, err
}

func fetchAlgorandTx(
	ctx context.Context,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	// Call the transaction endpoint of the Algorand Indexer REST API
	var response algorandTransactionResponse
	{
		// Perform the HTTP request
		url := fmt.Sprintf("%s/v2/transactions/%s", baseUrl, txHash)
		body, err := httpGet(ctx, url)
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
	}
	return &txDetail, nil
}
