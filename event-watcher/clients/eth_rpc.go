package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type logsResponse struct {
	Result []Log `json:"result"`
}

type Log struct {
	Address         string   `json:"address"`
	BlockHash       string   `json:"blockHash"`
	BlockNumber     string   `json:"blockNumber"`
	Data            string   `json:"data"`
	Topics          []string `json:"topics"`
	TransactionHash string   `json:"transactionHash"`
}

type EthRpcClient struct {
	Url  string
	Auth string
}

// TODO add rate limits
func NewEthRpcClient(url string, auth string) *EthRpcClient {
	return &EthRpcClient{Url: url, Auth: auth}
}

func (c *EthRpcClient) GetBlockNumber(ctx context.Context) (uint64, error) {

	// Create a new HTTP request
	payload := strings.NewReader(`{
		"id": 1,
		"jsonrpc": "2.0",
		"method": "eth_blockNumber"
	}`)
	req, err := http.NewRequestWithContext(ctx, "POST", c.Url, payload)
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer: "+c.Auth)

	// Send the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer res.Body.Close()
	if res.Status != "200 OK" {
		return 0, fmt.Errorf("encoutered unexpected HTTP status code in response: %s", res.Status)
	}

	// Deserialize response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read HTTP response body: %w", err)
	}
	var response struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to deserialize HTTP response body: %w", err)
	}

	// Parse the block number
	n, err := hexutil.DecodeUint64(response.Result)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block number from hex: %w", err)
	}

	return n, nil
}

func (c *EthRpcClient) GetLogs(
	ctx context.Context,
	fromBlock uint64,
	toBlock uint64,
	address string,
	topic string,
) ([]Log, error) {

	params := fmt.Sprintf(`{
		"id": 1,
		"jsonrpc": "2.0",
		"method": "eth_getLogs",
		"params": [{
			"address": ["%s"],
			"fromBlock":"0x%x",
			"toBlock":"0x%x",
			"topics": ["%s"]
		}]
	}`, address, fromBlock, toBlock, topic)
	payload := strings.NewReader(params)

	req, err := http.NewRequest("POST", c.Url, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer: "+c.Auth)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer res.Body.Close()

	// Deserialize response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body: %w", err)
	}
	var response logsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to deserialize HTTP response body: %w", err)
	}

	return response.Result, nil
}
