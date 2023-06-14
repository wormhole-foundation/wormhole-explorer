package chains

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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
	// From is the address that signed the transaction, encoded in the chain's native format.
	From string
	// Timestamp indicates the time at which the transaction was confirmed.
	Timestamp time.Time
	// NativeTxHash contains the transaction hash, encoded in the chain's native format.
	NativeTxHash string
}

var tickers = struct {
	solana   *time.Ticker
	ethereum *time.Ticker
}{}

func Initialize(cfg *config.RpcProviderSettings) {

	// f converts "requests per minute" into the associated time.Duration
	f := func(requestsPerMinute uint16) time.Duration {

		division := float64(time.Minute) / float64(time.Duration(requestsPerMinute))
		roundedUp := math.Ceil(division)

		return time.Duration(roundedUp)
	}

	// these adapters send 2 requests per txHash
	tickers.solana = time.NewTicker(f(cfg.SolanaRequestsPerMinute / 2))
	tickers.ethereum = time.NewTicker(f(cfg.EthereumRequestsPerMinute / 2))
}

func FetchTx(
	ctx context.Context,
	cfg *config.RpcProviderSettings,
	chainId vaa.ChainID,
	txHash string,
) (*TxDetail, error) {

	var fetchFunc func(context.Context, *config.RpcProviderSettings, string) (*TxDetail, error)
	var rateLimiter time.Ticker

	// decide which RPC/API service to use based on chain ID
	switch chainId {
	case vaa.ChainIDSolana:
		fetchFunc = fetchSolanaTx
		rateLimiter = *tickers.solana
	case vaa.ChainIDEthereum:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.EthereumBaseUrl)
		}
		rateLimiter = *tickers.ethereum
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

// timestampFromHex converts a hex timestamp into a `time.Time` value.
func timestampFromHex(s string) (time.Time, error) {

	// remove the leading "0x" or "0X" from the hex string
	hexDigits := strings.Replace(s, "0x", "", 1)
	hexDigits = strings.Replace(hexDigits, "0X", "", 1)

	// parse the hex digits into an integer
	epoch, err := strconv.ParseInt(hexDigits, 16, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse hex timestamp: %w", err)
	}

	// convert the unix epoch into a `time.Time` value
	timestamp := time.Unix(epoch, 0).UTC()
	return timestamp, nil
}
