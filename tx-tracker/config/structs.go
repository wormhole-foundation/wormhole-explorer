package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type BackfillingStrategy string

const (
	// StrategyReprocessAll will reprocess documents in the `globalTransactions`
	// collection that don't have the `sourceTx` field set, or that have the
	// `sourceTx.status` field set to "internalError".
	BackfillerStrategyReprocessFailed BackfillingStrategy = "reprocess_failed"
	// BackfillerStrategyTimeRange will reprocess all VAAs that have a timestamp between the specified range.
	BackfillerStrategyTimeRange BackfillingStrategy = "time_range"
)

type BackfillerSettings struct {
	LogLevel   string `split_words:"true" default:"INFO"`
	NumWorkers uint   `split_words:"true" required:"true"`
	BulkSize   uint   `split_words:"true" required:"true"`
	P2pNetwork string `split_words:"true" required:"true"`

	// Strategy determines which VAAs will be affected by the backfiller.
	Strategy struct {
		Name            BackfillingStrategy `split_words:"true" required:"true"`
		TimestampAfter  string              `split_words:"true" required:"false"`
		TimestampBefore string              `split_words:"true" required:"false"`
	}

	MongodbSettings
	RpcProviderSettings
}

type ServiceSettings struct {
	// MonitoringPort defines the TCP port for the /health and /ready endpoints.
	MonitoringPort string `split_words:"true" default:"8000"`
	Environment    string `split_words:"true" required:"true"`
	LogLevel       string `split_words:"true" default:"INFO"`
	PprofEnabled   bool   `split_words:"true" default:"false"`
	MetricsEnabled bool   `split_words:"true" default:"false"`
	P2pNetwork     string `split_words:"true" required:"true"`
	AwsSettings
	MongodbSettings
	RpcProviderSettings
}

type AwsSettings struct {
	AwsEndpoint         string `split_words:"true" required:"false"`
	AwsAccessKeyID      string `split_words:"true" required:"false"`
	AwsSecretAccessKey  string `split_words:"true" required:"false"`
	AwsRegion           string `split_words:"true" required:"true"`
	PipelineSqsUrl      string `split_words:"true" required:"true"`
	NotificationsSqsUrl string `split_words:"true" required:"true"`
}

type MongodbSettings struct {
	MongodbUri      string `split_words:"true" required:"true"`
	MongodbDatabase string `split_words:"true" required:"true"`
}

type RpcProviderSettings struct {
	AcalaBaseUrl                       string `split_words:"true" required:"true"`
	AcalaRequestsPerMinute             uint16 `split_words:"true" required:"true"`
	AcalaFallbackUrls                  string `split_words:"true" required:"false"`
	AcalaFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	AlgorandBaseUrl                    string `split_words:"true" required:"true"`
	AlgorandRequestsPerMinute          uint16 `split_words:"true" required:"true"`
	AlgorandFallbackUrls               string `split_words:"true" required:"false"`
	AlgorandFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	AptosBaseUrl                       string `split_words:"true" required:"true"`
	AptosRequestsPerMinute             uint16 `split_words:"true" required:"true"`
	AptosFallbackUrls                  string `split_words:"true" required:"false"`
	AptosFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	ArbitrumBaseUrl                    string `split_words:"true" required:"true"`
	ArbitrumRequestsPerMinute          uint16 `split_words:"true" required:"true"`
	ArbitrumFallbackUrls               string `split_words:"true" required:"false"`
	ArbitrumFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	AvalancheBaseUrl                   string `split_words:"true" required:"true"`
	AvalancheRequestsPerMinute         uint16 `split_words:"true" required:"true"`
	AvalancheFallbackUrls              string `split_words:"true" required:"false"`
	AvalancheFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	BaseBaseUrl                        string `split_words:"true" required:"true"`
	BaseRequestsPerMinute              uint16 `split_words:"true" required:"true"`
	BaseFallbackUrls                   string `split_words:"true" required:"false"`
	BaseFallbackRequestsPerMinute      string `split_words:"true" required:"false"`
	BscBaseUrl                         string `split_words:"true" required:"true"`
	BscRequestsPerMinute               uint16 `split_words:"true" required:"true"`
	BscFallbackUrls                    string `split_words:"true" required:"false"`
	BscFallbackRequestsPerMinute       string `split_words:"true" required:"false"`
	CeloBaseUrl                        string `split_words:"true" required:"true"`
	CeloRequestsPerMinute              uint16 `split_words:"true" required:"true"`
	CeloFallbackUrls                   string `split_words:"true" required:"false"`
	CeloFallbackRequestsPerMinute      string `split_words:"true" required:"false"`
	EthereumBaseUrl                    string `split_words:"true" required:"true"`
	EthereumRequestsPerMinute          uint16 `split_words:"true" required:"true"`
	EthereumFallbackUrls               string `split_words:"true" required:"false"`
	EthereumFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	EvmosBaseUrl                       string `split_words:"true" required:"true"`
	EvmosRequestsPerMinute             uint16 `split_words:"true" required:"true"`
	EvmosFallbackUrls                  string `split_words:"true" required:"false"`
	EvmosFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	FantomBaseUrl                      string `split_words:"true" required:"true"`
	FantomRequestsPerMinute            uint16 `split_words:"true" required:"true"`
	FantomFallbackUrls                 string `split_words:"true" required:"false"`
	FantomFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	InjectiveBaseUrl                   string `split_words:"true" required:"true"`
	InjectiveRequestsPerMinute         uint16 `split_words:"true" required:"true"`
	InjectiveFallbackUrls              string `split_words:"true" required:"false"`
	InjectiveFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	KaruraBaseUrl                      string `split_words:"true" required:"true"`
	KaruraRequestsPerMinute            uint16 `split_words:"true" required:"true"`
	KaruraFallbackUrls                 string `split_words:"true" required:"false"`
	KaruraFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	KlaytnBaseUrl                      string `split_words:"true" required:"true"`
	KlaytnRequestsPerMinute            uint16 `split_words:"true" required:"true"`
	KlaytnFallbackUrls                 string `split_words:"true" required:"false"`
	KlaytnFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	KujiraBaseUrl                      string `split_words:"true" required:"true"`
	KujiraRequestsPerMinute            uint16 `split_words:"true" required:"true"`
	KujiraFallbackUrls                 string `split_words:"true" required:"false"`
	KujiraFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	MoonbeamBaseUrl                    string `split_words:"true" required:"true"`
	MoonbeamRequestsPerMinute          uint16 `split_words:"true" required:"true"`
	MoonbeamFallbackUrls               string `split_words:"true" required:"false"`
	MoonbeamFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	OasisBaseUrl                       string `split_words:"true" required:"true"`
	OasisRequestsPerMinute             uint16 `split_words:"true" required:"true"`
	OasisFallbackUrls                  string `split_words:"true" required:"false"`
	OasisFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	OptimismBaseUrl                    string `split_words:"true" required:"true"`
	OptimismRequestsPerMinute          uint16 `split_words:"true" required:"true"`
	OptimismFallbackUrls               string `split_words:"true" required:"false"`
	OptimismFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	OsmosisBaseUrl                     string `split_words:"true" required:"true"`
	OsmosisRequestsPerMinute           uint16 `split_words:"true" required:"true"`
	OsmosisFallbackUrls                string `split_words:"true" required:"false"`
	OsmosisFallbackRequestsPerMinute   string `split_words:"true" required:"false"`
	PolygonBaseUrl                     string `split_words:"true" required:"true"`
	PolygonRequestsPerMinute           uint16 `split_words:"true" required:"true"`
	PolygonFallbackUrls                string `split_words:"true" required:"false"`
	PolygonFallbackRequestsPerMinute   string `split_words:"true" required:"false"`
	SeiBaseUrl                         string `split_words:"true" required:"true"`
	SeiRequestsPerMinute               uint16 `split_words:"true" required:"true"`
	SeiFallbackUrls                    string `split_words:"true" required:"false"`
	SeiFallbackRequestsPerMinute       string `split_words:"true" required:"false"`
	SolanaBaseUrl                      string `split_words:"true" required:"true"`
	SolanaRequestsPerMinute            uint16 `split_words:"true" required:"true"`
	SolanaFallbackUrls                 string `split_words:"true" required:"false"`
	SolanaFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	SuiBaseUrl                         string `split_words:"true" required:"true"`
	SuiRequestsPerMinute               uint16 `split_words:"true" required:"true"`
	SuiFallbackUrls                    string `split_words:"true" required:"false"`
	SuiFallbackRequestsPerMinute       string `split_words:"true" required:"false"`
	TerraBaseUrl                       string `split_words:"true" required:"true"`
	TerraRequestsPerMinute             uint16 `split_words:"true" required:"true"`
	TerraFallbackUrls                  string `split_words:"true" required:"false"`
	TerraFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	Terra2BaseUrl                      string `split_words:"true" required:"true"`
	Terra2RequestsPerMinute            uint16 `split_words:"true" required:"true"`
	Terra2FallbackUrls                 string `split_words:"true" required:"false"`
	Terra2FallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	XplaBaseUrl                        string `split_words:"true" required:"true"`
	XplaRequestsPerMinute              uint16 `split_words:"true" required:"true"`
	XplaFallbackUrls                   string `split_words:"true" required:"false"`
	XplaFallbackRequestsPerMinute      string `split_words:"true" required:"false"`
	WormchainBaseUrl                   string `split_words:"true" required:"true"`
	WormchainRequestsPerMinute         uint16 `split_words:"true" required:"true"`
	WormchainFallbackUrls              string `split_words:"true" required:"false"`
	WormchainFallbackRequestsPerMinute string `split_words:"true" required:"false"`
}

type TestnetRpcProviderSettings struct {
	ArbitrumSepoliaBaseUrl                   string `split_words:"true" required:"true"`
	ArbitrumSepoliaRequestsPerMinute         uint16 `split_words:"true" required:"true"`
	ArbitrumSepoliaFallbackUrls              string `split_words:"true" required:"false"`
	ArbitrumSepoliaFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	BaseSepoliaBaseUrl                       string `split_words:"true" required:"true"`
	BaseSepoliaRequestsPerMinute             uint16 `split_words:"true" required:"true"`
	BaseSepoliaFallbackUrls                  string `split_words:"true" required:"false"`
	BaseSepoliaFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	EthereumSepoliaBaseUrl                   string `split_words:"true" required:"true"`
	EthereumSepoliaRequestsPerMinute         uint16 `split_words:"true" required:"true"`
	EthereumSepoliaFallbackUrls              string `split_words:"true" required:"false"`
	EthereumSepoliaFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	OptimismSepoliaBaseUrl                   string `split_words:"true" required:"true"`
	OptimismSepoliaRequestsPerMinute         uint16 `split_words:"true" required:"true"`
	OptimismSepoliaFallbackUrls              string `split_words:"true" required:"false"`
	OptimismSepoliaFallbackRequestsPerMinute string `split_words:"true" required:"false"`
}

func LoadFromEnv[T any]() (*T, error) {

	_ = godotenv.Load()

	var settings T

	err := envconfig.Process("", &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from environment: %w", err)
	}

	return &settings, nil
}

// RpcConfig defines the configuration for a single RPC provider
type RpcConfig struct {
	Url               string
	Priority          uint8
	RequestsPerMinute uint16
}

// ToMap converts the RpcProviderSettings to a map of RpcConfig
func (r RpcProviderSettings) ToMap() (map[sdk.ChainID][]RpcConfig, error) {
	rpcs := make(map[sdk.ChainID][]RpcConfig)

	// add acala rpcs
	acalaRpcConfigs, err := addRpcConfig(
		r.AcalaBaseUrl,
		r.AcalaRequestsPerMinute,
		r.AcalaFallbackUrls,
		r.AcalaFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDAcala] = acalaRpcConfigs

	// add algorand rpcs
	algorandRpcConfigs, err := addRpcConfig(
		r.AlgorandBaseUrl,
		r.AlgorandRequestsPerMinute,
		r.AlgorandFallbackUrls,
		r.AlgorandFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDAlgorand] = algorandRpcConfigs

	// add aptos rpcs
	aptosRpcConfigs, err := addRpcConfig(
		r.AptosBaseUrl,
		r.AptosRequestsPerMinute,
		r.AptosFallbackUrls,
		r.AptosFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDAptos] = aptosRpcConfigs

	// add arbitrum rpcs
	arbitrumRpcConfigs, err := addRpcConfig(
		r.ArbitrumBaseUrl,
		r.ArbitrumRequestsPerMinute,
		r.ArbitrumFallbackUrls,
		r.ArbitrumFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDArbitrum] = arbitrumRpcConfigs

	// add avalanche rpcs
	avalancheRpcConfigs, err := addRpcConfig(
		r.AvalancheBaseUrl,
		r.AvalancheRequestsPerMinute,
		r.AvalancheFallbackUrls,
		r.AvalancheFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDAvalanche] = avalancheRpcConfigs

	// add base rpcs
	baseRpcConfigs, err := addRpcConfig(
		r.BaseBaseUrl,
		r.BaseRequestsPerMinute,
		r.BaseFallbackUrls,
		r.BaseFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDBase] = baseRpcConfigs

	// add bsc rpcs
	bscRpcConfigs, err := addRpcConfig(
		r.BscBaseUrl,
		r.BscRequestsPerMinute,
		r.BscFallbackUrls,
		r.BscFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDBSC] = bscRpcConfigs

	// add celo rpcs
	celoRpcConfigs, err := addRpcConfig(
		r.CeloBaseUrl,
		r.CeloRequestsPerMinute,
		r.CeloFallbackUrls,
		r.CeloFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDCelo] = celoRpcConfigs

	// add ethereum rpcs
	ethereumRpcConfigs, err := addRpcConfig(
		r.EthereumBaseUrl,
		r.EthereumRequestsPerMinute,
		r.EthereumFallbackUrls,
		r.EthereumFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDEthereum] = ethereumRpcConfigs

	// add evmos rpcs
	evmosRpcConfigs, err := addRpcConfig(
		r.EvmosBaseUrl,
		r.EvmosRequestsPerMinute,
		r.EvmosFallbackUrls,
		r.EvmosFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDEvmos] = evmosRpcConfigs

	// add fantom rpcs
	fantomRpcConfigs, err := addRpcConfig(
		r.FantomBaseUrl,
		r.FantomRequestsPerMinute,
		r.FantomFallbackUrls,
		r.FantomFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDFantom] = fantomRpcConfigs

	// add injective rpcs
	injectiveRpcConfigs, err := addRpcConfig(
		r.InjectiveBaseUrl,
		r.InjectiveRequestsPerMinute,
		r.InjectiveFallbackUrls,
		r.InjectiveFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDInjective] = injectiveRpcConfigs

	// add karura rpcs
	karuraRpcConfigs, err := addRpcConfig(
		r.KaruraBaseUrl,
		r.KaruraRequestsPerMinute,
		r.KaruraFallbackUrls,
		r.KaruraFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDKarura] = karuraRpcConfigs

	// add klaytn rpcs
	klaytnRpcConfigs, err := addRpcConfig(
		r.KlaytnBaseUrl,
		r.KlaytnRequestsPerMinute,
		r.KlaytnFallbackUrls,
		r.KlaytnFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDKlaytn] = klaytnRpcConfigs

	// add kujira rpcs
	kujiraRpcConfigs, err := addRpcConfig(
		r.KujiraBaseUrl,
		r.KujiraRequestsPerMinute,
		r.KujiraFallbackUrls,
		r.KujiraFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDKujira] = kujiraRpcConfigs

	// add moonbeam rpcs
	moonbeamRpcConfigs, err := addRpcConfig(
		r.MoonbeamBaseUrl,
		r.MoonbeamRequestsPerMinute,
		r.MoonbeamFallbackUrls,
		r.MoonbeamFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDMoonbeam] = moonbeamRpcConfigs

	// add oasis rpcs
	oasisRpcConfigs, err := addRpcConfig(
		r.OasisBaseUrl,
		r.OasisRequestsPerMinute,
		r.OasisFallbackUrls,
		r.OasisFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDOasis] = oasisRpcConfigs

	// add optimism rpcs
	optimismRpcConfigs, err := addRpcConfig(
		r.OptimismBaseUrl,
		r.OptimismRequestsPerMinute,
		r.OptimismFallbackUrls,
		r.OptimismFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDOptimism] = optimismRpcConfigs

	// add osmosis rpcs
	osmosisRpcConfigs, err := addRpcConfig(
		r.OsmosisBaseUrl,
		r.OsmosisRequestsPerMinute,
		r.OsmosisFallbackUrls,
		r.OsmosisFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDOsmosis] = osmosisRpcConfigs

	// add polygon rpcs
	polygonRpcConfigs, err := addRpcConfig(
		r.PolygonBaseUrl,
		r.PolygonRequestsPerMinute,
		r.PolygonFallbackUrls,
		r.PolygonFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDPolygon] = polygonRpcConfigs

	// add sei rpcs
	seiRpcConfigs, err := addRpcConfig(
		r.SeiBaseUrl,
		r.SeiRequestsPerMinute,
		r.SeiFallbackUrls,
		r.SeiFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDSei] = seiRpcConfigs

	// add solana rpcs
	solanaRpcConfigs, err := addRpcConfig(
		r.SolanaBaseUrl,
		r.SolanaRequestsPerMinute,
		r.SolanaFallbackUrls,
		r.SolanaFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDSolana] = solanaRpcConfigs

	// add sui rpcs
	suiRpcConfigs, err := addRpcConfig(
		r.SuiBaseUrl,
		r.SuiRequestsPerMinute,
		r.SuiFallbackUrls,
		r.SuiFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDSui] = suiRpcConfigs

	// add terra rpcs
	terraRpcConfigs, err := addRpcConfig(
		r.TerraBaseUrl,
		r.TerraRequestsPerMinute,
		r.TerraFallbackUrls,
		r.TerraFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDTerra] = terraRpcConfigs

	// add terra2 rpcs
	terra2RpcConfigs, err := addRpcConfig(
		r.Terra2BaseUrl,
		r.Terra2RequestsPerMinute,
		r.Terra2FallbackUrls,
		r.Terra2FallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDTerra2] = terra2RpcConfigs

	// add xpla rpcs
	xplaRpcConfigs, err := addRpcConfig(
		r.XplaBaseUrl,
		r.XplaRequestsPerMinute,
		r.XplaFallbackUrls,
		r.XplaFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDXpla] = xplaRpcConfigs

	// add wormchain rpcs
	wormchainRpcConfigs, err := addRpcConfig(
		r.WormchainBaseUrl,
		r.WormchainRequestsPerMinute,
		r.WormchainFallbackUrls,
		r.WormchainFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDWormchain] = wormchainRpcConfigs
	return rpcs, nil
}

// ToMap converts the TestnetRpcProviderSettings to a map of RpcConfig
func (r TestnetRpcProviderSettings) ToMap() (map[sdk.ChainID][]RpcConfig, error) {
	rpcs := make(map[sdk.ChainID][]RpcConfig)

	// add arbitrum sepolia rpcs
	arbitrumSepoliaRpcConfigs, err := addRpcConfig(
		r.ArbitrumSepoliaBaseUrl,
		r.ArbitrumSepoliaRequestsPerMinute,
		r.ArbitrumSepoliaFallbackUrls,
		r.ArbitrumSepoliaFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDArbitrumSepolia] = arbitrumSepoliaRpcConfigs

	// add base sepolia rpcs
	baseSepoliaRpcConfigs, err := addRpcConfig(
		r.BaseSepoliaBaseUrl,
		r.BaseSepoliaRequestsPerMinute,
		r.BaseSepoliaFallbackUrls,
		r.BaseSepoliaFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDBaseSepolia] = baseSepoliaRpcConfigs

	// add ethereum sepolia rpcs
	ethereumSepoliaRpcConfigs, err := addRpcConfig(
		r.EthereumSepoliaBaseUrl,
		r.EthereumSepoliaRequestsPerMinute,
		r.EthereumSepoliaFallbackUrls,
		r.EthereumSepoliaFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDSepolia] = ethereumSepoliaRpcConfigs

	// add optimism sepolia rpcs
	optimismSepoliaRpcConfigs, err := addRpcConfig(
		r.OptimismSepoliaBaseUrl,
		r.OptimismSepoliaRequestsPerMinute,
		r.OptimismSepoliaFallbackUrls,
		r.OptimismSepoliaFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDOptimismSepolia] = optimismSepoliaRpcConfigs
	return rpcs, nil
}

// addRpcConfig convert chain rpc settings to RpcConfig
func addRpcConfig(baseURl string, requestPerMinute uint16, fallbackUrls string, fallbackRequestPerMinute string) ([]RpcConfig, error) {
	// check if the primary rpc url and rate limit are empty
	if baseURl == "" {
		return []RpcConfig{}, errors.New("primary rpc url is empty")
	}
	if requestPerMinute == 0 {
		return []RpcConfig{}, errors.New("primary rpc rate limit is 0")
	}

	var rpcConfigs []RpcConfig
	// add primary rpc
	rpcConfigs = append(rpcConfigs, RpcConfig{
		Url:               baseURl,
		Priority:          1,
		RequestsPerMinute: requestPerMinute,
	})
	// add fallback rpc
	if fallbackUrls == "" {
		return rpcConfigs, nil
	}
	sfallbackUrls := strings.Split(fallbackUrls, ",")
	sFallbackRequestPerMinute := strings.Split(fallbackRequestPerMinute, ",")

	// check if the number of fallback urls and fallback rate limits are matched
	if len(sfallbackUrls) != len(sFallbackRequestPerMinute) {
		return rpcConfigs, errors.New("fallback urls and fallback rate limits are not matched")
	}
	// add fallback rpcs
	for i, v := range sFallbackRequestPerMinute {
		uRateLimiter, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return rpcConfigs, err
		}
		rpcConfigs = append(rpcConfigs, RpcConfig{
			Url:               sfallbackUrls[i],
			Priority:          2,
			RequestsPerMinute: uint16(uRateLimiter),
		})
	}
	return rpcConfigs, nil
}

// func addRpcConfig(chainID sdk.ChainID, primaryUrl string, primaryRateLimit *time.Ticker, fallbackUrls string, fallbackRateLimits string) error {
// 	rpcPool[chainID] = append(rpcPool[chainID], rpcConfig{
// 		url:       primaryUrl,
// 		rateLimit: primaryRateLimit,
// 		priority:  1,
// 	})

// 	// check if the fallback urls are empty
// 	if fallbackUrls == "" {
// 		return nil
// 	}

// 	fallback := strings.Split(fallbackUrls, ",")
// 	sFallbackRequestPerMinute := strings.Split(fallbackRateLimits, ",")

// 	// check if the number of fallback urls and fallback rate limits are matched
// 	if len(fallback) != len(sFallbackRequestPerMinute) {
// 		return errors.New("fallback urls and fallback rate limits are not matched")
// 	}

// 	// add fallback rpcs
// 	for i, v := range sFallbackRequestPerMinute {
// 		uRateLimiter, err := strconv.ParseUint(v, 10, 64)
// 		if err != nil {
// 			return err
// 		}
// 		rpcPool[chainID] = append(rpcPool[chainID], rpcConfig{
// 			url:       fallback[i],
// 			rateLimit: convertToRateLimiter(uint16(uRateLimiter)),
// 			priority:  2,
// 		})
// 	}
// 	return nil
// }
