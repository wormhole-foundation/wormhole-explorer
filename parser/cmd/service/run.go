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
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	vaaPayloadParser "github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
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

	storage, err := newStorageLayer(rootCtx, config, logger)
	if err != nil {
		logger.Fatal("failed to create storage layer", zap.Error(err))
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

	// get health check functions.
	logger.Info("creating health check functions...")
	healthChecks, err := newHealthChecks(rootCtx, config, storage.mongoDB.Database, storage.postgresDB)
	if err != nil {
		logger.Fatal("failed to create health checks", zap.Error(err))
	}
	// create a token provider
	tokenProvider := domain.NewTokenProvider(config.P2pNetwork)

	//create a processor
	processor := processor.New(parserVAAAPIClient, config.DbLayer, storage.mongoRepository, storage.postgresRepository,
		alertClient, metrics, tokenProvider, logger)

	// create and start a vaaConsumer
	vaaConsumer := consumer.New(vaaConsumeFunc, processor.Process, metrics, logger)
	vaaConsumer.Start(rootCtx)

	// create and start a notificationConsumer
	notificationConsumer := consumer.New(notificationConsumeFunc, processor.Process, metrics, logger)
	notificationConsumer.Start(rootCtx)

	vaaRepository := vaa.NewRepository(storage.mongoDB.Database, logger)
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
	storage.mongoDB.DisconnectWithTimeout(10 * time.Second)

	logger.Info("Closing Http server ...")
	server.Stop()

	logger.Info("Finished wormhole-explorer-parser")
}

type StorageLayer struct {
	mongoDB            *dbutil.Session
	mongoRepository    *parser.Repository
	postgresDB         *db.DB
	postgresRepository *parser.PostgresRepository
}

// newStorageLayer creates a new storage layer.
func newStorageLayer(ctx context.Context, cfg *config.ServiceConfiguration, logger *zap.Logger) (*StorageLayer, error) {

	var mongoDB *dbutil.Session
	var mongoRepository *parser.Repository
	var postgresDB *db.DB
	var postgresRepository *parser.PostgresRepository
	var err error
	switch cfg.DbLayer {
	case config.DbLayerMongo:
		// setup mongo db connection
		mongoDB, err = dbutil.Connect(ctx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
		if err != nil {
			logger.Error("failed to connect MongoDB", zap.Error(err))
			return nil, err
		}

		// run the mongo database migration.
		err = migration.Run(mongoDB.Database)
		if err != nil {
			logger.Error("error running migration", zap.Error(err))
			return nil, err
		}
		// create a mongo repository
		mongoRepository = parser.NewRepository(mongoDB.Database, logger)
		return &StorageLayer{
			mongoDB:         mongoDB,
			mongoRepository: mongoRepository,
		}, nil
	case config.DbLayerPostgres:
		// setup postgres db connection
		postgresDB, err = newPostgresDatabase(ctx, cfg, logger)
		if err != nil {
			logger.Error("failed to connect Postgres", zap.Error(err))
			return nil, err
		}

		// create a postgres repository
		postgresRepository = parser.NewPostgresRepository(postgresDB, logger)
		return &StorageLayer{
			postgresDB:         postgresDB,
			postgresRepository: postgresRepository,
		}, nil
	case config.DbLayerBoth:
		// setup mongo db connection
		mongoDB, err = dbutil.Connect(ctx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
		if err != nil {
			logger.Error("failed to connect MongoDB", zap.Error(err))
			return nil, err
		}

		// run the mongo database migration.
		err = migration.Run(mongoDB.Database)
		if err != nil {
			logger.Error("error running migration", zap.Error(err))
			return nil, err
		}
		// create a mongo repository
		mongoRepository = parser.NewRepository(mongoDB.Database, logger)

		// setup postgres db connection
		postgresDB, err = newPostgresDatabase(ctx, cfg, logger)
		if err != nil {
			logger.Error("failed to connect Postgres", zap.Error(err))
			return nil, err
		}

		// create a postgres repository
		postgresRepository = parser.NewPostgresRepository(postgresDB, logger)
		return &StorageLayer{
			mongoDB:            mongoDB,
			mongoRepository:    mongoRepository,
			postgresDB:         postgresDB,
			postgresRepository: postgresRepository,
		}, nil

	default:
		return nil, errors.New("invalid db layer")
	}
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
	sqsConsumer, err := newSQSConsumer(appCtx, config, config.PipelineSQSUrl)
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
	cfg *config.ServiceConfiguration,
	mongoDB *mongo.Database,
	postgresDB *db.DB,
) ([]health.Check, error) {

	awsConfig, err := newAwsConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	healthChecks := []health.Check{
		health.SQS(awsConfig, cfg.PipelineSQSUrl),
		health.SQS(awsConfig, cfg.NotificationsSQSUrl),
	}

	switch cfg.DbLayer {
	case config.DbLayerMongo:
		healthChecks = append(healthChecks, health.Mongo(mongoDB))
	case config.DbLayerPostgres:
		healthChecks = append(healthChecks, health.Postgres(postgresDB))
	case config.DbLayerBoth:
		healthChecks = append(healthChecks, health.Mongo(mongoDB))
		healthChecks = append(healthChecks, health.Postgres(postgresDB))
	default:
		return nil, errors.New("invalid db layer")
	}

	return healthChecks, nil
}

func newPostgresDatabase(ctx context.Context,
	cfg *config.ServiceConfiguration,
	logger *zap.Logger) (*db.DB, error) {

	// Enable database logging
	var options db.Option
	if cfg.DbLogEnable {
		options = db.WithTracer(logger)
	}

	return db.NewDB(ctx, cfg.DbURL, options)
}
