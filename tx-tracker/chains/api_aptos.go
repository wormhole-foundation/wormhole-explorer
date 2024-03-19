package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
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

func FetchAptosTx(
	ctx context.Context,
	pool *pool.Pool,
	txHash string,
	metrics metrics.Metrics,
	logger *zap.Logger,
) (*TxDetail, error) {

	// Parse the Aptos event creation number
	creationNumber, err := strconv.ParseUint(txHash, 16, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event creation number from Aptos tx hash: %w", err)
	}

	// get rpc sorted by score and priority.
	rpcs := pool.GetItems()
	if len(rpcs) == 0 {
		return nil, ErrChainNotSupported
	}

	// Get the event from the Aptos node API.
	var events []aptosEvent
	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		events, err = fetchAptosAccountEvents(ctx, rpc.Id, aptosCoreContractAddress, creationNumber, 1)
		if err != nil {
			metrics.IncCallRpcError(uint16(sdk.ChainIDAptos), rpc.Description)
			logger.Debug("Failed to fetch transaction from Aptos node", zap.String("url", rpc.Id), zap.Error(err))
			continue
		}
		metrics.IncCallRpcSuccess(uint16(sdk.ChainIDAptos), rpc.Description)
		break
	}

	// Return an error if the event is not found
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, ErrTransactionNotFound
	} else if len(events) > 1 {
		return nil, fmt.Errorf("expected exactly one event, but got %d", len(events))
	}

	// get rpc sorted by score and priority.
	rpcs = pool.GetItems()
	if len(rpcs) == 0 {
		return nil, ErrChainNotSupported
	}

	// Get the transaction from the Aptos node API.
	var tx *aptosTx
	for _, rpc := range rpcs {
		// Wait for the RPC rate limiter
		rpc.Wait(ctx)
		tx, err = fetchAptosTx(ctx, rpc.Id, events[0].Version)
		if err != nil {
			metrics.IncCallRpcError(uint16(sdk.ChainIDAptos), rpc.Description)
			logger.Debug("Failed to fetch transaction from Aptos node", zap.String("url", rpc.Id), zap.Error(err))
			continue
		}
		metrics.IncCallRpcSuccess(uint16(sdk.ChainIDAptos), rpc.Description)
		break
	}

	// Return an error if the transaction is not found
	if tx == nil {
		return nil, ErrTransactionNotFound
	}

	// Build the result struct and return
	TxDetail := TxDetail{
		NativeTxHash: tx.Hash,
		From:         tx.Sender,
	}
	return &TxDetail, nil
}

// fetchAptosAccountEvents queries the Aptos node API for the events of a given account.
func fetchAptosAccountEvents(ctx context.Context, baseUrl string, contractAddress string, start uint64, limit uint64) ([]aptosEvent, error) {
	// Build the URI for the events endpoint
	uri := fmt.Sprintf("%s/v1/accounts/%s/events/%s::state::WormholeMessageHandle/event?start=%d&limit=%d",
		baseUrl,
		contractAddress,
		contractAddress,
		start,
		limit,
	)

	// Query the events endpoint
	body, err := httpGet(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to query events endpoint: %w", err)
	}

	// Deserialize the response
	var events []aptosEvent
	err = json.Unmarshal(body, &events)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body from events endpoint: %w", err)
	}

	return events, nil
}

// fetchAptosTx queries the Aptos node API for the transaction details of a given version.
func fetchAptosTx(ctx context.Context, baseUrl string, version uint64) (*aptosTx, error) {
	// Build the URI for the events endpoint
	uri := fmt.Sprintf("%s/v1/transactions/by_version/%d", baseUrl, version)

	// Query the events endpoint
	body, err := httpGet(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions endpoint: %w", err)
	}

	// Deserialize the response
	var tx aptosTx
	err = json.Unmarshal(body, &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body from transactions endpoint: %w", err)
	}

	return &tx, nil
}
