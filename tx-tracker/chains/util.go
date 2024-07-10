package chains

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

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
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

// httpPost is a helper function that performs an HTTP request.
// func httpPost(ctx context.Context, rateLimiter *time.Ticker, url string, body any) ([]byte, error) {
func httpPost(ctx context.Context, url string, body any) ([]byte, error) {

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Build the HTTP request
	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

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
	result, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return result, nil
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
	result interface{},
	method string,
	args ...interface{},
) error {
	return c.client.CallContext(ctx, result, method, args...)
}

func (c *rateLimitedRpcClient) Close() {
	c.client.Close()
}

func txHashLowerCaseWith0x(v string) string {
	if strings.HasPrefix(v, "0x") {
		return strings.ToLower(v)
	}
	return "0x" + strings.ToLower(v)
}

func FormatTxHashByChain(chainId sdk.ChainID, txHash string) string {
	switch chainId {
	case sdk.ChainIDAcala,
		sdk.ChainIDArbitrum,
		sdk.ChainIDArbitrumSepolia,
		sdk.ChainIDAvalanche,
		sdk.ChainIDBase,
		sdk.ChainIDBaseSepolia,
		sdk.ChainIDBSC,
		sdk.ChainIDCelo,
		sdk.ChainIDEthereum,
		sdk.ChainIDSepolia,
		sdk.ChainIDFantom,
		sdk.ChainIDKarura,
		sdk.ChainIDKlaytn,
		sdk.ChainIDMoonbeam,
		sdk.ChainIDOasis,
		sdk.ChainIDOptimism,
		sdk.ChainIDOptimismSepolia,
		sdk.ChainIDPolygon,
		sdk.ChainIDScroll,
		sdk.ChainIDBlast,
		sdk.ChainIDXLayer,
		sdk.ChainIDMantle,
		sdk.ChainIDPolygonSepolia:
		return txHashLowerCaseWith0x(txHash)
	case sdk.ChainIDSei, sdk.ChainIDWormchain:
		return txHashLowerCaseWith0x(txHash)
	default:
		return txHash
	}
}

func CalculateFeeUSD(fee, txHash string, chainID sdk.ChainID, notionalCache *notional.NotionalCache, logger *zap.Logger) *float64 {

	var coingeckoID string
	switch chainID {
	case sdk.ChainIDSolana:
		coingeckoID = "solana"
	case sdk.ChainIDAvalanche:
		coingeckoID = "avalanche-2"
	default:
		coingeckoID = "ethereum"
	}

	price, errGetPrice := notionalCache.Get(coingeckoID)
	if errGetPrice != nil {
		logger.Error("Failed to fetch gas price", zap.String("txHash", txHash), zap.String("chainId", chainID.String()), zap.Error(errGetPrice))
		return nil
	} else {
		feeUSD, _ := price.NotionalUsd.Mul(decimal.RequireFromString(fee)).Float64() // todo: check if float64 is appropiate for feeUSD
		return &feeUSD
	}
}
