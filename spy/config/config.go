package config

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// Configuration represents the application configuration with the default values.
type Configuration struct {
	Env          string `env:"ENV,default=development"`
	LogLevel     string `env:"LOG_LEVEL,default=INFO"`
	Port         string `env:"PORT,default=8000"`
	GrpcAddress  string `env:"GRPC_ADDRESS,default=0.0.0.0:6789"`
	RedisURI     string `env:"REDIS_URI,required"`
	RedisPrefix  string `env:"REDIS_PREFIX,required"`
	RedisChannel string `env:"REDIS_VAA_CHANNEL,required"`
	PprofEnabled bool   `env:"PPROF_ENABLED,default=false"`
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
