package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type BackfillerSettings struct {
	LogLevel   string `split_words:"true" default:"INFO"`
	NumWorkers uint   `split_words:"true" required:"true"`
	BulkSize   uint   `split_words:"true" required:"true"`

	VaaPayloadParserSettings
	MongodbSettings
	RpcProviderSettings
}

type ServiceSettings struct {
	// MonitoringPort defines the TCP port for the /health and /ready endpoints.
	MonitoringPort string `split_words:"true" default:"8000"`
	LogLevel       string `split_words:"true" default:"INFO"`
	PprofEnabled   bool   `split_words:"true" default:"false"`

	AwsSettings
	VaaPayloadParserSettings
	MongodbSettings
	RpcProviderSettings
}

type VaaPayloadParserSettings struct {
	VaaPayloadParserUrl     string `split_words:"true" required:"true"`
	VaaPayloadParserTimeout int64  `split_words:"true" required:"true"`
}

type AwsSettings struct {
	AwsEndpoint        string `split_words:"true" required:"false"`
	AwsAccessKeyID     string `split_words:"true" required:"true"`
	AwsSecretAccessKey string `split_words:"true" required:"true"`
	AwsRegion          string `split_words:"true" required:"true"`
	SqsUrl             string `split_words:"true" required:"true"`
}

type MongodbSettings struct {
	MongodbUri      string `split_words:"true" required:"true"`
	MongodbDatabase string `split_words:"true" required:"true"`
}

type RpcProviderSettings struct {
	AnkrBaseUrl           string `split_words:"true" required:"true"`
	AnkrApiKey            string `split_words:"true" required:"false"`
	AnkrRequestsPerMinute uint16 `split_words:"true" required:"true"`

	SolanaBaseUrl           string `split_words:"true" required:"true"`
	SolanaRequestsPerMinute uint16 `split_words:"true" required:"true"`

	TerraBaseUrl           string `split_words:"true" required:"true"`
	TerraRequestsPerMinute uint16 `split_words:"true" required:"true"`
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
