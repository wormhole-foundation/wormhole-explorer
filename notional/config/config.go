package config

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// Configuration represents the application configuration with the default values.
type Configuration struct {
	LogLevel        string `env:"LOG_LEVEL,default=INFO"`
	Port            string `env:"PORT,default=8000"`
	MongodbURI      string `env:"MONGODB_URI,required"`
	MongodbDatabase string `env:"MONGODB_DATABASE,required"`
	P2pNetwork      string `env:"P2P_NETWORK,required"`
	PprofEnabled    bool   `env:"PPROF_ENABLED,default=false"`
	CacheURL        string `env:"CACHE_URL,required"`
	CachePrefix     string `env:"CACHE_PREFIX,required"`
	CacheChannel    string `env:"CACHE_CHANNEL,required"`
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
