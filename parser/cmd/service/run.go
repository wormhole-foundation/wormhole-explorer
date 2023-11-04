package service

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
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	vaaPayloadParser "github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/parser/config"
	"github.com/wormhole-foundation/wormhole-explorer/parser/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/parser/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/parser/http/vaa"
	parserAlert "github.com/wormhole-foundation/wormhole-explorer/parser/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/parser/migration"
	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole-explorer/parser/processor"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
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

func Run() {

	defer handleExit()
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	config, err := config.New(rootCtx)
	if err != nil {
		log.Fatal("Error creating config", err)
	}

	logger := logger.New("wormhole-explorer-parser", logger.WithLevel(config.LogLevel))

	logger.Info("Starting wormhole-explorer-parser ...")

	// setup DB connection
	db, err := dbutil.Connect(rootCtx, logger, config.MongoURI, config.MongoDatabase, false)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	// run the database migration.
	err = migration.Run(db.Database)
	if err != nil {
		logger.Fatal("error running migration", zap.Error(err))
	}

	// get alert client.
	alertClient, err := newAlertClient(config)
	if err != nil {
		logger.Fatal("failed to create alert client", zap.Error(err))
	}

	// create a metrics
	metrics := newMetrics(config)

	// create a parserVAAAPIClient
	parserVAAAPIClient, err := vaaPayloadParser.NewParserVAAAPIClient(config.VaaPayloadParserTimeout,
		config.VaaPayloadParserURL, logger)
	if err != nil {
		logger.Fatal("failed to create parse vaa api client")
	}

	// get vaa consumer function.
	vaaConsumeFunc := newVAAConsume(rootCtx, config, metrics, logger)

	//get notification consumer function.
	notificationConsumeFunc := newNotificationConsume(rootCtx, config, metrics, logger)

	// create a repository
	repository := parser.NewRepository(db.Database, logger)

	// get health check functions.
	logger.Info("creating health check functions...")
	healthChecks, err := newHealthChecks(rootCtx, config, db.Database)
	if err != nil {
		logger.Fatal("failed to create health checks", zap.Error(err))
	}

	//create a processor
	processor := processor.New(parserVAAAPIClient, repository, alertClient, metrics, logger)

	// create and start a vaaConsumer
	vaaConsumer := consumer.New(vaaConsumeFunc, processor.Process, metrics, logger)
	vaaConsumer.Start(rootCtx)

	// create and start a notificationConsumer
	notificationConsumer := consumer.New(notificationConsumeFunc, processor.Process, metrics, logger)
	notificationConsumer.Start(rootCtx)

	vaaRepository := vaa.NewRepository(db.Database, logger)
	vaaController := vaa.NewController(vaaRepository, processor.Process, logger)
	server := infrastructure.NewServer(logger, config.Port, config.PprofEnabled, vaaController, healthChecks...)
	server.Start()

	logger.Info("Started wormhole-explorer-parser")

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

	logger.Info("closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	logger.Info("Closing Http server ...")
	server.Stop()

	logger.Info("Finished wormhole-explorer-parser")
}

// Creates a new AWS config depending on whether the execution is local (localstack) or not (AWS)
func newAwsConfig(appCtx context.Context, cfg *config.ServiceConfiguration) (aws.Config, error) {
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

func newVAAConsume(appCtx context.Context, config *config.ServiceConfiguration, metrics metrics.Metrics, logger *zap.Logger) queue.ConsumeFunc {
	sqsConsumer, err := newSQSConsumer(appCtx, config, config.SQSUrl)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	filterConsumeFunc := newFilterFunc(config)
	vaaQueue := queue.NewEventSQS(sqsConsumer, queue.NewVaaConverter(logger), filterConsumeFunc, metrics, logger)
	return vaaQueue.Consume
}

func newNotificationConsume(appCtx context.Context, config *config.ServiceConfiguration, metrics metrics.Metrics, logger *zap.Logger) queue.ConsumeFunc {
	sqsConsumer, err := newSQSConsumer(appCtx, config, config.NotificationsSQSUrl)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	filterConsumeFunc := newFilterFunc(config)
	vaaQueue := queue.NewEventSQS(sqsConsumer, queue.NewNotificationEvent(logger), filterConsumeFunc, metrics, logger)
	return vaaQueue.Consume
}

// Create a new SQS consumer.
func newSQSConsumer(appCtx context.Context, config *config.ServiceConfiguration, sqsUrl string) (*sqs.Consumer, error) {
	awsconfig, err := newAwsConfig(appCtx, config)
	if err != nil {
		return nil, err
	}

	return sqs.NewConsumer(awsconfig, sqsUrl,
		sqs.WithMaxMessages(10),
		sqs.WithVisibilityTimeout(120))
}

// Creates a filter depending on whether the execution is local (dummy filter) or not (Pyth filter)
func newFilterFunc(cfg *config.ServiceConfiguration) queue.FilterConsumeFunc {
	if cfg.P2pNetwork == config.P2pMainNet {
		return queue.PythFilter
	}
	return queue.NonFilter
}

// Creates a metrics depending on whether the execution is local (dummy metrics) or not (Prometheus metrics)
func newMetrics(cfg *config.ServiceConfiguration) metrics.Metrics {
	if !cfg.MetricsEnabled {
		return metrics.NewDummyMetrics()
	}
	return metrics.NewPrometheusMetrics(cfg.Environment)
}

func newAlertClient(cfg *config.ServiceConfiguration) (alert.AlertClient, error) {
	if !cfg.AlertEnabled {
		return alert.NewDummyClient(), nil
	}

	alertConfig := alert.AlertConfig{
		Environment: cfg.Environment,
		ApiKey:      cfg.AlertApiKey,
		Enabled:     cfg.AlertEnabled,
	}

	return alert.NewAlertService(alertConfig, parserAlert.LoadAlerts)
}

func newHealthChecks(
	ctx context.Context,
	config *config.ServiceConfiguration,
	db *mongo.Database,
) ([]health.Check, error) {

	awsConfig, err := newAwsConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	healthChecks := []health.Check{
		health.SQS(awsConfig, config.SQSUrl),
		health.SQS(awsConfig, config.NotificationsSQSUrl),
		health.Mongo(db),
	}
	return healthChecks, nil
}
