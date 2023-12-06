package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const (
	aptosCoreContractAddress = "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625"
)

type aptosEvent struct {
	Version uint64 `json:"version,string"`
}

type aptosTx struct {
	Timestamp uint64 `json:"timestamp,string"`
	Sender    string `json:"sender"`
	Hash      string `json:"hash"`
}

func fetchAptosTx(
	ctx context.Context,
	rateLimiter *time.Ticker,
	baseUrl string,
	txHash string,
) (*TxDetail, error) {

	// Parse the Aptos event creation number
	creationNumber, err := strconv.ParseUint(txHash, 16, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event creation number from Aptos tx hash: %w", err)
	}

	// Get the event from the Aptos node API.
	var events []aptosEvent
	{
		// Build the URI for the events endpoint
		uri := fmt.Sprintf("%s/v1/accounts/%s/events/%s::state::WormholeMessageHandle/event?start=%d&limit=1",
			baseUrl,
			aptosCoreContractAddress,
			aptosCoreContractAddress,
			creationNumber,
		)

		// Query the events endpoint
		body, err := httpGet(ctx, rateLimiter, uri)
		if err != nil {
			return nil, fmt.Errorf("failed to query events endpoint: %w", err)
		}

		// Deserialize the response
		err = json.Unmarshal(body, &events)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response body from events endpoint: %w", err)
		}
	}
	if len(events) == 0 {
		return nil, ErrTransactionNotFound
	} else if len(events) > 1 {
		return nil, fmt.Errorf("expected exactly one event, but got %d", len(events))
	}

	// Get the transaction
	var tx aptosTx
	{
		// Build the URI for the events endpoint
		uri := fmt.Sprintf("%s/v1/transactions/by_version/%d", baseUrl, events[0].Version)

		// Query the events endpoint
		body, err := httpGet(ctx, rateLimiter, uri)
		if err != nil {
			return nil, fmt.Errorf("failed to query transactions endpoint: %w", err)
		}

		// Deserialize the response
		err = json.Unmarshal(body, &tx)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response body from transactions endpoint: %w", err)
		}
	}

	// Build the result struct and return
	TxDetail := TxDetail{
		NativeTxHash: tx.Hash,
		From:         tx.Sender,
	}
	return &TxDetail, nil
}
