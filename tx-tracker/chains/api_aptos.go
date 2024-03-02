package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
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
	chainID sdk.ChainID,
	rpcPool map[sdk.ChainID]*pool.Pool,
	txHash string,
	logger *zap.Logger,
) (*TxDetail, error) {

	// Parse the Aptos event creation number
	creationNumber, err := strconv.ParseUint(txHash, 16, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event creation number from Aptos tx hash: %w", err)
	}

	// Get rpc pool
	rpcs, err := getRpcPool(rpcPool, chainID)
	if err != nil {
		return nil, err
	}

	// Get the event from the Aptos node API.
	var events []aptosEvent

	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		// Make the HTTP request
		uri := fmt.Sprintf("%s/v1/accounts/%s/events/%s::state::WormholeMessageHandle/event?start=%d&limit=1",
			rpc.Id,
			aptosCoreContractAddress,
			aptosCoreContractAddress,
			creationNumber,
		)
		body, err := httpGet(ctx, uri)
		if err != nil {
			logger.Error("HTTP request to Aptos events endpoint failed", zap.Error(err), zap.String("url", uri))
			continue
		}

		// Deserialize the response
		err = json.Unmarshal(body, &events)
		if err == nil {
			// If the response is not nil, break the loop
			fmt.Printf("rpc.Id = %s \n", rpc.Id)
			break
		} else {
			logger.Error("Failed to decode Aptos events response as JSON", zap.Error(err), zap.String("url", uri))
			continue
		}

	}

	if len(events) == 0 {
		return nil, ErrTransactionNotFound
	} else if len(events) > 1 {
		return nil, fmt.Errorf("expected exactly one event, but got %d", len(events))
	}

	// Get rpc pool
	rpcs, err = getRpcPool(rpcPool, sdk.ChainIDAptos)
	if err != nil {
		return nil, err
	}

	var tx *aptosTx
	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		// Build the URI for the events endpoint
		uri := fmt.Sprintf("%s/v1/transactions/by_version/%d", rpc.Id, events[0].Version)
		// Query the events endpoint
		body, err := httpGet(ctx, uri)
		if err != nil {
			logger.Error("HTTP request to Aptos transactions endpoint failed", zap.Error(err), zap.String("url", uri))
			continue
		}

		// Deserialize the response
		err = json.Unmarshal(body, &tx)
		if err == nil {
			fmt.Printf("rpc.Id = %s \n", rpc.Id)
			// If the response is not nil, break the loop
			break
		} else {
			logger.Error("Failed to decode Aptos transactions response as JSON", zap.Error(err), zap.String("url", uri))
			continue
		}
	}

	if tx == nil {
		return nil, fmt.Errorf("failed to fetch transaction from Aptos indexer")
	}

	// Populate the result struct and return
	TxDetail := TxDetail{
		NativeTxHash: tx.Hash,
		From:         tx.Sender,
	}
	return &TxDetail, nil
}
