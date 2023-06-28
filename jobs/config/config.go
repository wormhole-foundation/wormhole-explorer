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
	Env             string `env:"ENV,default=development"`
	LogLevel        string `env:"LOG_LEVEL,default=INFO"`
	JobID           string `env:"JOB_ID,required"`
	CoingeckoURL    string `env:"COINGECKO_URL,required"`
	CacheURL        string `env:"CACHE_URL,required"`
	CachePrefix     string `env:"CACHE_PREFIX,required"`
	NotionalChannel string `env:"NOTIONAL_CHANNEL,required"`
	P2pNetwork      string `env:"P2P_NETWORK,required"`
}

// New creates a configuration with the values from .env file and environment variables.
func New(ctx context.Context) (*Configuration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration Configuration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	return &configuration, nil
}
