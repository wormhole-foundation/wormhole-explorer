package aptos

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/support"
	"go.uber.org/ratelimit"
)

var ErrTooManyRequests = fmt.Errorf("too many requests")

// AptosSDK is a client for the Aptos API.
type AptosSDK struct {
	client *resty.Client
	rl     ratelimit.Limiter
}

type GetLatestBlock struct {
	BlockHeight string `json:"block_height"`
}

type Payload struct {
	Function      string   `json:"function"`
	TypeArguments []string `json:"type_arguments"`
	Arguments     []any    `json:"arguments"`
	Type          string   `json:"type"`
}

type Transaction struct {
	Version string  `json:"version"`
	Hash    string  `json:"hash"`
	Payload Payload `json:"payload,omitempty"`
}

type GetBlockResult struct {
	BlockHeight    string        `json:"block_height"`
	BlockHash      string        `json:"block_hash"`
	BlockTimestamp string        `json:"block_timestamp"`
	Transactions   []Transaction `json:"transactions"`
}

func (r *GetBlockResult) GetBlockTime() (*time.Time, error) {
	t, err := strconv.ParseUint(r.BlockTimestamp, 10, 64)
	if err != nil {
		return nil, err
	}
	tm := time.UnixMicro(int64(t))
	return &tm, nil
}

// NewAptosSDK creates a new AptosSDK.
func NewAptosSDK(url string, rl ratelimit.Limiter) *AptosSDK {
	return &AptosSDK{
		rl:     rl,
		client: resty.New().SetBaseURL(url),
	}
}

func (s *AptosSDK) GetLatestBlock(ctx context.Context) (uint64, error) {
	s.rl.Take()
	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&GetLatestBlock{}).
		Get("v1")

	if err != nil {
		return 0, err
	}

	if resp.IsError() {
		return 0, fmt.Errorf("status code: %s. %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*GetLatestBlock)
	if result == nil {
		return 0, fmt.Errorf("empty response")
	}
	if result.BlockHeight == "" {
		return 0, fmt.Errorf("empty block height")
	}
	return support.DecodeUint64(result.BlockHeight)
}

func (s *AptosSDK) GetBlock(ctx context.Context, block uint64) (*GetBlockResult, error) {
	s.rl.Take()
	resp, err := s.client.R().
		SetContext(ctx).
		SetResult(&GetBlockResult{}).
		SetQueryParam("with_transactions", "true").
		Get(fmt.Sprintf("v1/blocks/by_height/%d", block))

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		if resp.StatusCode() == http.StatusTooManyRequests {
			return nil, ErrTooManyRequests
		}
		return nil, fmt.Errorf("status code: %s. %s", resp.Status(), string(resp.Body()))
	}

	return resp.Result().(*GetBlockResult), nil
}
