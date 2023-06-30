package config

import (
	"context"

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
	Environment             string `env:"ENVIRONMENT,default=development"`
	LogLevel                string `env:"LOG_LEVEL,default=INFO"`
	Port                    string `env:"PORT,default=8000"`
	ConsumerMode            string `env:"CONSUMER_MODE,default=QUEUE"`
	MongoURI                string `env:"MONGODB_URI,required"`
	MongoDatabase           string `env:"MONGODB_DATABASE,required"`
	AwsEndpoint             string `env:"AWS_ENDPOINT"`
	AwsAccessKeyID          string `env:"AWS_ACCESS_KEY_ID"`
	AwsSecretAccessKey      string `env:"AWS_SECRET_ACCESS_KEY"`
	AwsRegion               string `env:"AWS_REGION"`
	SQSUrl                  string `env:"SQS_URL"`
	VaaPayloadParserURL     string `env:"VAA_PAYLOAD_PARSER_URL, required"`
	VaaPayloadParserTimeout int64  `env:"VAA_PAYLOAD_PARSER_TIMEOUT, required"`
	PprofEnabled            bool   `env:"PPROF_ENABLED,default=false"`
	P2pNetwork              string `env:"P2P_NETWORK,required"`
}

// BackfillerConfiguration represents the application configuration when running as backfiller with default values.
type BackfillerConfiguration struct {
	LogLevel                string `env:"LOG_LEVEL,default=INFO"`
	MongoURI                string `env:"MONGODB_URI,required"`
	MongoDatabase           string `env:"MONGODB_DATABASE,required"`
	VaaPayloadParserURL     string `env:"VAA_PAYLOAD_PARSER_URL, required"`
	VaaPayloadParserTimeout int64  `env:"VAA_PAYLOAD_PARSER_TIMEOUT, required"`
	StartTime               string `env:"START_TIME"`
	EndTime                 string `env:"END_TIME"`
	PageSize                int64  `env:"PAGE_SIZE,default=100"`
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

// IsQueueConsumer check if consumer mode is QUEUE.
func (c *ServiceConfiguration) IsQueueConsumer() bool {
	return c.ConsumerMode == "QUEUE"
}
