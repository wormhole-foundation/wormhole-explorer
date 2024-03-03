package chains

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"go.uber.org/zap"
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

func FetchSuiTx(
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
		txDetail, err = fetchSuiTx(ctx, rpc.Id, txHash)
		if err != nil {
			logger.Debug("Failed to fetch transaction from SUI node", zap.String("url", rpc.Id), zap.Error(err))
			continue
		}
		return txDetail, nil
	}
	return txDetail, err
}

func fetchSuiTx(
	ctx context.Context,
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
		// Execute the remote procedure call
		opts := suiGetTransactionBlockOpts{ShowInput: true}
		err = client.CallContext(ctx, &reply, "sui_getTransactionBlock", txHash, opts)
		if err != nil {
			if strings.Contains(err.Error(), "Could not find the referenced transaction") {
				return nil, ErrTransactionNotFound
			}
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
