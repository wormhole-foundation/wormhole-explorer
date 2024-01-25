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
	"github.com/wormhole-foundation/wormhole-explorer/common/configuration"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/http/vaa"
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

	var testRpcConfig *config.TestnetRpcProviderSettings
	if configuration.IsTestnet(cfg.P2pNetwork) {
		testRpcConfig, err = config.LoadFromEnv[config.TestnetRpcProviderSettings]()
		if err != nil {
			log.Fatal("Error loading testnet rpc config: ", err)
		}
	}

	// initialize metrics
	metrics := newMetrics(cfg)

	// build logger
	logger := logger.New("wormhole-explorer-tx-tracker", logger.WithLevel(cfg.LogLevel))

	logger.Info("Starting wormhole-explorer-tx-tracker ...")

	// initialize rate limiters
	chains.Initialize(&cfg.RpcProviderSettings, testRpcConfig)

	// initialize the database client
	db, err := dbutil.Connect(rootCtx, logger, cfg.MongodbUri, cfg.MongodbDatabase, false)
	if err != nil {
		log.Fatal("Failed to initialize MongoDB client: ", err)
	}

	// create repositories
	repository := consumer.NewRepository(logger, db.Database)
	vaaRepository := vaa.NewRepository(db.Database, logger)

	// create controller
	vaaController := vaa.NewController(vaaRepository, repository, &cfg.RpcProviderSettings, cfg.P2pNetwork, logger)

	// start serving /health and /ready endpoints
	healthChecks, err := makeHealthChecks(rootCtx, cfg, db.Database)
	if err != nil {
		logger.Fatal("Failed to create health checks", zap.Error(err))
	}
	server := infrastructure.NewServer(logger, cfg.MonitoringPort, cfg.PprofEnabled, vaaController, healthChecks...)
	server.Start()

	// create and start a pipeline consumer.
	vaaConsumeFunc := newVAAConsumeFunc(rootCtx, cfg, metrics, logger)
	vaaConsumer := consumer.New(vaaConsumeFunc, &cfg.RpcProviderSettings, rootCtx, logger, repository, metrics, cfg.P2pNetwork)
	vaaConsumer.Start(rootCtx)

	// create and start a notification consumer.
	notificationConsumeFunc := newNotificationConsumeFunc(rootCtx, cfg, metrics, logger)
	notificationConsumer := consumer.New(notificationConsumeFunc, &cfg.RpcProviderSettings, rootCtx, logger, repository, metrics, cfg.P2pNetwork)
	notificationConsumer.Start(rootCtx)

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
) queue.ConsumeFunc {

	sqsConsumer, err := newSqsConsumer(ctx, cfg, cfg.PipelineSqsUrl)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewEventSqs(sqsConsumer, queue.NewVaaConverter(logger), metrics, logger)
	return vaaQueue.Consume
}

func newNotificationConsumeFunc(
	ctx context.Context,
	cfg *config.ServiceSettings,
	metrics metrics.Metrics,
	logger *zap.Logger,
) queue.ConsumeFunc {

	sqsConsumer, err := newSqsConsumer(ctx, cfg, cfg.NotificationsSqsUrl)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewEventSqs(sqsConsumer, queue.NewNotificationEvent(logger), metrics, logger)
	return vaaQueue.Consume
}

func newSqsConsumer(ctx context.Context, cfg *config.ServiceSettings, sqsUrl string) (*sqs.Consumer, error) {

	awsconfig, err := newAwsConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	consumer, err := sqs.NewConsumer(
		awsconfig,
		sqsUrl,
		sqs.WithMaxMessages(10),
		sqs.WithVisibilityTimeout(60),
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
		health.SQS(awsConfig, config.PipelineSqsUrl),
		health.SQS(awsConfig, config.NotificationsSqsUrl),
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
