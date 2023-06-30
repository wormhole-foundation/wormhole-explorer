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

	"github.com/ethereum/go-ethereum/rpc"
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

var (
	// rateLimitersByChain maps a chain ID to the request rate limiter for that chain.
	rateLimitersByChain map[sdk.ChainID]*time.Ticker
	// baseUrlsByChain maps a chain ID to the base URL of the RPC/API service for that chain.
	baseUrlsByChain map[sdk.ChainID]string
)

func Initialize(cfg *config.RpcProviderSettings) {

	// convertToRateLimiter converts "requests per minute" into the associated *time.Ticker
	convertToRateLimiter := func(requestsPerMinute uint16) *time.Ticker {

		division := float64(time.Minute) / float64(time.Duration(requestsPerMinute))
		roundedUp := math.Ceil(division)

		duration := time.Duration(roundedUp)

		return time.NewTicker(duration)
	}

	// Initialize rate limiters for each chain
	rateLimitersByChain = make(map[sdk.ChainID]*time.Ticker)
	rateLimitersByChain[sdk.ChainIDArbitrum] = convertToRateLimiter(cfg.ArbitrumRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDAlgorand] = convertToRateLimiter(cfg.AlgorandRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDAptos] = convertToRateLimiter(cfg.AptosRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDAvalanche] = convertToRateLimiter(cfg.AvalancheRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDBSC] = convertToRateLimiter(cfg.BscRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDCelo] = convertToRateLimiter(cfg.CeloRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDEthereum] = convertToRateLimiter(cfg.EthereumRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDFantom] = convertToRateLimiter(cfg.FantomRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDKlaytn] = convertToRateLimiter(cfg.KlaytnRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDMoonbeam] = convertToRateLimiter(cfg.MoonbeamRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDOasis] = convertToRateLimiter(cfg.OasisRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDOptimism] = convertToRateLimiter(cfg.OptimismRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDPolygon] = convertToRateLimiter(cfg.PolygonRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDSolana] = convertToRateLimiter(cfg.SolanaRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDTerra2] = convertToRateLimiter(cfg.Terra2RequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDSui] = convertToRateLimiter(cfg.SuiRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDXpla] = convertToRateLimiter(cfg.XplaRequestsPerMinute)

	// Initialize the RPC base URLs for each chain
	baseUrlsByChain = make(map[sdk.ChainID]string)
	baseUrlsByChain[sdk.ChainIDArbitrum] = cfg.ArbitrumBaseUrl
	baseUrlsByChain[sdk.ChainIDAlgorand] = cfg.AlgorandBaseUrl
	baseUrlsByChain[sdk.ChainIDAptos] = cfg.AptosBaseUrl
	baseUrlsByChain[sdk.ChainIDAvalanche] = cfg.AvalancheBaseUrl
	baseUrlsByChain[sdk.ChainIDBSC] = cfg.BscBaseUrl
	baseUrlsByChain[sdk.ChainIDCelo] = cfg.CeloBaseUrl
	baseUrlsByChain[sdk.ChainIDEthereum] = cfg.EthereumBaseUrl
	baseUrlsByChain[sdk.ChainIDFantom] = cfg.FantomBaseUrl
	baseUrlsByChain[sdk.ChainIDKlaytn] = cfg.KlaytnBaseUrl
	baseUrlsByChain[sdk.ChainIDMoonbeam] = cfg.MoonbeamBaseUrl
	baseUrlsByChain[sdk.ChainIDOasis] = cfg.OasisBaseUrl
	baseUrlsByChain[sdk.ChainIDOptimism] = cfg.OptimismBaseUrl
	baseUrlsByChain[sdk.ChainIDPolygon] = cfg.PolygonBaseUrl
	baseUrlsByChain[sdk.ChainIDSolana] = cfg.SolanaBaseUrl
	baseUrlsByChain[sdk.ChainIDTerra2] = cfg.Terra2BaseUrl
	baseUrlsByChain[sdk.ChainIDSui] = cfg.SuiBaseUrl
	baseUrlsByChain[sdk.ChainIDXpla] = cfg.XplaBaseUrl
}

func FetchTx(
	ctx context.Context,
	cfg *config.RpcProviderSettings,
	chainId sdk.ChainID,
	txHash string,
) (*TxDetail, error) {

	// Decide which RPC/API service to use based on chain ID
	var fetchFunc func(ctx context.Context, rateLimiter *time.Ticker, baseUrl string, txHash string) (*TxDetail, error)
	switch chainId {
	case sdk.ChainIDSolana:
		fetchFunc = fetchSolanaTx
	case sdk.ChainIDAlgorand:
		fetchFunc = fetchAlgorandTx
	case sdk.ChainIDAptos:
		fetchFunc = fetchAptosTx
	case sdk.ChainIDSui:
		fetchFunc = fetchSuiTx
	case sdk.ChainIDTerra2,
		sdk.ChainIDXpla:
		fetchFunc = fetchCosmosTx
	case sdk.ChainIDArbitrum,
		sdk.ChainIDAvalanche,
		sdk.ChainIDBSC,
		sdk.ChainIDCelo,
		sdk.ChainIDEthereum,
		sdk.ChainIDFantom,
		sdk.ChainIDKlaytn,
		sdk.ChainIDMoonbeam,
		sdk.ChainIDOasis,
		sdk.ChainIDOptimism,
		sdk.ChainIDPolygon:
		fetchFunc = fetchEthTx
	default:
		return nil, ErrChainNotSupported
	}

	// Get the rate limiter and base URL for the given chain ID
	rateLimiter, ok := rateLimitersByChain[chainId]
	if !ok {
		return nil, fmt.Errorf("found no rate limiter for chain %s", chainId.String())
	}
	baseUrl, ok := baseUrlsByChain[chainId]
	if !ok {
		return nil, fmt.Errorf("found no base URL for chain %s", chainId.String())
	}

	// Get transaction details from the RPC/API service
	txDetail, err := fetchFunc(ctx, rateLimiter, baseUrl, txHash)
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
func httpGet(ctx context.Context, rateLimiter *time.Ticker, url string) ([]byte, error) {

	// Wait for the rate limiter
	if !waitForRateLimiter(ctx, rateLimiter) {
		return nil, ctx.Err()
	}

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

type rateLimitedRpcClient struct {
	client *rpc.Client
}

func rpcDialContext(ctx context.Context, url string) (*rateLimitedRpcClient, error) {

	client, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}

	tmp := rateLimitedRpcClient{
		client: client,
	}
	return &tmp, nil
}

func (c *rateLimitedRpcClient) CallContext(
	ctx context.Context,
	rateLimiter *time.Ticker,
	result interface{},
	method string,
	args ...interface{},
) error {

	if !waitForRateLimiter(ctx, rateLimiter) {
		return ctx.Err()
	}

	return c.client.CallContext(ctx, result, method, args...)
}

func (c *rateLimitedRpcClient) Close() {
	c.client.Close()
}
