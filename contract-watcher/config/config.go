package config

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// ServiceConfiguration represents the application configuration when running as service with the default values.
type ServiceConfiguration struct {
	Environment               string `env:"ENVIRONMENT,required"`
	LogLevel                  string `env:"LOG_LEVEL,default=INFO"`
	Port                      string `env:"PORT,default=8000"`
	MongoURI                  string `env:"MONGODB_URI,required"`
	MongoDatabase             string `env:"MONGODB_DATABASE,required"`
	AnkrUrl                   string `env:"ANKR_URL,required"`
	AnkrRequestsPerSecond     int    `env:"ANKR_REQUESTS_PER_SECOND,required"`
	SolanaUrl                 string `env:"SOLANA_URL,required"`
	SolanaRequestsPerSecond   int    `env:"SOLANA_REQUESTS_PER_SECOND,required"`
	TerraUrl                  string `env:"TERRA_URL,required"`
	TerraRequestsPerSecond    int    `env:"TERRA_REQUESTS_PER_SECOND,required"`
	AptosUrl                  string `env:"APTOS_URL,required"`
	AptosRequestsPerSecond    int    `env:"APTOS_REQUESTS_PER_SECOND,required"`
	OasisUrl                  string `env:"OASIS_URL,required"`
	OasisRequestsPerSecond    int    `env:"OASIS_REQUESTS_PER_SECOND,required"`
	MoonbeamUrl               string `env:"MOONBEAM_URL,required"`
	MoonbeamRequestsPerSecond int    `env:"MOONBEAM_REQUESTS_PER_SECOND,required"`
	CeloUrl                   string `env:"CELO_URL,required"`
	CeloRequestsPerSecond     int    `env:"CELO_REQUESTS_PER_SECOND,required"`
	PprofEnabled              bool   `env:"PPROF_ENABLED,default=false"`
	P2pNetwork                string `env:"P2P_NETWORK,required"`
	AlertEnabled              bool   `env:"ALERT_ENABLED,required"`
	AlertApiKey               string `env:"ALERT_API_KEY"`
}

// BackfillerConfiguration represents the application configuration when running as backfiller.
type BackfillerConfiguration struct {
	LogLevel           string `env:"LOG_LEVEL,default=INFO"`
	MongoURI           string `env:"MONGODB_URI,required"`
	MongoDatabase      string `env:"MONGODB_DATABASE,required"`
	ChainName          string `env:"CHAIN_NAME,required"`
	ChainUrl           string `env:"CHAIN_URL,required"`
	FromBlock          uint64 `env:"FROM_BLOCK,required"`
	ToBlock            uint64 `env:"TO_BLOCK,required"`
	Network            string `env:"NETWORK,required"`
	RateLimitPerSecond int    `env:"RATE_LIMIT_PER_SECOND,default=10"`
	PageSize           uint64 `env:"PAGE_SIZE,default=100"`
	PersistBlock       bool   `env:"PERSIST_BLOCK,default=false"`
}

// New creates a configuration with the values from .env file and environment variables.
func New(ctx context.Context) (*ServiceConfiguration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration ServiceConfiguration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	return &configuration, nil
}
