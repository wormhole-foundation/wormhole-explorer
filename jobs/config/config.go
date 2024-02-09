// Package config implement a simple configuration package.
// It define a type [Configuration] that represent the aplication configuration
package config

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// Configuration is the configuration for the job
type Configuration struct {
	JobID    string `env:"JOB_ID,required"`
	LogLevel string `env:"LOG_LEVEL,default=INFO"`
}

type NotionalConfiguration struct {
	Environment     string `env:"ENVIRONMENT,required"`
	CoingeckoURL    string `env:"COINGECKO_URL,required"`
	CacheURL        string `env:"CACHE_URL,required"`
	CachePrefix     string `env:"CACHE_PREFIX,required"`
	NotionalChannel string `env:"NOTIONAL_CHANNEL,required"`
	P2pNetwork      string `env:"P2P_NETWORK,required"`
}

type TransferReportConfiguration struct {
	MongoURI      string `env:"MONGODB_URI,required"`
	MongoDatabase string `env:"MONGODB_DATABASE,required"`
	PageSize      int64  `env:"PAGE_SIZE,default=100"`
	PricesType    string `env:"PRICES_TYPE,required"`
	PricesUri     string `env:"PRICES_URI,required"`
	OutputPath    string `env:"OUTPUT_PATH,required"`
	P2pNetwork    string `env:"P2P_NETWORK,required"`
}

type HistoricalPricesConfiguration struct {
	MongoURI                string `env:"MONGODB_URI,required"`
	MongoDatabase           string `env:"MONGODB_DATABASE,required"`
	P2pNetwork              string `env:"P2P_NETWORK,required"`
	CoingeckoURL            string `env:"COINGECKO_URL,required"`
	CoingeckoHeaderKey      string `env:"COINGECKO_HEADER_KEY"`
	CoingeckoApiKey         string `env:"COINGECKO_API_KEY"`
	RequestLimitTimeSeconds int    `env:"REQUEST_LIMIT_TIME_SECONDS,default=5"`
	PriceDays               string `env:"PRICE_DAYS,default=max"`
}

type MigrateSourceTxConfiguration struct {
	MongoURI         string `env:"MONGODB_URI,required"`
	MongoDatabase    string `env:"MONGODB_DATABASE,required"`
	PageSize         int    `env:"PAGE_SIZE,default=100"`
	ChainID          int64  `env:"CHAIN_ID,default=0"`
	FromDate         string `env:"FROM_DATE,required"`
	ToDate           string `env:"TO_DATE,required"`
	TxTrackerURL     string `env:"TX_TRACKER_URL,required"`
	TxTrackerTimeout int64  `env:"TX_TRACKER_TIMEOUT,default=30"`
	SleepTimeSeconds int64  `env:"SLEEP_TIME_SECONDS,default=5"`
}

type ContributorsStatsConfiguration struct {
	InfluxUrl          string `env:"INFLUX_URL"`
	InfluxToken        string `env:"INFLUX_TOKEN"`
	InfluxOrganization string `env:"INFLUX_ORGANIZATION"`
	InfluxBucket       string `env:"INFLUX_CONTRIBUTORS_STATS_BUCKET"`
	Contributors       []struct {
		Name string `env:"name"`
		Url  string `env:"url"`
	}
}

type ContributorsActivityConfiguration struct {
	ContributorsStatsConfiguration
}

// New creates a default configuration with the values from .env file and environment variables.
func New(ctx context.Context) (*Configuration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration Configuration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	return &configuration, nil
}

// New creates a notional configuration with the values from .env file and environment variables.
func NewNotionalConfiguration(ctx context.Context) (*NotionalConfiguration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration NotionalConfiguration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	return &configuration, nil
}

// New creates a transfer report configuration with the values from .env file and environment variables.
func NewTransferReportConfiguration(ctx context.Context) (*TransferReportConfiguration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration TransferReportConfiguration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	return &configuration, nil
}

// New creates a history prices configuration with the values from .env file and environment variables.
func NewHistoricalPricesConfiguration(ctx context.Context) (*HistoricalPricesConfiguration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration HistoricalPricesConfiguration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	return &configuration, nil
}

// New creates a migration source tx configuration with the values from .env file and environment variables.
func NewMigrateSourceTxConfiguration(ctx context.Context) (*MigrateSourceTxConfiguration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration MigrateSourceTxConfiguration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	return &configuration, nil
}
