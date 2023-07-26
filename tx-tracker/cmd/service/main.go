package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbhelpers"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	"go.uber.org/zap"
)

func main() {
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	// load config
	cfg, err := config.LoadFromEnv[config.ServiceSettings]()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	// initialize metrics
	metrics := newMetrics(cfg)

	// build logger
	logger := logger.New("wormhole-explorer-tx-tracker", logger.WithLevel(cfg.LogLevel))

	logger.Info("Starting wormhole-explorer-tx-tracker ...")

	// initialize rate limiters
	chains.Initialize(&cfg.RpcProviderSettings)

	// initialize the database client
	db, err := dbhelpers.Connect(rootCtx, cfg.MongodbUri, cfg.MongodbDatabase)
	if err != nil {
		log.Fatal("Failed to initialize MongoDB client: ", err)
	}

	// start serving /health and /ready endpoints
	healthChecks, err := makeHealthChecks(rootCtx, cfg, db.Database)
	if err != nil {
		logger.Fatal("Failed to create health checks", zap.Error(err))
	}
	server := infrastructure.NewServer(logger, cfg.MonitoringPort, cfg.PprofEnabled, healthChecks...)
	server.Start()

	// create and start a consumer.
	vaaConsumeFunc := newVAAConsumeFunc(rootCtx, cfg, metrics, logger)
	repository := consumer.NewRepository(logger, db.Database)
	consumer := consumer.New(vaaConsumeFunc, &cfg.RpcProviderSettings, rootCtx, logger, repository, metrics)
	consumer.Start(rootCtx)

	logger.Info("Started wormhole-explorer-tx-tracker")

	// Waiting for signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-rootCtx.Done():
		logger.Warn("Terminating with root context cancelled.")
	case signal := <-sigterm:
		logger.Info("Terminating with signal.", zap.String("signal", signal.String()))
	}

	// graceful shutdown
	logger.Info("Cancelling root context...")
	rootCtxCancel()

	logger.Info("Closing Http server...")
	server.Stop()

	logger.Info("Closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	logger.Info("Terminated wormhole-explorer-tx-tracker")
}

func newVAAConsumeFunc(
	ctx context.Context,
	cfg *config.ServiceSettings,
	metrics metrics.Metrics,
	logger *zap.Logger,
) queue.VAAConsumeFunc {

	sqsConsumer, err := newSqsConsumer(ctx, cfg)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewVaaSqs(sqsConsumer, metrics, logger)
	return vaaQueue.Consume
}

func newSqsConsumer(ctx context.Context, cfg *config.ServiceSettings) (*sqs.Consumer, error) {

	awsconfig, err := newAwsConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	consumer, err := sqs.NewConsumer(
		awsconfig,
		cfg.SqsUrl,
		sqs.WithMaxMessages(10),
		sqs.WithVisibilityTimeout(4*60),
	)
	return consumer, err
}

func newAwsConfig(ctx context.Context, cfg *config.ServiceSettings) (aws.Config, error) {

	region := cfg.AwsRegion

	if cfg.AwsAccessKeyID != "" && cfg.AwsSecretAccessKey != "" {

		credentials := credentials.NewStaticCredentialsProvider(cfg.AwsAccessKeyID, cfg.AwsSecretAccessKey, "")

		customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if cfg.AwsEndpoint != "" {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           cfg.AwsEndpoint,
					SigningRegion: region,
				}, nil
			}

			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		awsCfg, err := awsconfig.LoadDefaultConfig(
			ctx,
			awsconfig.WithRegion(region),
			awsconfig.WithEndpointResolver(customResolver),
			awsconfig.WithCredentialsProvider(credentials),
		)
		return awsCfg, err
	}
	return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
}

func makeHealthChecks(
	ctx context.Context,
	config *config.ServiceSettings,
	db *mongo.Database,
) ([]health.Check, error) {

	awsConfig, err := newAwsConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	plugins := []health.Check{
		health.SQS(awsConfig, config.SqsUrl),
		health.Mongo(db),
	}

	return plugins, nil
}

func newMetrics(cfg *config.ServiceSettings) metrics.Metrics {
	if !cfg.MetricsEnabled {
		return metrics.NewDummyMetrics()
	}
	return metrics.NewPrometheusMetrics(cfg.Environment)
}
