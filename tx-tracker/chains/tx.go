package chains

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

const requestTimeout = 30 * time.Second

var (
	ErrChainNotSupported   = errors.New("chain id not supported")
	ErrTransactionNotFound = errors.New("transaction not found")
)

type TxDetail struct {
	// Signer is the address that signed the transaction, encoded in the chain's native format.
	Signer string
	// Timestamp indicates the time at which the transaction was confirmed.
	Timestamp time.Time
	// NativeTxHash contains the transaction hash, encoded in the chain's native format.
	NativeTxHash string
}

var tickers = struct {
	ankr   *time.Ticker
	solana *time.Ticker
	terra  *time.Ticker
}{}

func Initialize(cfg *config.Settings) {

	// f converts "requests per minute" into the associated time.Duration
	f := func(requestsPerMinute uint16) time.Duration {

		division := float64(time.Minute) / float64(time.Duration(requestsPerMinute))
		roundedUp := math.Ceil(division)

		return time.Duration(roundedUp)
	}

	tickers.ankr = time.NewTicker(f(cfg.AnkrRequestsPerMinute))
	tickers.terra = time.NewTicker(f(cfg.TerraRequestsPerMinute))

	// the Solana adapter sends 2 requests per txHash
	tickers.solana = time.NewTicker(f(cfg.SolanaRequestsPerMinute / 2))
}

func FetchTx(
	ctx context.Context,
	cfg *config.Settings,
	chainId vaa.ChainID,
	txHash string,
) (*TxDetail, error) {

	var fetchFunc func(context.Context, *config.Settings, string) (*TxDetail, error)
	var rateLimiter time.Ticker

	// decide which RPC/API service to use based on chain ID
	switch chainId {
	case vaa.ChainIDSolana:
		fetchFunc = fetchSolanaTx
		rateLimiter = *tickers.solana
	case vaa.ChainIDTerra:
		fetchFunc = fetchTerraTx
		rateLimiter = *tickers.terra
	// most EVM-compatible chains use the same RPC service
	case vaa.ChainIDEthereum,
		vaa.ChainIDBSC,
		vaa.ChainIDPolygon,
		vaa.ChainIDAvalanche,
		vaa.ChainIDFantom,
		vaa.ChainIDArbitrum,
		vaa.ChainIDOptimism:

		fetchFunc = ankrFetchTx
		rateLimiter = *tickers.ankr
	default:
		return nil, ErrChainNotSupported
	}

	// wait for rate limit - fail fast if context was cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-rateLimiter.C:
	}

	// get transaction details from the RPC/API service
	subContext, cancelFunc := context.WithTimeout(ctx, requestTimeout)
	defer cancelFunc()
	txDetail, err := fetchFunc(subContext, cfg, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tx information: %w", err)
	}

	return txDetail, nil
}
