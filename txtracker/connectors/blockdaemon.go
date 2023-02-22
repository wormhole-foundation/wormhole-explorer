package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
)

const (
	TokenBridgeBsc   = "0xb6f6d86a8f9879a9c87f643768d9efc38c1da6e7"
	TokenBridgeEth   = "0x3ee18b2214aff97000d974cf647e7c347e8fa585"
	TokenBridgeMatic = "0x5a58505a96d1dbf8df91cb21b54419fc36e93fde"
)

type TxData struct {
	Source      string
	Destination string
	Amount      *big.Int
	Decimals    uint8
	Date        time.Time
}

type blockdaemonFetchTxParams struct {
	chainName   string
	txHash      string
	eventFilter func(*EthereumEvent) bool
}

func FetchPolygonTx(
	ctx context.Context,
	cfg *config.Settings,
	txHash string,
) (*TxData, error) {

	eventFilter := func(e *EthereumEvent) bool {

		if e.Type_ != "transfer" {
			return false
		}

		if e.Meta.ContractEventName != "Transfer" && e.Meta.ContractEventName != "LogTransfer" {
			return false
		}

		if strings.ToLower(e.Destination) != TokenBridgeMatic {
			return false
		}

		return true
	}

	p := blockdaemonFetchTxParams{
		chainName:   "polygon",
		txHash:      txHash,
		eventFilter: eventFilter,
	}

	return blockdaemonFetchTx(ctx, cfg, &p)
}

func FetchEthereumTx(
	ctx context.Context,
	cfg *config.Settings,
	txHash string,
) (*TxData, error) {

	eventFilter := func(e *EthereumEvent) bool {

		if e.Type_ != "transfer" {
			return false
		}

		if strings.ToLower(e.Destination) != TokenBridgeEth {
			return false
		}

		return true
	}

	p := blockdaemonFetchTxParams{
		chainName:   "ethereum",
		txHash:      txHash,
		eventFilter: eventFilter,
	}

	return blockdaemonFetchTx(ctx, cfg, &p)
}

func blockdaemonFetchTx(
	ctx context.Context,
	cfg *config.Settings,
	params *blockdaemonFetchTxParams,
) (*TxData, error) {

	// build the HTTP request
	url := fmt.Sprintf(
		"%s/universal/v1/%s/mainnet/tx/%s",
		cfg.BlockdaemonBaseUrl,
		params.chainName,
		params.txHash,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+cfg.BlockdaemonApiKey)

	// send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected HTTP status code: %d", resp.StatusCode)
	}

	// parse the response
	body, err := io.ReadAll(resp.Body)
	var ethereumResponse ethereumResponse
	err = json.Unmarshal(body, &ethereumResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ethereum response from blockdaemon API: %w", err)
	}

	// extract relevant fields
	var txData TxData
	var found bool
	for i := range ethereumResponse.Events {

		e := &ethereumResponse.Events[i]
		if !params.eventFilter(e) {
			continue
		}

		if found {
			return nil, fmt.Errorf("encountered two transfer events for chain=%s txHash=%s", params.chainName, params.txHash)
		}

		found = true
		txData = TxData{
			Source:      e.Source,
			Destination: e.Destination,
			Amount:      e.Amount,
			Decimals:    e.Decimals,
			Date:        time.Unix(e.Date, 0),
		}
	}
	if !found {
		return nil, fmt.Errorf("no matching events for chain=%s txHash=%s", params.chainName, params.txHash)
	}

	return &txData, nil
}

type ethereumResponse struct {
	Events []EthereumEvent `json:"events"`
}

type Meta struct {
	ContractEventName string `json:"contract_event_name"`
}

type EthereumEvent struct {
	Type_       string   `json:"type"`
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Date        int64    `json:"date"`
	Meta        Meta     `json:"meta"`
	Amount      *big.Int `json:"amount"`
	Decimals    uint8    `json:"decimals"`
}
