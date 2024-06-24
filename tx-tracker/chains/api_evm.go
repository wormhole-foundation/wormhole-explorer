package chains

import (
	"context"
	"fmt"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	methodEthTxByHash  = "eth_getTransactionByHash"
	methodEthTxReceipt = "eth_getTransactionReceipt"
)

type ethGetTransactionByHashResponse struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	To          string `json:"to"`
}

type ethGetTransactionReceiptResponse struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	To               string `json:"to"`
	EfectiveGasPrice string `json:"effectiveGasPrice"`
	GasUsed          string `json:"gasUsed"`
}

type apiEvm struct {
	chainId sdk.ChainID
}

func (e *apiEvm) FetchEvmTx(
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

	var txDetail *TxDetail
	var err error
	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		txDetail, err = e.fetchEvmTx(ctx, rpc.Id, txHash, methodEthTxReceipt)
		if err != nil {
			metrics.IncCallRpcError(uint16(e.chainId), rpc.Description)
			logger.Debug("Failed to fetch transaction from evm node", zap.String("url", rpc.Id), zap.Error(err))
			continue
		}
		metrics.IncCallRpcSuccess(uint16(e.chainId), rpc.Description)
		break
	}
	return txDetail, err
}

func (e *apiEvm) fetchEvmTx(
	ctx context.Context,
	baseUrl string,
	txHash string,
	method string,
) (*TxDetail, error) {
	switch method {
	case "eth_getTransactionByHash":
		return e.fetchEvmTxByTxHash(ctx, baseUrl, txHash)
	case "eth_getTransactionReceipt":
		return e.fetchEvmTxReceiptByTxHash(ctx, baseUrl, txHash)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func (e *apiEvm) fetchEvmTxByTxHash(
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

func (e *apiEvm) fetchEvmTxReceiptByTxHash(
	ctx context.Context,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {
	client, err := rpcDialContext(ctx, baseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RPC client: %w", err)
	}
	defer client.Close()

	nativeTxHash := txHashLowerCaseWith0x(txHash)
	var txReceiptResponse ethGetTransactionReceiptResponse
	{
		err = client.CallContext(ctx, &txReceiptResponse, "eth_getTransactionReceipt", nativeTxHash)
		if err != nil {
			return nil, fmt.Errorf("failed to get tx receipt: %w", err)
		}
		if txReceiptResponse.BlockHash == "" || txReceiptResponse.From == "" {
			return nil, ErrTransactionNotFound
		}
	}

	return &TxDetail{
		From:         strings.ToLower(txReceiptResponse.From),
		NativeTxHash: nativeTxHash,
		Fee: &FeeDetail{
			Fee: 0,
		},
	}, nil
}
