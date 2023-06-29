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
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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

var tickers map[sdk.ChainID]*time.Ticker

func Initialize(cfg *config.RpcProviderSettings) {

	// f converts "requests per minute" into the associated *time.Ticker
	f := func(requestsPerMinute uint16) *time.Ticker {

		division := float64(time.Minute) / float64(time.Duration(requestsPerMinute))
		roundedUp := math.Ceil(division)

		duration := time.Duration(roundedUp)

		return time.NewTicker(duration)
	}

	// initialize tickers for each chain
	tickers = make(map[sdk.ChainID]*time.Ticker)
	tickers[sdk.ChainIDArbitrum] = f(cfg.ArbitrumRequestsPerMinute)
	tickers[sdk.ChainIDAlgorand] = f(cfg.AlgorandRequestsPerMinute)
	tickers[sdk.ChainIDAptos] = f(cfg.AptosRequestsPerMinute)
	tickers[sdk.ChainIDAvalanche] = f(cfg.AvalancheRequestsPerMinute)
	tickers[sdk.ChainIDBSC] = f(cfg.BscRequestsPerMinute)
	tickers[sdk.ChainIDCelo] = f(cfg.CeloRequestsPerMinute)
	tickers[sdk.ChainIDEthereum] = f(cfg.EthereumRequestsPerMinute)
	tickers[sdk.ChainIDFantom] = f(cfg.FantomRequestsPerMinute)
	tickers[sdk.ChainIDKlaytn] = f(cfg.KlaytnRequestsPerMinute)
	tickers[sdk.ChainIDMoonbeam] = f(cfg.MoonbeamRequestsPerMinute)
	tickers[sdk.ChainIDOasis] = f(cfg.OasisRequestsPerMinute)
	tickers[sdk.ChainIDOptimism] = f(cfg.OptimismRequestsPerMinute)
	tickers[sdk.ChainIDPolygon] = f(cfg.PolygonRequestsPerMinute)
	tickers[sdk.ChainIDSolana] = f(cfg.SolanaRequestsPerMinute)
	tickers[sdk.ChainIDTerra2] = f(cfg.Terra2RequestsPerMinute)
	tickers[sdk.ChainIDSui] = f(cfg.SuiRequestsPerMinute)
	tickers[sdk.ChainIDXpla] = f(cfg.XplaRequestsPerMinute)
}

func FetchTx(
	ctx context.Context,
	cfg *config.RpcProviderSettings,
	chainId sdk.ChainID,
	txHash string,
) (*TxDetail, error) {

	rateLimiter, ok := tickers[chainId]
	if !ok {
		return nil, fmt.Errorf("found no rate limiter for chain %s", chainId.String())
	}

	var fetchFunc func(context.Context, *time.Ticker, *config.RpcProviderSettings, string) (*TxDetail, error)

	// decide which RPC/API service to use based on chain ID
	switch chainId {
	case sdk.ChainIDSolana:
		fetchFunc = fetchSolanaTx
	case sdk.ChainIDAlgorand:
		fetchFunc = fetchAlgorandTx
	case sdk.ChainIDCelo:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.CeloBaseUrl, txHash)
		}
	case sdk.ChainIDEthereum:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.EthereumBaseUrl, txHash)
		}
	case sdk.ChainIDBSC:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.BscBaseUrl, txHash)
		}
	case sdk.ChainIDPolygon:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.PolygonBaseUrl, txHash)
		}
	case sdk.ChainIDFantom:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.FantomBaseUrl, txHash)
		}
	case sdk.ChainIDKlaytn:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.KlaytnBaseUrl, txHash)
		}
	case sdk.ChainIDArbitrum:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.ArbitrumBaseUrl, txHash)
		}
	case sdk.ChainIDOasis:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.OasisBaseUrl, txHash)
		}
	case sdk.ChainIDOptimism:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.OptimismBaseUrl, txHash)
		}
	case sdk.ChainIDAvalanche:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.AvalancheBaseUrl, txHash)
		}
	case sdk.ChainIDMoonbeam:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchEthTx(ctx, rateLimiter, cfg.MoonbeamBaseUrl, txHash)
		}
	case sdk.ChainIDAptos:
		fetchFunc = fetchAptosTx
	case sdk.ChainIDSui:
		fetchFunc = fetchSuiTx
	case sdk.ChainIDTerra2:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchCosmosTx(ctx, rateLimiter, cfg.Terra2BaseUrl, txHash)
		}
	case sdk.ChainIDXpla:
		fetchFunc = func(ctx context.Context, rateLimiter *time.Ticker, cfg *config.RpcProviderSettings, txHash string) (*TxDetail, error) {
			return fetchCosmosTx(ctx, rateLimiter, cfg.XplaBaseUrl, txHash)
		}
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
	txDetail, err := fetchFunc(ctx, rateLimiter, cfg, txHash)
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
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status code: %d", response.StatusCode)
	}

	// Read the response body and return
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

func waitForRateLimiter(ctx context.Context, t *time.Ticker) bool {
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}
