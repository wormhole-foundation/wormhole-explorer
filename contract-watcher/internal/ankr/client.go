package ankr

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/metrics"
	"go.uber.org/ratelimit"
)

const clientName = "ankr"

type AnkrSDK struct {
	url     string
	client  *http.Client
	rl      ratelimit.Limiter
	metrics metrics.Metrics
}

func NewAnkrSDK(url string, rl ratelimit.Limiter, metrics metrics.Metrics) *AnkrSDK {
	return &AnkrSDK{
		url:     url,
		rl:      rl,
		client:  &http.Client{},
		metrics: metrics,
	}
}

func (s AnkrSDK) GetTransactionsByAddress(ctx context.Context, request TransactionsByAddressRequest) (*TransactionsByAddressResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	s.rl.Take()

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	s.metrics.IncRpcRequest(clientName, "get-transaction-by-address", res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response TransactionsByAddressResponse
	err = json.Unmarshal(body, &response)

	return &response, err

}

func (s AnkrSDK) GetBlockchainStats(ctx context.Context, blockchain string) (*BlockchainStatsResponse, error) {
	request := TransactionsByAddressRequest{
		ID:      rand.Int63(),
		Jsonrpc: "2.0",
		Method:  "ankr_getBlockchainStats",
		RequestParams: RequestParams{
			Blockchain: blockchain,
		},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", s.url, bytes.NewReader(payload))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	s.rl.Take()

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	s.metrics.IncRpcRequest(clientName, "get-blockchain-stats", res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response BlockchainStatsResponse
	err = json.Unmarshal(body, &response)

	return &response, err

}
