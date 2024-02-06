package chains

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

var (
	ErrChainNotSupported   = errors.New("chain id not supported")
	ErrTransactionNotFound = errors.New("transaction not found")
)

var (
	// rpcPool maps a chain ID to a list of primary and fallback RPC services for that chain.
	// The first element is the primary rpc service.
	rpcPool map[sdk.ChainID][]rpcConfig
)

// rpcConfig contains the configuration for an RPC service.
type rpcConfig struct {
	// url is the base URL of the RPC service.
	url string
	// rateLimit is the rate limiter for the RPC service.
	rateLimit *time.Ticker
	// priority is the priority of the RPC service. 1 is primary, 2 is fallback.
	priority uint8
}

type TxDetail struct {
	// From is the address that signed the transaction, encoded in the chain's native format.
	From string
	// NativeTxHash contains the transaction hash, encoded in the chain's native format.
	NativeTxHash string
	// Attribute contains the specific information of the transaction.
	Attribute *AttributeTxDetail
}

type AttributeTxDetail struct {
	Type  string
	Value any
}

func Initialize(cfg *config.RpcProviderSettings, testnetConfig *config.TestnetRpcProviderSettings) error {
	rpcPool = make(map[sdk.ChainID][]rpcConfig)

	// add Acala rpc pool configuration.
	err := addRpcConfig(sdk.ChainIDAcala,
		cfg.AcalaBaseUrl,
		convertToRateLimiter(cfg.AcalaRequestsPerMinute),
		cfg.AcalaFallbackUrls,
		cfg.AcalaFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Arbitrum rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDArbitrum,
		cfg.ArbitrumBaseUrl,
		convertToRateLimiter(cfg.ArbitrumRequestsPerMinute),
		cfg.ArbitrumFallbackUrls,
		cfg.ArbitrumFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Algorand rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDAlgorand,
		cfg.AlgorandBaseUrl,
		convertToRateLimiter(cfg.AlgorandRequestsPerMinute),
		cfg.AlgorandFallbackUrls,
		cfg.AlgorandFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Aptos rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDAptos,
		cfg.AptosBaseUrl,
		convertToRateLimiter(cfg.AptosRequestsPerMinute),
		cfg.AptosFallbackUrls,
		cfg.AptosFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Avalanche rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDAvalanche,
		cfg.AvalancheBaseUrl,
		convertToRateLimiter(cfg.AvalancheRequestsPerMinute),
		cfg.AvalancheFallbackUrls,
		cfg.AvalancheFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Base rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDBase,
		cfg.BaseBaseUrl,
		convertToRateLimiter(cfg.BaseRequestsPerMinute),
		cfg.BaseFallbackUrls,
		cfg.BaseFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add BSC rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDBSC,
		cfg.BscBaseUrl,
		convertToRateLimiter(cfg.BscRequestsPerMinute),
		cfg.BscFallbackUrls,
		cfg.BscFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Celo rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDCelo,
		cfg.CeloBaseUrl,
		convertToRateLimiter(cfg.CeloRequestsPerMinute),
		cfg.CeloFallbackUrls,
		cfg.CeloFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Ethereum rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDEthereum,
		cfg.EthereumBaseUrl,
		convertToRateLimiter(cfg.EthereumRequestsPerMinute),
		cfg.EthereumFallbackUrls,
		cfg.EthereumFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Evmos rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDEvmos,
		cfg.EvmosBaseUrl,
		convertToRateLimiter(cfg.EvmosRequestsPerMinute),
		cfg.EvmosFallbackUrls,
		cfg.EvmosFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Fantom rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDFantom,
		cfg.FantomBaseUrl,
		convertToRateLimiter(cfg.FantomRequestsPerMinute),
		cfg.FantomFallbackUrls,
		cfg.FantomFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Injective rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDInjective,
		cfg.InjectiveBaseUrl,
		convertToRateLimiter(cfg.InjectiveRequestsPerMinute),
		cfg.InjectiveFallbackUrls,
		cfg.InjectiveFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Karura rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDKarura,
		cfg.KaruraBaseUrl,
		convertToRateLimiter(cfg.KaruraRequestsPerMinute),
		cfg.KaruraFallbackUrls,
		cfg.KaruraFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Klaytn rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDKlaytn,
		cfg.KlaytnBaseUrl,
		convertToRateLimiter(cfg.KlaytnRequestsPerMinute),
		cfg.KlaytnFallbackUrls,
		cfg.KlaytnFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Kujira rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDKujira,
		cfg.KujiraBaseUrl,
		convertToRateLimiter(cfg.KujiraRequestsPerMinute),
		cfg.KujiraFallbackUrls,
		cfg.KujiraFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Moonbeam rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDMoonbeam,
		cfg.MoonbeamBaseUrl,
		convertToRateLimiter(cfg.MoonbeamRequestsPerMinute),
		cfg.MoonbeamFallbackUrls,
		cfg.MoonbeamFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Oasis rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDOasis,
		cfg.OasisBaseUrl,
		convertToRateLimiter(cfg.OasisRequestsPerMinute),
		cfg.OasisFallbackUrls,
		cfg.OasisFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Optimism rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDOptimism,
		cfg.OptimismBaseUrl,
		convertToRateLimiter(cfg.OptimismRequestsPerMinute),
		cfg.OptimismFallbackUrls,
		cfg.OptimismFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Osmosis rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDOsmosis,
		cfg.OsmosisBaseUrl,
		convertToRateLimiter(cfg.OsmosisRequestsPerMinute),
		cfg.OsmosisFallbackUrls,
		cfg.OsmosisFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Polygon rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDPolygon,
		cfg.PolygonBaseUrl,
		convertToRateLimiter(cfg.PolygonRequestsPerMinute),
		cfg.PolygonFallbackUrls,
		cfg.PolygonFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Sei rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDSei,
		cfg.SeiBaseUrl,
		convertToRateLimiter(cfg.SeiRequestsPerMinute),
		cfg.SeiFallbackUrls,
		cfg.SeiFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Solana rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDSolana,
		cfg.SolanaBaseUrl,
		convertToRateLimiter(cfg.SolanaRequestsPerMinute),
		cfg.SolanaFallbackUrls,
		cfg.SolanaFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Sui rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDSui,
		cfg.SuiBaseUrl,
		convertToRateLimiter(cfg.SuiRequestsPerMinute),
		cfg.SuiFallbackUrls,
		cfg.SuiFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Terra rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDTerra,
		cfg.TerraBaseUrl,
		convertToRateLimiter(cfg.TerraRequestsPerMinute),
		cfg.TerraFallbackUrls,
		cfg.TerraFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Terra2 rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDTerra2,
		cfg.Terra2BaseUrl,
		convertToRateLimiter(cfg.Terra2RequestsPerMinute),
		cfg.Terra2FallbackUrls,
		cfg.Terra2FallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	// add Wormchain rpc pool configuration.
	err = addRpcConfig(sdk.ChainIDWormchain,
		cfg.WormchainBaseUrl,
		convertToRateLimiter(cfg.WormchainRequestsPerMinute),
		cfg.WormchainFallbackUrls,
		cfg.WormchainFallbackRequestsPerMinute)
	if err != nil {
		return err
	}

	if testnetConfig != nil {
		// add ArbitrumSepolia rpc pool configuration.
		err = addRpcConfig(sdk.ChainIDArbitrumSepolia,
			testnetConfig.ArbitrumSepoliaBaseUrl,
			convertToRateLimiter(testnetConfig.ArbitrumSepoliaRequestsPerMinute),
			testnetConfig.ArbitrumSepoliaFallbackUrls,
			testnetConfig.ArbitrumSepoliaFallbackRequestsPerMinute)
		if err != nil {
			return err
		}

		// add BaseSepolia rpc pool configuration.
		err = addRpcConfig(sdk.ChainIDBaseSepolia,
			testnetConfig.BaseSepoliaBaseUrl,
			convertToRateLimiter(testnetConfig.BaseSepoliaRequestsPerMinute),
			testnetConfig.BaseSepoliaFallbackUrls,
			testnetConfig.BaseSepoliaFallbackRequestsPerMinute)
		if err != nil {
			return err
		}

		// add EthereumSepolia rpc pool configuration.
		err = addRpcConfig(sdk.ChainIDSepolia,
			testnetConfig.EthereumSepoliaBaseUrl,
			convertToRateLimiter(testnetConfig.EthereumSepoliaRequestsPerMinute),
			testnetConfig.EthereumSepoliaFallbackUrls,
			testnetConfig.EthereumSepoliaFallbackRequestsPerMinute)
		if err != nil {
			return err
		}

		// add OptimismSepolia rpc pool configuration.
		err = addRpcConfig(sdk.ChainIDOptimismSepolia,
			testnetConfig.OptimismSepoliaBaseUrl,
			convertToRateLimiter(testnetConfig.OptimismSepoliaRequestsPerMinute),
			testnetConfig.OptimismSepoliaFallbackUrls,
			testnetConfig.OptimismSepoliaFallbackRequestsPerMinute)
		if err != nil {
			return err
		}
	}

	return nil
}

func addRpcConfig(chainID sdk.ChainID, primaryUrl string, primaryRateLimit *time.Ticker, fallbackUrls string, fallbackRateLimits string) error {
	rpcPool[chainID] = append(rpcPool[chainID], rpcConfig{
		url:       primaryUrl,
		rateLimit: primaryRateLimit,
		priority:  1,
	})

	// check if the fallback urls are empty
	if fallbackUrls == "" {
		return nil
	}

	fallback := strings.Split(fallbackUrls, ",")
	sFallbackRequestPerMinute := strings.Split(fallbackRateLimits, ",")

	// check if the number of fallback urls and fallback rate limits are matched
	if len(fallback) != len(sFallbackRequestPerMinute) {
		return errors.New("fallback urls and fallback rate limits are not matched")
	}

	// add fallback rpcs
	for i, v := range sFallbackRequestPerMinute {
		uRateLimiter, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return err
		}
		rpcPool[chainID] = append(rpcPool[chainID], rpcConfig{
			url:       fallback[i],
			rateLimit: convertToRateLimiter(uint16(uRateLimiter)),
			priority:  2,
		})
	}
	return nil
}

func convertToRateLimiter(requestsPerMinute uint16) *time.Ticker {
	division := float64(time.Minute) / float64(time.Duration(requestsPerMinute))
	roundedUp := math.Ceil(division)
	duration := time.Duration(roundedUp)
	return time.NewTicker(duration)
}

func FetchTx(
	ctx context.Context,
	cfg *config.RpcProviderSettings,
	chainId sdk.ChainID,
	txHash string,
	timestamp *time.Time,
	p2pNetwork string,
	logger *zap.Logger,
) (*TxDetail, error) {

	// Decide which RPC/API service to use based on chain ID
	var fetchFunc func(ctx context.Context, rateLimiter *time.Ticker, baseUrl string, txHash string) (*TxDetail, error)
	switch chainId {
	case sdk.ChainIDSolana:
		apiSolana := &apiSolana{
			timestamp: timestamp,
		}
		fetchFunc = apiSolana.fetchSolanaTx
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
		sdk.ChainIDPolygon:
		fetchFunc = fetchEthTx
	case sdk.ChainIDWormchain:
		// TODO
		//rateLimiter, ok := rateLimitersByChain[sdk.ChainIDOsmosis]
		var ok bool
		var rateLimiter *time.Ticker
		if !ok {
			return nil, errors.New("found no rate limiter for chain osmosis")
		}
		apiWormchain := &apiWormchain{
			osmosisUrl:         cfg.OsmosisBaseUrl,
			osmosisRateLimiter: rateLimiter,
			evmosUrl:           cfg.EvmosBaseUrl,
			evmosRateLimiter:   rateLimiter,
			kujiraUrl:          cfg.KujiraBaseUrl,
			kujiraRateLimiter:  rateLimiter,
			p2pNetwork:         p2pNetwork,
		}
		fetchFunc = apiWormchain.fetchWormchainTx
	case sdk.ChainIDSei:
		// TODO
		//rateLimiter, ok := rateLimitersByChain[sdk.ChainIDWormchain]
		var ok bool
		var rateLimiter *time.Ticker
		if !ok {
			return nil, errors.New("found no rate limiter for chain osmosis")
		}
		apiSei := &apiSei{
			wormchainRateLimiter: rateLimiter,
			wormchainUrl:         cfg.WormchainBaseUrl,
			p2pNetwork:           p2pNetwork,
		}
		fetchFunc = apiSei.fetchSeiTx

	default:
		return nil, ErrChainNotSupported
	}

	// check if the chain is supported
	if _, ok := rpcPool[chainId]; !ok {
		logger.Error("not found rpc pool configuration", zap.String("chainId", chainId.String()))
		return nil, ErrChainNotSupported
	}

	// Fetch transactions from the pool of RPC services
	for _, rpc := range rpcPool[chainId] {

		TxDetail, err := fetchFunc(ctx, rpc.rateLimit, rpc.url, txHash)
		if err == nil {
			logger.Debug("Fetched transaction details",
				zap.String("txHash", txHash),
				zap.String("chainId", chainId.String()),
				zap.String("from", TxDetail.From))
			return TxDetail, nil
		}
	}

	return nil, errors.New("failed to fetch transaction details")
}
