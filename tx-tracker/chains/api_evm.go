package chains

import (
	"context"
	"fmt"
	"strings"
	"time"
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
	client, err := rpcDialContext(ctx, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	// query transaction data
	var txReply ethGetTransactionByHashResponse
	{
		err = client.CallContext(ctx, rateLimiter, &txReply, "eth_getTransactionByHash", "0x"+txHash)
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
		blkParams := []interface{}{
			txReply.BlockHash, // tx hash
			false,             // include transactions?
		}
		err = client.CallContext(ctx, rateLimiter, &blkReply, "eth_getBlockByHash", blkParams...)
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
