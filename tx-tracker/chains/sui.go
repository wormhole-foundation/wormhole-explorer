package chains

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
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
	cfg *config.RpcProviderSettings,
	txHash string,
) (*TxDetail, error) {

	// Initialize RPC client
	client, err := rpc.DialContext(ctx, cfg.SuiBaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	// Query transaction data
	var reply suiGetTransactionBlockResponse
	opts := suiGetTransactionBlockOpts{ShowInput: true}
	err = client.CallContext(ctx, &reply, "sui_getTransactionBlock", txHash, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get tx by hash: %w", err)
	}

	// Populate the response struct and return
	txDetail := TxDetail{
		NativeTxHash: reply.Digest,
		From:         reply.Transaction.Data.Sender,
		Timestamp:    time.UnixMilli(reply.TimestampMs),
	}
	return &txDetail, nil
}
