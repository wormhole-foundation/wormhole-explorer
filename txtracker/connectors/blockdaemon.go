package connectors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
)

type TxData struct {
	Source      string
	Destination string
	Amount      uint64
	Date        time.Time
}

func FetchEthereumTx(cfg *config.Settings, txHash string) (*TxData, error) {
	return blockdaemonFetchTx(cfg, "ethereum", txHash)
}

func blockdaemonFetchTx(cfg *config.Settings, chain string, txHash string) (*TxData, error) {

	// build the HTTP request
	url := fmt.Sprintf(
		"%s/universal/v1/%s/mainnet/tx/%s",
		cfg.BlockdaemonBaseUrl,
		chain,
		txHash,
	)
	req, err := http.NewRequest("GET", url, nil)
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
	var ethereumResponse EthereumResponse
	err = json.Unmarshal(body, &ethereumResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal ethereum response from blockdaemon API: %w", err)
	}

	// extract relevant fields
	var txData TxData
	for i := range ethereumResponse.Events {

		e := &ethereumResponse.Events[i]
		if e.Type_ != "transfer" {
			continue
		}

		if txData != (TxData{}) {
			return nil, fmt.Errorf("encountered two transfer events for chain=%s txHash=%s", chain, txHash)
		}

		txData = TxData{
			Source:      e.Source,
			Destination: e.Destination,
			Amount:      e.Amount,
			Date:        time.Unix(e.Date, 0),
		}
	}
	if txData == (TxData{}) {
		return nil, fmt.Errorf("expected at least one 'transfer' event for chain=%s txHash=%s", chain, txHash)
	}

	return &txData, nil
}

type EthereumResponse struct {
	Events []Event `json:"events"`
}

type Event struct {
	Type_       string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Date        int64  `json:"date"`
	Amount      uint64 `json:"amount"`
}
