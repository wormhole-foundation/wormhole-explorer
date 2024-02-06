package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
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
