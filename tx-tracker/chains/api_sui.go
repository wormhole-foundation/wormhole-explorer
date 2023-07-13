package chains

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
)

type suiGetTransactionBlockResponse struct {
	Digest      string `json:"digest"`
	TimestampMs int64  `json:"timestampMs,string"`
	Transaction struct {
		Data struct {
			Sender string `json:"sender"`
		} `json:"data"`
	} `json:"transaction"`
}

type suiGetTransactionBlockOpts struct {
	ShowInput          bool `json:"showInput"`
	ShowRawInput       bool `json:"showRawInput"`
	ShowEffects        bool `json:"showEffects"`
	ShowEvents         bool `json:"showEvents"`
	ShowObjectChanges  bool `json:"showObjectChanges"`
	ShowBalanceChanges bool `json:"showBalanceChanges"`
}

func fetchSuiTx(
	ctx context.Context,
	rateLimiter *time.Ticker,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	// Initialize RPC client
	client, err := rpc.DialContext(ctx, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	// Query transaction data
	var reply suiGetTransactionBlockResponse
	{
		// Wait for the rate limiter
		if !waitForRateLimiter(ctx, rateLimiter) {
			return nil, ctx.Err()
		}

		// Execute the remote procedure call
		opts := suiGetTransactionBlockOpts{ShowInput: true}
		err = client.CallContext(ctx, &reply, "sui_getTransactionBlock", txHash, opts)
		if strings.Contains(err.Error(), "Could not find the referenced transaction") {
			return nil, ErrTransactionNotFound
		} else if err != nil {
			return nil, fmt.Errorf("failed to get tx by hash: %w", err)
		}
	}

	// Populate the response struct and return
	txDetail := TxDetail{
		NativeTxHash: reply.Digest,
		From:         reply.Transaction.Data.Sender,
	}
	return &txDetail, nil
}
