package chains

import (
	"context"
	"fmt"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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

func fetchSuiTx(
	ctx context.Context,
	chainID sdk.ChainID,
	rpcPool map[sdk.ChainID]*pool.Pool,
	txHash string,
	logger *zap.Logger,
) (*TxDetail, error) {

	// Call the transaction endpoint of the Algorand Indexer REST API
	rpcs, err := getRpcPool(rpcPool, chainID)
	if err != nil {
		return nil, err
	}

	var reply *suiGetTransactionBlockResponse
	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		// Initialize RPC client
		client, err := rpcDialContext(ctx, rpc.Id)
		if err != nil {
			logger.Error("failed to initialize RPC client", zap.Error(err))
			continue
			//return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
		}
		defer client.Close()

		// Execute the remote procedure call
		opts := suiGetTransactionBlockOpts{ShowInput: true}
		err = client.CallContext(ctx, &reply, "sui_getTransactionBlock", txHash, opts)
		if err != nil {
			if strings.Contains(err.Error(), "Could not find the referenced transaction") {
				return nil, ErrTransactionNotFound
			}
			return nil, fmt.Errorf("failed to get tx by hash: %w", err)
		}
		break
	}

	// Populate the response struct and return
	txDetail := TxDetail{
		NativeTxHash: reply.Digest,
		From:         reply.Transaction.Data.Sender,
	}
	return &txDetail, nil
}
