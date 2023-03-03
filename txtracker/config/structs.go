package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	// MonitoringPort defines the TCP port for the /health and /ready endpoints.
	MonitoringPort     string `split_words:"true" default:"8000"`
	LogLevel           string `split_words:"true" default:"INFO"`
	PprofEnabled       bool   `split_words:"true" default:"false"`
	AwsEndpoint        string `split_words:"true" required:"true"`
	AwsAccessKeyID     string `split_words:"true" required:"true"`
	AwsSecretAccessKey string `split_words:"true" required:"true"`
	AwsRegion          string `split_words:"true" required:"true"`
	SqsUrl             string `split_words:"true" required:"true"`
	P2pNetwork         string `split_words:"true" required:"true"`
	MongodbUri         string `split_words:"true" required:"true"`
	MongodbDatabase    string `split_words:"true" required:"true"`

	AnkrBaseUrl   string `split_words:"true" required:"true"`
	SolanaBaseUrl string `split_words:"true" required:"true"`
	TerraBaseUrl  string `split_words:"true" required:"true"`

	// The Blockdaemon provider is not being used currently -
	// 	don't need to set it via environment variables
	BlockdaemonBaseUrl string `split_words:"true"`
	BlockdaemonApiKey  string `split_words:"true"`
}

func LoadFromEnv() (*Settings, error) {

	_ = godotenv.Load()

	var s Settings

	err := envconfig.Process("", &s)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from environment: %w", err)
	}

	return &s, nil
}
