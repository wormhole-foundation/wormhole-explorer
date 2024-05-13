package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// p2p network constants.
const (
	P2pMainNet = "mainnet"
	P2pTestNet = "testnet"
	P2pDevNet  = "devnet"
)

// ServiceConfiguration represents the application configuration when running as service with default values.
type ServiceConfiguration struct {
	// Global configuration
	Environment    string `env:"ENVIRONMENT,required"`
	LogLevel       string `env:"LOG_LEVEL,default=INFO"`
	Port           string `env:"PORT,default=8000"`
	PprofEnabled   bool   `env:"PPROF_ENABLED,default=false"`
	P2pNetwork     string `env:"P2P_NETWORK,required"`
	AlertEnabled   bool   `env:"ALERT_ENABLED,default=false"`
	AlertApiKey    string `env:"ALERT_API_KEY"`
	MetricsEnabled bool   `env:"METRICS_ENABLED,default=false"`
	// Fly event consumer configuration
	ConsumerWorkerSize         int `env:"CONSUMER_WORKER_SIZE,default=1"`
	GovernorConsumerWorkerSize int `env:"GOVERNOR_CONSUMER_WORKER_SIZE,default=1"`

	// Database configuration
	MongoURI      string `env:"MONGODB_URI,required"`
	MongoDatabase string `env:"MONGODB_DATABASE,required"`
	// AWS configuration
	AwsEndpoint        string `env:"AWS_ENDPOINT"`
	AwsAccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	AwsSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	AwsRegion          string `env:"AWS_REGION"`
	DuplicateVaaSQSUrl string `env:"DUPLICATE_VAA_SQS_URL"`
	GovernorSQSUrl     string `env:"GOVERNOR_SQS_URL"`
	// Guardian api provider configuration
	GuardianAPIProviderPath       string `env:"GUARDIAN_API_PROVIDER_PATH,required"`
	*GuardianAPIConfigurationJson `required:"false"`
}

type GuardianAPIConfigurationJson struct {
	GuardianProviders []GuardianProvider `json:"guardian_providers"`
}

type GuardianProvider struct {
	ProviderName      string `json:"name"`
	ProviderUrl       string `json:"url"`
	RequestsPerMinute uint16 `json:"requests_per_minute"`
	Priority          uint8  `json:"priority"`
}

// New creates a configuration with the values from .env file and environment variables.
func New(ctx context.Context) (*ServiceConfiguration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration ServiceConfiguration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	// Load guardian api provider configuration.
	if configuration.GuardianAPIProviderPath != "" {
		guardianAPIJsonFile, err := os.ReadFile(configuration.GuardianAPIProviderPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read guardian API provider settings from file: %w", err)
		}
		var GuardianAPIConfigurationJson GuardianAPIConfigurationJson
		if err := json.Unmarshal(guardianAPIJsonFile, &GuardianAPIConfigurationJson); err != nil {
			return nil, fmt.Errorf("failed to unmarshal guardian API provider settings: %w", err)
		}
		configuration.GuardianAPIConfigurationJson = &GuardianAPIConfigurationJson
	} else {
		return nil, fmt.Errorf("guardian API provider settings file is required")
	}

	return &configuration, nil
}
