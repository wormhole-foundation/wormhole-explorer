package ankr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	"golang.org/x/time/rate"
)

type AnkrSDK struct {
	url    string
	client *http.Client
	rl     *rate.Limiter
}

func NewAnkrSDK(url string, rl *rate.Limiter) *AnkrSDK {
	return &AnkrSDK{
		url:    url,
		rl:     rl,
		client: &http.Client{},
	}
}

func (s AnkrSDK) GetTransactionsByAddress(ctx context.Context, request TransactionsByAddressRequest) (*TransactionsByAddressResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", s.url, bytes.NewReader(payload))

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	s.rl.Wait(ctx)

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

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

	s.rl.Wait(ctx)

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response BlockchainStatsResponse
	err = json.Unmarshal(body, &response)

	return &response, err

}
