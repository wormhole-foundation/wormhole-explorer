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

// Configuration represents the application configuration with the default values.
type Configuration struct {
	Env                     string `env:"ENV,default=development"`
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
	InfluxUrl               string `env:"INFLUX_URL,required"`
	InfluxToken             string `env:"INFLUX_TOKEN,required"`
	InfluxOrg               string `env:"INFLUX_ORG,required"`
	InfluxBucket            string `env:"INFLUX_BUCKET,required"`
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

// IsQueueConsumer check if consumer mode is QUEUE.
func (c *Configuration) IsQueueConsumer() bool {
	return c.ConsumerMode == "QUEUE"
}
