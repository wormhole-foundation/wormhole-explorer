package evm

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/support"
	"go.uber.org/ratelimit"
)

var ErrTooManyRequests = fmt.Errorf("too many requests")

type EvmSDK struct {
	client *resty.Client
	rl     ratelimit.Limiter
}

func NewEvmSDK(url string, rl ratelimit.Limiter) *EvmSDK {
	return &EvmSDK{
		rl:     rl,
		client: resty.New().SetBaseURL(url),
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
	return support.DecodeUint64(result.Result)
}

func (s *EvmSDK) GetBlock(ctx context.Context, block uint64) (*GetBlockResult, error) {
	s.rl.Take()
	req := newEvmRequest("eth_getBlockByNumber", support.EncodeHex(block), true)
	resp, err := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&getBlockResponse{}).
		Post("")

	if err != nil {
		return nil, err
	}

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

func newEvmRequest(method string, params ...any) EvmRequest {
	return EvmRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}
}
