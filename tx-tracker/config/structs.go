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
	LogLevel       string `split_words:"true" default:"INFO"`
	PprofEnabled   bool   `split_words:"true" default:"false"`

	AwsSettings
	MongodbSettings
	RpcProviderSettings
}

type AwsSettings struct {
	AwsEndpoint        string `split_words:"true" required:"false"`
	AwsAccessKeyID     string `split_words:"true" required:"false"`
	AwsSecretAccessKey string `split_words:"true" required:"false"`
	AwsRegion          string `split_words:"true" required:"true"`
	SqsUrl             string `split_words:"true" required:"true"`
}

type MongodbSettings struct {
	MongodbUri      string `split_words:"true" required:"true"`
	MongodbDatabase string `split_words:"true" required:"true"`
}

type RpcProviderSettings struct {
	ArbitrumBaseUrl            string `split_words:"true" required:"true"`
	ArbitrumRequestsPerMinute  uint16 `split_words:"true" required:"true"`
	AvalancheBaseUrl           string `split_words:"true" required:"true"`
	AvalancheRequestsPerMinute uint16 `split_words:"true" required:"true"`
	BscBaseUrl                 string `split_words:"true" required:"true"`
	BscRequestsPerMinute       uint16 `split_words:"true" required:"true"`
	CeloBaseUrl                string `split_words:"true" required:"true"`
	CeloRequestsPerMinute      uint16 `split_words:"true" required:"true"`
	EthereumBaseUrl            string `split_words:"true" required:"true"`
	EthereumRequestsPerMinute  uint16 `split_words:"true" required:"true"`
	FantomBaseUrl              string `split_words:"true" required:"true"`
	FantomRequestsPerMinute    uint16 `split_words:"true" required:"true"`
	OptimismBaseUrl            string `split_words:"true" required:"true"`
	OptimismRequestsPerMinute  uint16 `split_words:"true" required:"true"`
	PolygonBaseUrl             string `split_words:"true" required:"true"`
	PolygonRequestsPerMinute   uint16 `split_words:"true" required:"true"`
	SolanaBaseUrl              string `split_words:"true" required:"true"`
	SolanaRequestsPerMinute    uint16 `split_words:"true" required:"true"`
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
