package config

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// Configuration represents the application configuration with the default values.
type Configuration struct {
	Env           string `env:"ENV,default=development"`
	LogLevel      string `env:"LOG_LEVEL,default=INFO"`
	Port          string `env:"PORT,default=8000"`
	MongoURI      string `env:"MONGODB_URI,required"`
	MongoDatabase string `env:"MONGODB_DATABASE,required"`
	AnkrUrl       string `env:"ANKR_URL,required"`
	PprofEnabled  bool   `env:"PPROF_ENABLED,default=false"`
	P2pNetwork    string `env:"P2P_NETWORK,required"`
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
