package chains

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

var (
	ErrChainNotSupported   = errors.New("chain id not supported")
	ErrTransactionNotFound = errors.New("transaction not found")
)

var (
	// rateLimitersByChain maps a chain ID to the request rate limiter for that chain.
	rateLimitersByChain map[sdk.ChainID]*time.Ticker
	// baseUrlsByChain maps a chain ID to the base URL of the RPC/API service for that chain.
	baseUrlsByChain map[sdk.ChainID]string
)

type TxDetail struct {
	// From is the address that signed the transaction, encoded in the chain's native format.
	From string
	// NativeTxHash contains the transaction hash, encoded in the chain's native format.
	NativeTxHash string
}

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
	rateLimitersByChain[sdk.ChainIDAcala] = convertToRateLimiter(cfg.AcalaRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDArbitrum] = convertToRateLimiter(cfg.ArbitrumRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDAlgorand] = convertToRateLimiter(cfg.AlgorandRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDAptos] = convertToRateLimiter(cfg.AptosRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDAvalanche] = convertToRateLimiter(cfg.AvalancheRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDBSC] = convertToRateLimiter(cfg.BscRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDCelo] = convertToRateLimiter(cfg.CeloRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDEthereum] = convertToRateLimiter(cfg.EthereumRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDFantom] = convertToRateLimiter(cfg.FantomRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDInjective] = convertToRateLimiter(cfg.InjectiveRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDKarura] = convertToRateLimiter(cfg.KaruraRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDKlaytn] = convertToRateLimiter(cfg.KlaytnRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDMoonbeam] = convertToRateLimiter(cfg.MoonbeamRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDOasis] = convertToRateLimiter(cfg.OasisRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDOptimism] = convertToRateLimiter(cfg.OptimismRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDPolygon] = convertToRateLimiter(cfg.PolygonRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDSolana] = convertToRateLimiter(cfg.SolanaRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDTerra] = convertToRateLimiter(cfg.TerraRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDTerra2] = convertToRateLimiter(cfg.Terra2RequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDSui] = convertToRateLimiter(cfg.SuiRequestsPerMinute)
	rateLimitersByChain[sdk.ChainIDXpla] = convertToRateLimiter(cfg.XplaRequestsPerMinute)

	// Initialize the RPC base URLs for each chain
	baseUrlsByChain = make(map[sdk.ChainID]string)
	baseUrlsByChain[sdk.ChainIDAcala] = cfg.AcalaBaseUrl
	baseUrlsByChain[sdk.ChainIDArbitrum] = cfg.ArbitrumBaseUrl
	baseUrlsByChain[sdk.ChainIDAlgorand] = cfg.AlgorandBaseUrl
	baseUrlsByChain[sdk.ChainIDAptos] = cfg.AptosBaseUrl
	baseUrlsByChain[sdk.ChainIDAvalanche] = cfg.AvalancheBaseUrl
	baseUrlsByChain[sdk.ChainIDBSC] = cfg.BscBaseUrl
	baseUrlsByChain[sdk.ChainIDCelo] = cfg.CeloBaseUrl
	baseUrlsByChain[sdk.ChainIDEthereum] = cfg.EthereumBaseUrl
	baseUrlsByChain[sdk.ChainIDFantom] = cfg.FantomBaseUrl
	baseUrlsByChain[sdk.ChainIDInjective] = cfg.InjectiveBaseUrl
	baseUrlsByChain[sdk.ChainIDKarura] = cfg.KaruraBaseUrl
	baseUrlsByChain[sdk.ChainIDKlaytn] = cfg.KlaytnBaseUrl
	baseUrlsByChain[sdk.ChainIDMoonbeam] = cfg.MoonbeamBaseUrl
	baseUrlsByChain[sdk.ChainIDOasis] = cfg.OasisBaseUrl
	baseUrlsByChain[sdk.ChainIDOptimism] = cfg.OptimismBaseUrl
	baseUrlsByChain[sdk.ChainIDPolygon] = cfg.PolygonBaseUrl
	baseUrlsByChain[sdk.ChainIDSolana] = cfg.SolanaBaseUrl
	baseUrlsByChain[sdk.ChainIDTerra] = cfg.TerraBaseUrl
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
	case sdk.ChainIDInjective,
		sdk.ChainIDTerra,
		sdk.ChainIDTerra2,
		sdk.ChainIDXpla:
		fetchFunc = fetchCosmosTx
	case sdk.ChainIDAcala,
		sdk.ChainIDArbitrum,
		sdk.ChainIDAvalanche,
		sdk.ChainIDBSC,
		sdk.ChainIDCelo,
		sdk.ChainIDEthereum,
		sdk.ChainIDFantom,
		sdk.ChainIDKarura,
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
