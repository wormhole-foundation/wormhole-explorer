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
	PricesPath    string `env:"PRICES_PATH,required"`
	OutputPath    string `env:"OUTPUT_PATH,required"`
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
