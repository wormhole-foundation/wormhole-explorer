package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/config"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/healthcheck"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/sns"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/pipeline"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/watcher"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type exitCode int

func handleExit() {
	if r := recover(); r != nil {
		if e, ok := r.(exitCode); ok {
			os.Exit(int(e))
		}
		panic(r) // not an Exit, bubble up
	}
}

func main() {

	defer handleExit()
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	config, err := config.New(rootCtx)
	if err != nil {
		log.Fatal("Error creating config", err)
	}

	logger := logger.New("wormhole-explorer-pipeline", logger.WithLevel(config.LogLevel))

	logger.Info("Starting wormhole-explorer-pipeline ...")

	//setup DB connection
	db, err := db.New(rootCtx, logger, config.MongoURI, config.MongoDatabase)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	// get metrics.
	metrics := newMetrics(config)

	// get publish function.
	pushFunc, err := newTopicProducer(rootCtx, config, metrics, logger)
	if err != nil {
		logger.Fatal("failed to create publish function", zap.Error(err))
	}

	// get health check functions.
	healthChecks, err := newHealthChecks(rootCtx, config, db.Database)
	if err != nil {
		logger.Fatal("failed to create health checks", zap.Error(err))
	}

	// create a new pipeline repository.
	repository := pipeline.NewRepository(db.Database, logger)

	// create and start a new tx hash handler.
	quit := make(chan bool)
	txHashHandler := pipeline.NewTxHashHandler(repository, pushFunc, metrics, logger, quit)
	go txHashHandler.Run(rootCtx)

	// create a new publisher.
	publisher := pipeline.NewPublisher(pushFunc, metrics, repository, config.P2pNetwork, txHashHandler, logger)
	watcher := watcher.NewWatcher(rootCtx, db.Database, config.MongoDatabase, publisher.Publish, metrics, logger)
	err = watcher.Start(rootCtx)
	if err != nil {
		logger.Fatal("failed to watch MongoDB", zap.Error(err))
	}

	server := infrastructure.NewServer(logger, config.Port, config.PprofEnabled, healthChecks...)
	server.Start()

	logger.Info("Started wormhole-explorer-pipeline")

	// Waiting for signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-rootCtx.Done():
		logger.Warn("Terminating with root context cancelled.")
	case signal := <-sigterm:
		logger.Info("Terminating with signal.", zap.String("signal", signal.String()))
	}

	logger.Info("root context cancelled, exiting...")
	rootCtxCancel()

	logger.Info("Closing tx hash handler ...")
	close(quit)

	logger.Info("Closing database connections ...")
	db.Close()
	logger.Info("Closing Http server ...")
	server.Stop()
	logger.Info("Finished wormhole-explorer-pipeline")

}

func newAwsConfig(appCtx context.Context, cfg *config.Configuration) (aws.Config, error) {
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

		awsCfg, err := awsconfig.LoadDefaultConfig(appCtx,
			awsconfig.WithRegion(region),
			awsconfig.WithEndpointResolver(customResolver),
			awsconfig.WithCredentialsProvider(credentials),
		)
		return awsCfg, err
	}

	return awsconfig.LoadDefaultConfig(appCtx, awsconfig.WithRegion(region))
}

func newTopicProducer(appCtx context.Context, config *config.Configuration, metrics metrics.Metrics, logger *zap.Logger) (topic.PushFunc, error) {
	awsConfig, err := newAwsConfig(appCtx, config)
	if err != nil {
		return nil, err
	}

	snsProducer, err := sns.NewProducer(awsConfig, config.SNSUrl)
	if err != nil {
		return nil, err
	}

	return topic.NewVAASNS(snsProducer, metrics, logger).Publish, nil
}

func newHealthChecks(ctx context.Context, config *config.Configuration, db *mongo.Database) ([]healthcheck.Check, error) {
	awsConfig, err := newAwsConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return []healthcheck.Check{healthcheck.Mongo(db), healthcheck.SNS(awsConfig, config.SNSUrl)}, nil
}

func newMetrics(cfg *config.Configuration) metrics.Metrics {
	metricsEnabled := cfg.MetricsEnabled
	if !metricsEnabled {
		return metrics.NewDummyMetrics()
	}
	return metrics.NewPrometheusMetrics(cfg.P2pNetwork)
}
