package config

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// supported db layers.
const (
	DbLayerMongo    = "mongo"
	DbLayerPostgres = "postgres"
)

// Configuration represents the application configuration with the default values.
type Configuration struct {
	Environment        string `env:"ENVIRONMENT,required"`
	LogLevel           string `env:"LOG_LEVEL,default=INFO"`
	Port               string `env:"PORT,default=8000"`
	P2pNetwork         string `env:"P2P_NETWORK,required"`
	MongoURI           string `env:"MONGODB_URI,required"`
	MongoDatabase      string `env:"MONGODB_DATABASE,required"`
	AwsEndpoint        string `env:"AWS_ENDPOINT"`
	AwsAccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	AwsSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	AwsRegion          string `env:"AWS_REGION"`
	SNSUrl             string `env:"SNS_URL"`
	PprofEnabled       bool   `env:"PPROF_ENABLED,default=false"`
	AlertEnabled       bool   `env:"ALERT_ENABLED,default=false"`
	AlertApiKey        string `env:"ALERT_API_KEY"`
	MetricsEnabled     bool   `env:"METRICS_ENABLED,default=false"`
	VaaSqsUrl          string `env:"VAA_SQS_URL,default=false"`
	DbLayer            string `env:"DB_LAYER,default=mongo"` // mongo, postgres
	PostreSQLUrl       string `env:"POSTGRESQL_URL"`
	WorkersSize        int    `env:"WORKERS_SIZE,default=10"`
}

type Backfiller struct {
	LogLevel           string
	MongoURI           string
	MongoDatabase      string
	AwsEndpoint        string
	AwsAccessKeyID     string
	AwsSecretAccessKey string
	AwsRegion          string
	SNSUrl             string
	RequestsPerSecond  int64
	StartTime          string
	EndTime            string
	PageSize           int64
	NumWorkers         int
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
