package ankr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
)

type AnkrSDK struct {
	url    string
	client *http.Client
}

func NewAnkrSDK(url string) *AnkrSDK {
	return &AnkrSDK{
		url:    url,
		client: &http.Client{},
	}
}

func (s AnkrSDK) TransactionByAddressRequest(blockChain, contractAddress string, fromBlock int64, toBlock int64) TransactionsByAddressRequest {

	request := TransactionsByAddressRequest{
		ID:      rand.Int63(),
		Jsonrpc: "2.0",
		Method:  "ankr_getTransactionsByAddress",
		RequestParams: RequestParams{
			Blockchain: blockChain,
			Address:    contractAddress,
			FromBlock:  fromBlock,
			ToBlock:    toBlock,
			DescOrder:  false,
		},
	}

	return request
}

func (s AnkrSDK) GetTransactionsByAddress(request TransactionsByAddressRequest) (*TransactionsByAddressResponse, error) {
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

func (s AnkrSDK) GetBlockchainStats(blockchain string) (*BlockchainStatsResponse, error) {
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
