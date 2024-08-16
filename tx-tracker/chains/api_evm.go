package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
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
	chainId       sdk.ChainID
	notionalCache *notional.NotionalCache
	p2pNetwork    string
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

	// calculate tx fee
	if txDetail != nil && txDetail.FeeDetail != nil {
		fee, err := EvmCalculateFee(e.chainId, txDetail.FeeDetail.RawFee["gasUsed"],
			txDetail.FeeDetail.RawFee["effectiveGasPrice"])
		if err != nil {
			logger.Debug("can not calculated fee",
				zap.Error(err),
				zap.String("txHash", txHash),
				zap.String("chainId", e.chainId.String()))
		} else if fee == nil {
			txDetail.FeeDetail = nil
		} else {
			txDetail.FeeDetail.Fee = fee.String()
			if e.p2pNetwork == domain.P2pMainNet {
				gasPrice, errGasPrice := GetGasTokenNotional(e.chainId, e.notionalCache)
				if errGasPrice != nil {
					logger.Error("Failed to get gas price",
						zap.Error(errGasPrice),
						zap.String("chainId", e.chainId.String()),
						zap.String("txHash", txHash))
				} else {
					txDetail.FeeDetail.GasTokenNotional = gasPrice.NotionalUsd.String()
					txDetail.FeeDetail.FeeUSD = gasPrice.NotionalUsd.Mul(*fee).String()
				}
			}
		}
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

	respStr, _ := json.Marshal(txReply)

	// build results and return
	txDetail := &TxDetail{
		From:             strings.ToLower(txReply.From),
		To:               strings.ToLower(txReply.To),
		NativeTxHash:     nativeTxHash,
		NormalizedTxHash: utils.NormalizeHex(txHash),
		BlockNumber:      txReply.BlockNumber,
		RpcResponse:      string(respStr),
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

	var feeDetail *FeeDetail
	if txReceiptResponse.EfectiveGasPrice != "" && txReceiptResponse.GasUsed != "" {
		feeDetail = &FeeDetail{
			RawFee: map[string]string{
				"gasUsed":           txReceiptResponse.GasUsed,
				"effectiveGasPrice": txReceiptResponse.EfectiveGasPrice,
			},
		}
	}

	respStr, _ := json.Marshal(txReceiptResponse)

	return &TxDetail{
		From:             strings.ToLower(txReceiptResponse.From),
		To:               strings.ToLower(txReceiptResponse.To),
		NativeTxHash:     nativeTxHash,
		NormalizedTxHash: utils.NormalizeHex(txHash),
		BlockNumber:      txReceiptResponse.BlockNumber,
		RpcResponse:      string(respStr),
		FeeDetail:        feeDetail,
	}, nil
}

func EvmCalculateFee(chainID sdk.ChainID, gasUsed string, effectiveGasPrice string) (*decimal.Decimal, error) {
	//ignore if the blockchain is L2
	if chainID == sdk.ChainIDBase || chainID == sdk.ChainIDOptimism || chainID == sdk.ChainIDScroll {
		return nil, nil
	}

	// get decimal gasUsed
	gs := new(big.Int)
	_, ok := gs.SetString(gasUsed, 0)
	if !ok {
		return nil, fmt.Errorf("failed to convert gasUsed to big.Int")
	}
	decimalGasUsed := decimal.NewFromBigInt(gs, 0)

	// get decimal gasPrice
	gp := new(big.Int)
	_, ok = gp.SetString(effectiveGasPrice, 0)
	if !ok {
		return nil, fmt.Errorf("failed to convert gasPrice to big.Int")
	}
	decimalGasPrice := decimal.NewFromBigInt(gp, 0)

	// calculate gasUsed * (gasPrice / 1e18)
	decimalFee := decimalGasUsed.Mul(decimalGasPrice)
	decimalFee = decimalFee.DivRound(decimal.NewFromInt(1e18), 18)
	return &decimalFee, nil
}
