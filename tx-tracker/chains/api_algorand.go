package chains

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
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

func fetchAlgorandTx(
	ctx context.Context,
	chainID sdk.ChainID,
	rpcPool map[sdk.ChainID]*pool.Pool,
	txHash string,
	logger *zap.Logger,
) (*TxDetail, error) {

	// Call the transaction endpoint of the Algorand Indexer REST API
	var response *algorandTransactionResponse
	rpcs, err := getRpcPool(rpcPool, chainID)
	if err != nil {
		return nil, err
	}

	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		// Make the HTTP request
		url := fmt.Sprintf("%s/v2/transactions/%s", rpc.Id, txHash)
		body, err := httpGet(ctx, url)
		if err != nil {
			logger.Error("HTTP request to Algorand transactions endpoint failed", zap.Error(err), zap.String("url", url))
			continue
		}

		// Decode the response
		err = json.Unmarshal(body, &response)
		if err == nil {
			// If the response is not nil, break the loop
			break
		} else {
			logger.Error("Failed to decode Algorand transactions response as JSON", zap.Error(err), zap.String("url", url))
			continue
		}

	}

	if response == nil {
		return nil, fmt.Errorf("failed to fetch transaction from Algorand indexer")
	}

	// Populate the result struct and return
	txDetail := TxDetail{
		NativeTxHash: response.Transaction.ID,
		From:         response.Transaction.Sender,
	}
	return &txDetail, nil
}
