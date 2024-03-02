package chains

import (
	"context"
	"fmt"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
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

func fetchEvmTx(
	ctx context.Context,
	chainID sdk.ChainID,
	rpcPool map[sdk.ChainID]*pool.Pool,
	txHash string,
	logger *zap.Logger,
) (*TxDetail, error) {

	rpcs, err := getRpcPool(rpcPool, chainID)
	if err != nil {
		return nil, err
	}

	var txReply *ethGetTransactionByHashResponse
	nativeTxHash := txHashLowerCaseWith0x(txHash)

	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		// initialize RPC client
		client, err := rpcDialContext(ctx, rpc.Id)
		if err != nil {
			logger.Error("failed to initialize RPC client", zap.Error(err), zap.String("rpc", rpc.Id))
			continue
		}
		defer client.Close()

		err = client.CallContext(ctx, &txReply, "eth_getTransactionByHash", nativeTxHash)
		if err == nil {
			if txReply == nil {
				continue
			}
			if txReply.BlockHash == "" || txReply.From == "" {
				continue
			}
			fmt.Printf("rpc.Id = %s \n", rpc.Id)
			break
		} else {
			logger.Error("failed to get tx by hash", zap.Error(err), zap.String("rpc", rpc.Id))
			continue
		}

	}

	if txReply == nil {
		return nil, ErrTransactionNotFound
	}
	if txReply.BlockHash == "" || txReply.From == "" {
		return nil, ErrTransactionNotFound
	}

	// build results and return
	txDetail := &TxDetail{
		From:         strings.ToLower(txReply.From),
		NativeTxHash: nativeTxHash,
	}
	return txDetail, nil
}
