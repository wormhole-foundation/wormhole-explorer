package chains

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
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
	aptos     *time.Ticker
	arbitrum  *time.Ticker
	avalanche *time.Ticker
	bsc       *time.Ticker
	celo      *time.Ticker
	ethereum  *time.Ticker
	fantom    *time.Ticker
	moonbeam  *time.Ticker
	optimism  *time.Ticker
	polygon   *time.Ticker
	solana    *time.Ticker
	sui       *time.Ticker
}{}

func Initialize(cfg *config.RpcProviderSettings) {

	// f converts "requests per minute" into the associated time.Duration
	f := func(requestsPerMinute uint16) time.Duration {

		division := float64(time.Minute) / float64(time.Duration(requestsPerMinute))
		roundedUp := math.Ceil(division)

		return time.Duration(roundedUp)
	}

	// this adapter sends 1 request per txHash
	tickers.sui = time.NewTicker(f(cfg.SuiRequestsPerMinute))

	// these adapters send 2 requests per txHash
	tickers.aptos = time.NewTicker(f(cfg.AptosRequestsPerMinute / 2))
	tickers.arbitrum = time.NewTicker(f(cfg.ArbitrumRequestsPerMinute / 2))
	tickers.avalanche = time.NewTicker(f(cfg.AvalancheRequestsPerMinute / 2))
	tickers.bsc = time.NewTicker(f(cfg.BscRequestsPerMinute / 2))
	tickers.celo = time.NewTicker(f(cfg.CeloRequestsPerMinute / 2))
	tickers.ethereum = time.NewTicker(f(cfg.EthereumRequestsPerMinute / 2))
	tickers.fantom = time.NewTicker(f(cfg.FantomRequestsPerMinute / 2))
	tickers.moonbeam = time.NewTicker(f(cfg.MoonbeamRequestsPerMinute / 2))
	tickers.optimism = time.NewTicker(f(cfg.OptimismRequestsPerMinute / 2))
	tickers.polygon = time.NewTicker(f(cfg.PolygonRequestsPerMinute / 2))
	tickers.solana = time.NewTicker(f(cfg.SolanaRequestsPerMinute / 2))
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
	case vaa.ChainIDCelo:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.CeloBaseUrl)
		}
		rateLimiter = *tickers.celo
	case vaa.ChainIDEthereum:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.EthereumBaseUrl)
		}
		rateLimiter = *tickers.ethereum
	case vaa.ChainIDBSC:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.BscBaseUrl)
		}
		rateLimiter = *tickers.bsc
	case vaa.ChainIDPolygon:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.PolygonBaseUrl)
		}
		rateLimiter = *tickers.polygon
	case vaa.ChainIDFantom:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.FantomBaseUrl)
		}
		rateLimiter = *tickers.fantom
	case vaa.ChainIDArbitrum:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.ArbitrumBaseUrl)
		}
		rateLimiter = *tickers.arbitrum
	case vaa.ChainIDOptimism:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.OptimismBaseUrl)
		}
		rateLimiter = *tickers.optimism
	case vaa.ChainIDAvalanche:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.AvalancheBaseUrl)
		}
		rateLimiter = *tickers.avalanche
	case vaa.ChainIDMoonbeam:
		fetchFunc = func(ctx context.Context, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, txHash, cfg.MoonbeamBaseUrl)
		}
		rateLimiter = *tickers.avalanche
	case vaa.ChainIDAptos:
		fetchFunc = fetchAptosTx
		rateLimiter = *tickers.aptos
	case vaa.ChainIDSui:
		fetchFunc = fetchSuiTx
		rateLimiter = *tickers.sui
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

// httpGet is a helper function that performs an HTTP request.
func httpGet(ctx context.Context, url string) ([]byte, error) {

	// Build the HTTP request
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send it
	var client http.Client
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to query url: %w", err)
	}
	defer response.Body.Close()

	// Read the response body and return
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}
