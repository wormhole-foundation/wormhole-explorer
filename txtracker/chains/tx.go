package chains

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

const requestTimeout = 30 * time.Second

var (
	ErrChainNotSupported = errors.New("chain id not supported")
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
	ankr        *time.Ticker
	blockdaemon *time.Ticker
	solana      *time.Ticker
	terra       *time.Ticker
}{}

func init() {

	tickers.ankr = time.NewTicker(2 * time.Second)
	tickers.blockdaemon = time.NewTicker(5 * time.Second)
	tickers.terra = time.NewTicker(5 * time.Second)

	// the Solana adapter sends 2 requests per txHash
	tickers.solana = time.NewTicker(10 * time.Second)
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
