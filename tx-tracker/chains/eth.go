package chains

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
)

type ethGetTransactionByHashResponse struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	To          string `json:"to"`
}

type ethGetBlockByHashResponse struct {
	Timestamp string `json:"timestamp"`
	Number    string `json:"number"`
}

func fetchEthTx(
	ctx context.Context,
	rateLimiter *time.Ticker,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	// initialize RPC client
	client, err := rpc.DialContext(ctx, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	// query transaction data
	var txReply ethGetTransactionByHashResponse
	{
		// Wait for the rate limiter
		if !waitForRateLimiter(ctx, rateLimiter) {
			return nil, ctx.Err()
		}

		// Call the RPC method
		err = client.CallContext(ctx, &txReply, "eth_getTransactionByHash", "0x"+txHash)
		if err != nil {
			return nil, fmt.Errorf("failed to get tx by hash: %w", err)
		}
		if txReply.BlockHash == "" || txReply.From == "" {
			return nil, ErrTransactionNotFound
		}
	}

	// query block data
	var blkReply ethGetBlockByHashResponse
	{
		// Wait for the rate limiter
		if !waitForRateLimiter(ctx, rateLimiter) {
			return nil, ctx.Err()
		}

		// Call the RPC method
		blkParams := []interface{}{
			txReply.BlockHash, // tx hash
			false,             // include transactions?
		}
		err = client.CallContext(ctx, &blkReply, "eth_getBlockByHash", blkParams...)
		if err != nil {
			return nil, fmt.Errorf("failed to get block by hash: %w", err)
		}
	}

	// parse transaction timestamp
	timestamp, err := timestampFromHex(blkReply.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse block timestamp: %w", err)
	}

	// build results and return
	txDetail := &TxDetail{
		From:         strings.ToLower(txReply.From),
		Timestamp:    timestamp,
		NativeTxHash: fmt.Sprintf("0x%s", strings.ToLower(txHash)),
	}
	return txDetail, nil
}
