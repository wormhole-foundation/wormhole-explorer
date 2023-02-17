package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	AnkrBaseUrl        string `required:"true" split_words:"true"`
	BlockdaemonBaseUrl string `required:"true" split_words:"true"`
	BlockdaemonApiKey  string `required:"true" split_words:"true"`
	SolanaRpcEndpoint  string `required:"true" split_words:"true"`
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
