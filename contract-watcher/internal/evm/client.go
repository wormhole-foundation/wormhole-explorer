package evm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/metrics"
	"go.uber.org/ratelimit"
)

var ErrTooManyRequests = fmt.Errorf("too many requests")

const clientName = "evm"

type EvmSDK struct {
	client  *resty.Client
	rl      ratelimit.Limiter
	metrics metrics.Metrics
}

func NewEvmSDK(url string, rl ratelimit.Limiter, metrics metrics.Metrics) *EvmSDK {
	return &EvmSDK{
		rl:      rl,
		client:  resty.New().SetBaseURL(url),
		metrics: metrics,
	}
}

func (s *EvmSDK) GetLatestBlock(ctx context.Context) (uint64, error) {
	s.rl.Take()
	req := newEvmRequest("eth_blockNumber")
	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&getLatestBlockResponse{}).
		Post("")

	if err != nil {
		return 0, err
	}

	s.metrics.IncRpcRequest(clientName, "get-latest-block", resp.StatusCode())

	if resp.IsError() {
		if resp.StatusCode() == http.StatusTooManyRequests {
			return 0, ErrTooManyRequests
		}
		return 0, fmt.Errorf("status code: %s. %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*getLatestBlockResponse)
	if result == nil {
		return 0, fmt.Errorf("empty response")
	}
	return utils.DecodeUint64(result.Result)
}

func (s *EvmSDK) GetBlock(ctx context.Context, block uint64) (*GetBlockResult, error) {
	s.rl.Take()
	req := newEvmRequest("eth_getBlockByNumber", utils.EncodeHex(block), true)
	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&getBlockResponse{}).
		Post("")

	if err != nil {
		return nil, err
	}

	s.metrics.IncRpcRequest(clientName, "get-block", resp.StatusCode())

	if resp.IsError() {
		if resp.StatusCode() == http.StatusTooManyRequests {
			return nil, ErrTooManyRequests
		}
		return nil, fmt.Errorf("status code: %s. %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*getBlockResponse)
	if result == nil {
		return nil, fmt.Errorf("empty response")
	}
	return &result.Result, nil
}

func (s *EvmSDK) GetTransactionReceipt(ctx context.Context, txHash string) (*TransactionReceiptResult, error) {
	s.rl.Take()
	req := newEvmRequest("eth_getTransactionReceipt", txHash)
	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&getTransactionReceiptResponse{}).
		Post("")

	if err != nil {
		return nil, err
	}

	s.metrics.IncRpcRequest(clientName, "get-transaction-receipt", resp.StatusCode())

	if resp.IsError() {
		if resp.StatusCode() == http.StatusTooManyRequests {
			return nil, ErrTooManyRequests
		}
		return nil, fmt.Errorf("status code: %s. %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*getTransactionReceiptResponse)
	if result == nil {
		return nil, fmt.Errorf("empty response")
	}
	return &result.Result, nil
}

func newEvmRequest(method string, params ...any) EvmRequest {
	return EvmRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
}
