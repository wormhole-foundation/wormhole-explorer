package service

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/http/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
)

func Run() {
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	// load config
	cfg, err := config.New(rootCtx)
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	// initialize metrics
	metrics := newMetrics(cfg)

	// build logger
	logger := logger.New("wormholescan-fly-event-processor", logger.WithLevel(cfg.LogLevel))
	logger.Info("Starting wormholescan-fly-event-processor ...")

	// create guardian provider pool
	guardianApiProviderPool, err := newGuardianProviderPool(cfg)
	if err != nil {
		logger.Fatal("Error creating guardian provider pool: ", zap.Error(err))
	}

	// initialize the database client
	db, err := dbutil.Connect(rootCtx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
	if err != nil {
		log.Fatal("Failed to initialize MongoDB client: ", err)
	}

	// create a new repository
	repository := storage.NewRepository(logger, db.Database)

	// create a new processor
	processor := processor.NewProcessor(guardianApiProviderPool, repository, logger, metrics)

	// start serving /health and /ready endpoints
	healthChecks, err := makeHealthChecks(rootCtx, cfg, db.Database)
	if err != nil {
		logger.Fatal("Failed to create health checks", zap.Error(err))
	}
	vaaCtrl := vaa.NewController(processor.Process, repository, logger)
	server := infrastructure.NewServer(logger, cfg.Port, vaaCtrl, cfg.PprofEnabled, healthChecks...)
	server.Start()

	// create and start a duplicate VAA consumer.
	duplicateVaaConsumeFunc := newDuplicateVaaConsumeFunc(rootCtx, cfg, metrics, logger)
	duplicateVaa := consumer.New(duplicateVaaConsumeFunc, processor.Process, logger, metrics, cfg.P2pNetwork, cfg.ConsumerWorkerSize)
	duplicateVaa.Start(rootCtx)

	logger.Info("Started wormholescan-fly-event-processor")

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

	logger.Info("Terminated wormholescan-fly-event-processor")

}

func newAwsConfig(ctx context.Context, cfg *config.ServiceConfiguration) (aws.Config, error) {

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

func newSqsConsumer(ctx context.Context, cfg *config.ServiceConfiguration, sqsUrl string) (*sqs.Consumer, error) {

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

func makeHealthChecks(
	ctx context.Context,
	cfg *config.ServiceConfiguration,
	db *mongo.Database,
) ([]health.Check, error) {

	awsConfig, err := newAwsConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	plugins := []health.Check{
		health.SQS(awsConfig, cfg.DuplicateVaaSQSUrl),
		health.Mongo(db),
	}

	return plugins, nil
}

func newMetrics(cfg *config.ServiceConfiguration) metrics.Metrics {
	if !cfg.MetricsEnabled {
		return metrics.NewDummyMetrics()
	}
	return metrics.NewPrometheusMetrics(cfg.Environment)
}

func newGuardianProviderPool(cfg *config.ServiceConfiguration) (*pool.Pool, error) {
	if cfg.GuardianAPIConfigurationJson == nil {
		return nil, errors.New("guardian api provider configuration is missing")
	}

	var guardianCfgs []pool.Config
	for _, provider := range cfg.GuardianAPIConfigurationJson.GuardianProviders {
		guardianCfgs = append(guardianCfgs, pool.Config{
			Id:                provider.ProviderUrl,
			Description:       provider.ProviderName,
			RequestsPerMinute: provider.RequestsPerMinute,
			Priority:          provider.Priority,
		})
	}

	if len(guardianCfgs) == 0 {
		return nil, errors.New("guardian api provider configuration is empty")
	}
	return pool.NewPool(guardianCfgs), nil
}

func newDuplicateVaaConsumeFunc(
	ctx context.Context,
	cfg *config.ServiceConfiguration,
	metrics metrics.Metrics,
	logger *zap.Logger,
) queue.ConsumeFunc {

	sqsConsumer, err := newSqsConsumer(ctx, cfg, cfg.DuplicateVaaSQSUrl)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewEventSqs(sqsConsumer, metrics, logger)
	return vaaQueue.Consume
}
