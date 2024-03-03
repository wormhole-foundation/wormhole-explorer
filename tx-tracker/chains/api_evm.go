package chains

import (
	"context"
	"fmt"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
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

func FetchEvmTx(
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
		txDetail, err = fetchEvmTx(ctx, rpc.Id, txHash)
		if err != nil {
			logger.Debug("Failed to fetch transaction from evm node", zap.String("url", rpc.Id), zap.Error(err))
			continue
		}
		break
	}
	return txDetail, err
}

func fetchEvmTx(
	ctx context.Context,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	// initialize RPC client
	client, err := rpcDialContext(ctx, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	nativeTxHash := txHashLowerCaseWith0x(txHash)
	// query transaction data
	var txReply ethGetTransactionByHashResponse
	{
		err = client.CallContext(ctx, &txReply, "eth_getTransactionByHash", nativeTxHash)
		if err != nil {
			return nil, fmt.Errorf("failed to get tx by hash: %w", err)
		}
		if txReply.BlockHash == "" || txReply.From == "" {
			return nil, ErrTransactionNotFound
		}
	}

	// build results and return
	txDetail := &TxDetail{
		From:         strings.ToLower(txReply.From),
		NativeTxHash: nativeTxHash,
	}
	return txDetail, nil
}
