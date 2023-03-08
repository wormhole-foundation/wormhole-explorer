package ankr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type AnkrSDK struct {
	url    string
	Client *http.Client
}

func NewAnkrSDK(url string) *AnkrSDK {
	return &AnkrSDK{
		url:    url,
		Client: &http.Client{},
	}
}

func (s AnkrSDK) TransactionByAddressRequest(blockChain, contractAddress string, fromBlock int64, toBlock int64) TransactionsByAddressRequest {

	request := TransactionsByAddressRequest{
		ID:      1,
		Jsonrpc: "2.0",
		Method:  "ankr_getTransactionsByAddress",
		RquestParams: RquestParams{
			Blockchain: blockChain,
			Address:    contractAddress,
			FromBlock:  fromBlock,
			ToBlock:    toBlock,
			DescOrder:  true,
		},
	}

	return request
}

func (s AnkrSDK) GetTransactionsByAddress(request TransactionsByAddressRequest) (*TransactionsByAddressResponse, error) {

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", s.url, bytes.NewReader(payload))

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
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
		ID:      1,
		Jsonrpc: "2.0",
		Method:  "ankr_getBlockchainStats",
		RquestParams: RquestParams{
			Blockchain: blockchain,
		},
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", s.url, bytes.NewReader(payload))

	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
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
