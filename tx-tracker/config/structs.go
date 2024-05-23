package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type ServiceSettings struct {
	// MonitoringPort defines the TCP port for the /health and /ready endpoints.
	MonitoringPort      string `split_words:"true" default:"8000"`
	Environment         string `split_words:"true" required:"true"`
	LogLevel            string `split_words:"true" default:"INFO"`
	PprofEnabled        bool   `split_words:"true" default:"false"`
	MetricsEnabled      bool   `split_words:"true" default:"false"`
	P2pNetwork          string `split_words:"true" required:"true"`
	RpcProviderPath     string `split_words:"true" required:"false"`
	ConsumerWorkersSize int    `split_words:"true" default:"10"`
	AwsSettings
	MongodbSettings
	*RpcProviderSettings        `required:"false"`
	*WormchainProviderSettings  `required:"false"`
	*TestnetRpcProviderSettings `required:"false"`
	*RpcProviderSettingsJson    `required:"false"`
}

type RpcProviderSettingsJson struct {
	RpcProviders          []ChainRpcProviderSettings `json:"rpcProviders"`
	WormchainRpcProviders []ChainRpcProviderSettings `json:"wormchainRpcProviders"`
}

type ChainRpcProviderSettings struct {
	ChainId     uint16        `json:"chainId"`
	Chain       string        `json:"chain"`
	RpcSettings []RpcSettings `json:"rpcs"`
}

type RpcSettings struct {
	Url              string `json:"url"`
	RequestPerMinute uint16 `json:"requestPerMinute"`
	Priority         uint8  `json:"priority"`
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
	AcalaBaseUrl                       string `split_words:"true" required:"false"`
	AcalaRequestsPerMinute             uint16 `split_words:"true" required:"false"`
	AcalaFallbackUrls                  string `split_words:"true" required:"false"`
	AcalaFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	AlgorandBaseUrl                    string `split_words:"true" required:"false"`
	AlgorandRequestsPerMinute          uint16 `split_words:"true" required:"false"`
	AlgorandFallbackUrls               string `split_words:"true" required:"false"`
	AlgorandFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	AptosBaseUrl                       string `split_words:"true" required:"false"`
	AptosRequestsPerMinute             uint16 `split_words:"true" required:"false"`
	AptosFallbackUrls                  string `split_words:"true" required:"false"`
	AptosFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	ArbitrumBaseUrl                    string `split_words:"true" required:"false"`
	ArbitrumRequestsPerMinute          uint16 `split_words:"true" required:"false"`
	ArbitrumFallbackUrls               string `split_words:"true" required:"false"`
	ArbitrumFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	AvalancheBaseUrl                   string `split_words:"true" required:"false"`
	AvalancheRequestsPerMinute         uint16 `split_words:"true" required:"false"`
	AvalancheFallbackUrls              string `split_words:"true" required:"false"`
	AvalancheFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	BaseBaseUrl                        string `split_words:"true" required:"false"`
	BaseRequestsPerMinute              uint16 `split_words:"true" required:"false"`
	BaseFallbackUrls                   string `split_words:"true" required:"false"`
	BaseFallbackRequestsPerMinute      string `split_words:"true" required:"false"`
	BlastBaseUrl                       string `split_words:"true" required:"false"`
	BlastRequestsPerMinute             uint16 `split_words:"true" required:"false"`
	BlastFallbackUrls                  string `split_words:"true" required:"false"`
	BlastFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	XLayerBaseUrl                      string `split_words:"true" required:"false"`
	XLayerRequestsPerMinute            uint16 `split_words:"true" required:"false"`
	XLayerFallbackUrls                 string `split_words:"true" required:"false"`
	XLayerFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	BscBaseUrl                         string `split_words:"true" required:"false"`
	BscRequestsPerMinute               uint16 `split_words:"true" required:"false"`
	BscFallbackUrls                    string `split_words:"true" required:"false"`
	BscFallbackRequestsPerMinute       string `split_words:"true" required:"false"`
	CeloBaseUrl                        string `split_words:"true" required:"false"`
	CeloRequestsPerMinute              uint16 `split_words:"true" required:"false"`
	CeloFallbackUrls                   string `split_words:"true" required:"false"`
	CeloFallbackRequestsPerMinute      string `split_words:"true" required:"false"`
	EthereumBaseUrl                    string `split_words:"true" required:"false"`
	EthereumRequestsPerMinute          uint16 `split_words:"true" required:"false"`
	EthereumFallbackUrls               string `split_words:"true" required:"false"`
	EthereumFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	FantomBaseUrl                      string `split_words:"true" required:"false"`
	FantomRequestsPerMinute            uint16 `split_words:"true" required:"false"`
	FantomFallbackUrls                 string `split_words:"true" required:"false"`
	FantomFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	InjectiveBaseUrl                   string `split_words:"true" required:"false"`
	InjectiveRequestsPerMinute         uint16 `split_words:"true" required:"false"`
	InjectiveFallbackUrls              string `split_words:"true" required:"false"`
	InjectiveFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	KaruraBaseUrl                      string `split_words:"true" required:"false"`
	KaruraRequestsPerMinute            uint16 `split_words:"true" required:"false"`
	KaruraFallbackUrls                 string `split_words:"true" required:"false"`
	KaruraFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	KlaytnBaseUrl                      string `split_words:"true" required:"false"`
	KlaytnRequestsPerMinute            uint16 `split_words:"true" required:"false"`
	KlaytnFallbackUrls                 string `split_words:"true" required:"false"`
	KlaytnFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	MoonbeamBaseUrl                    string `split_words:"true" required:"false"`
	MoonbeamRequestsPerMinute          uint16 `split_words:"true" required:"false"`
	MoonbeamFallbackUrls               string `split_words:"true" required:"false"`
	MoonbeamFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	OasisBaseUrl                       string `split_words:"true" required:"false"`
	OasisRequestsPerMinute             uint16 `split_words:"true" required:"false"`
	OasisFallbackUrls                  string `split_words:"true" required:"false"`
	OasisFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	OptimismBaseUrl                    string `split_words:"true" required:"false"`
	OptimismRequestsPerMinute          uint16 `split_words:"true" required:"false"`
	OptimismFallbackUrls               string `split_words:"true" required:"false"`
	OptimismFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
	PolygonBaseUrl                     string `split_words:"true" required:"false"`
	PolygonRequestsPerMinute           uint16 `split_words:"true" required:"false"`
	PolygonFallbackUrls                string `split_words:"true" required:"false"`
	PolygonFallbackRequestsPerMinute   string `split_words:"true" required:"false"`
	ScrollBaseUrl                      string `split_words:"true" required:"false"`
	ScrollRequestsPerMinute            uint16 `split_words:"true" required:"false"`
	ScrollFallbackUrls                 string `split_words:"true" required:"false"`
	ScrollFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	SeiBaseUrl                         string `split_words:"true" required:"false"`
	SeiRequestsPerMinute               uint16 `split_words:"true" required:"false"`
	SeiFallbackUrls                    string `split_words:"true" required:"false"`
	SeiFallbackRequestsPerMinute       string `split_words:"true" required:"false"`
	SolanaBaseUrl                      string `split_words:"true" required:"false"`
	SolanaRequestsPerMinute            uint16 `split_words:"true" required:"false"`
	SolanaFallbackUrls                 string `split_words:"true" required:"false"`
	SolanaFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	SuiBaseUrl                         string `split_words:"true" required:"false"`
	SuiRequestsPerMinute               uint16 `split_words:"true" required:"false"`
	SuiFallbackUrls                    string `split_words:"true" required:"false"`
	SuiFallbackRequestsPerMinute       string `split_words:"true" required:"false"`
	TerraBaseUrl                       string `split_words:"true" required:"false"`
	TerraRequestsPerMinute             uint16 `split_words:"true" required:"false"`
	TerraFallbackUrls                  string `split_words:"true" required:"false"`
	TerraFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	Terra2BaseUrl                      string `split_words:"true" required:"false"`
	Terra2RequestsPerMinute            uint16 `split_words:"true" required:"false"`
	Terra2FallbackUrls                 string `split_words:"true" required:"false"`
	Terra2FallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	XplaBaseUrl                        string `split_words:"true" required:"false"`
	XplaRequestsPerMinute              uint16 `split_words:"true" required:"false"`
	XplaFallbackUrls                   string `split_words:"true" required:"false"`
	XplaFallbackRequestsPerMinute      string `split_words:"true" required:"false"`
	WormchainBaseUrl                   string `split_words:"true" required:"false"`
	WormchainRequestsPerMinute         uint16 `split_words:"true" required:"false"`
	WormchainFallbackUrls              string `split_words:"true" required:"false"`
	WormchainFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	WormchainProviderSettings
}

type WormchainProviderSettings struct {
	WormchainEvmosBaseUrl                       string `split_words:"true" required:"false"`
	WormchainEvmosRequestsPerMinute             uint16 `split_words:"true" required:"false"`
	WormchainEvmosFallbackUrls                  string `split_words:"true" required:"false"`
	WormchainEvmosFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	WormchainKujiraBaseUrl                      string `split_words:"true" required:"false"`
	WormchainKujiraRequestsPerMinute            uint16 `split_words:"true" required:"false"`
	WormchainKujiraFallbackUrls                 string `split_words:"true" required:"false"`
	WormchainKujiraFallbackRequestsPerMinute    string `split_words:"true" required:"false"`
	WormchainOsmosisBaseUrl                     string `split_words:"true" required:"false"`
	WormchainOsmosisRequestsPerMinute           uint16 `split_words:"true" required:"false"`
	WormchainOsmosisFallbackUrls                string `split_words:"true" required:"false"`
	WormchainOsmosisFallbackRequestsPerMinute   string `split_words:"true" required:"false"`
	WormchainInjectiveBaseUrl                   string `split_words:"true" required:"false"`
	WormchainInjectiveRequestsPerMinute         uint16 `split_words:"true" required:"false"`
	WormchainInjectiveFallbackUrls              string `split_words:"true" required:"false"`
	WormchainInjectiveFallbackRequestsPerMinute string `split_words:"true" required:"false"`
}

type TestnetRpcProviderSettings struct {
	ArbitrumSepoliaBaseUrl                   string `split_words:"true" required:"false"`
	ArbitrumSepoliaRequestsPerMinute         uint16 `split_words:"true" required:"false"`
	ArbitrumSepoliaFallbackUrls              string `split_words:"true" required:"false"`
	ArbitrumSepoliaFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	BaseSepoliaBaseUrl                       string `split_words:"true" required:"false"`
	BaseSepoliaRequestsPerMinute             uint16 `split_words:"true" required:"false"`
	BaseSepoliaFallbackUrls                  string `split_words:"true" required:"false"`
	BaseSepoliaFallbackRequestsPerMinute     string `split_words:"true" required:"false"`
	EthereumSepoliaBaseUrl                   string `split_words:"true" required:"false"`
	EthereumSepoliaRequestsPerMinute         uint16 `split_words:"true" required:"false"`
	EthereumSepoliaFallbackUrls              string `split_words:"true" required:"false"`
	EthereumSepoliaFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	OptimismSepoliaBaseUrl                   string `split_words:"true" required:"false"`
	OptimismSepoliaRequestsPerMinute         uint16 `split_words:"true" required:"false"`
	OptimismSepoliaFallbackUrls              string `split_words:"true" required:"false"`
	OptimismSepoliaFallbackRequestsPerMinute string `split_words:"true" required:"false"`
	PolygonSepoliaBaseUrl                    string `split_words:"true" required:"false"`
	PolygonSepoliaRequestsPerMinute          uint16 `split_words:"true" required:"false"`
	PolygonSepoliaFallbackUrls               string `split_words:"true" required:"false"`
	PolygonSepoliaFallbackRequestsPerMinute  string `split_words:"true" required:"false"`
}

func NewRpcProviderSettingJson(path string) (*RpcProviderSettingsJson, error) {

	rpcJsonFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read rpc provider settings from file: %w", err)
	}

	var rpcProviderSettingsJson RpcProviderSettingsJson
	err = json.Unmarshal(rpcJsonFile, &rpcProviderSettingsJson)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rpc provider settings from file: %w", err)
	}
	return &rpcProviderSettingsJson, nil
}

func New() (*ServiceSettings, error) {
	_ = godotenv.Load()
	var settings ServiceSettings

	err := envconfig.Process("", &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from environment: %w", err)
	}

	if settings.RpcProviderPath != "" {
		rpcJsonFile, err := os.ReadFile(settings.RpcProviderPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read rpc provider settings from file: %w", err)
		}

		var rpcProviderSettingsJson RpcProviderSettingsJson
		err = json.Unmarshal(rpcJsonFile, &rpcProviderSettingsJson)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal rpc provider settings from file: %w", err)
		}
		settings.RpcProviderSettingsJson = &rpcProviderSettingsJson
		settings.RpcProviderSettings = nil

	} else {
		rpcProviderSettings, err := LoadFromEnv[RpcProviderSettings]()
		if err != nil {
			return nil, err
		}
		settings.RpcProviderSettings = rpcProviderSettings
		settings.RpcProviderSettingsJson = nil
	}

	return &settings, nil
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

// MapRpcProviderToRpcConfig converts the RpcProviderSettings to a map of RpcConfig
func (s *ServiceSettings) MapRpcProviderToRpcConfig() (map[sdk.ChainID][]RpcConfig, map[sdk.ChainID][]RpcConfig, error) {
	if s.RpcProviderSettingsJson != nil {
		rpcPoolConfig, err := s.RpcProviderSettingsJson.ToMap()
		if err != nil {
			return nil, nil, err
		}
		wormchainRpcPoolConfig, err := s.RpcProviderSettingsJson.WormchainToMap()
		if err != nil {
			return nil, nil, err
		}
		return rpcPoolConfig, wormchainRpcPoolConfig, nil
	}
	if s.RpcProviderSettings != nil {
		rpcPoolConfig, err := s.RpcProviderSettings.ToMap()
		if err != nil {
			return nil, nil, err
		}
		wormchainRpcPoolConfig, err := s.RpcProviderSettings.WormchainProviderSettings.ToMap()
		if err != nil {
			return nil, nil, err
		}
		return rpcPoolConfig, wormchainRpcPoolConfig, nil
	}
	return nil, nil, errors.New("rpc provider settings not found")
}

// ToMap converts the RpcProviderSettingsJson to a map of RpcConfig
func (r RpcProviderSettingsJson) ToMap() (map[sdk.ChainID][]RpcConfig, error) {
	rpcs := make(map[sdk.ChainID][]RpcConfig)
	for _, rpcProvider := range r.RpcProviders {
		chainID := sdk.ChainID(rpcProvider.ChainId)
		var rpcConfigs []RpcConfig
		for _, rpcSetting := range rpcProvider.RpcSettings {
			rpcConfigs = append(rpcConfigs, RpcConfig{
				Url:               rpcSetting.Url,
				Priority:          rpcSetting.Priority,
				RequestsPerMinute: rpcSetting.RequestPerMinute,
			})
		}
		rpcs[chainID] = rpcConfigs
	}
	return rpcs, nil
}

func (r RpcProviderSettingsJson) WormchainToMap() (map[sdk.ChainID][]RpcConfig, error) {
	rpcs := make(map[sdk.ChainID][]RpcConfig)
	for _, rpcProvider := range r.WormchainRpcProviders {
		chainID := sdk.ChainID(rpcProvider.ChainId)
		var rpcConfigs []RpcConfig
		for _, rpcSetting := range rpcProvider.RpcSettings {
			rpcConfigs = append(rpcConfigs, RpcConfig{
				Url:               rpcSetting.Url,
				Priority:          rpcSetting.Priority,
				RequestsPerMinute: rpcSetting.RequestPerMinute,
			})
		}
		rpcs[chainID] = rpcConfigs
	}
	return rpcs, nil

}

func (w WormchainProviderSettings) ToMap() (map[sdk.ChainID][]RpcConfig, error) {
	rpcs := make(map[sdk.ChainID][]RpcConfig)

	// add wormchain rpcs
	wormchainRpcConfigs, err := addRpcConfig(
		w.WormchainInjectiveBaseUrl,
		w.WormchainInjectiveRequestsPerMinute,
		w.WormchainInjectiveFallbackUrls,
		w.WormchainInjectiveFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDInjective] = wormchainRpcConfigs

	// add evmos rpcs
	evmosRpcConfigs, err := addRpcConfig(
		w.WormchainEvmosBaseUrl,
		w.WormchainEvmosRequestsPerMinute,
		w.WormchainEvmosFallbackUrls,
		w.WormchainEvmosFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDEvmos] = evmosRpcConfigs

	// add kujira rpcs
	kujiraRpcConfigs, err := addRpcConfig(
		w.WormchainKujiraBaseUrl,
		w.WormchainKujiraRequestsPerMinute,
		w.WormchainKujiraFallbackUrls,
		w.WormchainKujiraFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDKujira] = kujiraRpcConfigs

	// add osmosis rpcs
	osmosisRpcConfigs, err := addRpcConfig(
		w.WormchainOsmosisBaseUrl,
		w.WormchainOsmosisRequestsPerMinute,
		w.WormchainOsmosisFallbackUrls,
		w.WormchainOsmosisFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDOsmosis] = osmosisRpcConfigs

	return rpcs, nil
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

	// add blast rpcs
	blastRpcConfigs, err := addRpcConfig(
		r.BlastBaseUrl,
		r.BlastRequestsPerMinute,
		r.BlastFallbackUrls,
		r.BlastFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDBlast] = blastRpcConfigs

	// add xlayer rpcs
	xlayerRpcConfigs, err := addRpcConfig(
		r.XplaBaseUrl,
		r.XplaRequestsPerMinute,
		r.XplaFallbackUrls,
		r.XplaFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDXLayer] = xlayerRpcConfigs

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

	// add scroll rpcs
	scrollRpcConfigs, err := addRpcConfig(
		r.ScrollBaseUrl,
		r.ScrollRequestsPerMinute,
		r.ScrollFallbackUrls,
		r.ScrollFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	rpcs[sdk.ChainIDScroll] = scrollRpcConfigs

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

	// add
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

	// add polygon sepolia rpcs
	polygonSepoliaRpcConfigs, err := addRpcConfig(
		r.PolygonSepoliaBaseUrl,
		r.PolygonSepoliaRequestsPerMinute,
		r.PolygonSepoliaFallbackUrls,
		r.PolygonSepoliaFallbackRequestsPerMinute)
	if err != nil {
		return nil, err
	}
	// polygon sepolia is the same as polygon amoy
	rpcs[sdk.ChainIDPolygonSepolia] = polygonSepoliaRpcConfigs

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
