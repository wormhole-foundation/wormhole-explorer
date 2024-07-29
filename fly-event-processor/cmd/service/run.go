package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"

	governorConsumer "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/consumer/governor"
	vaaConsumer "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/consumer/vaa"
	governorConfigProcessor "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor/governor_config"
	governorStatusProcessor "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor/governor_status"
	vaaprocessor "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"

	txTracker "github.com/wormhole-foundation/wormhole-explorer/common/client/txtracker"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/http/vaa"
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

	// initialize db and repositories
	s, err := newStorageLayer(rootCtx, cfg, logger)
	if err != nil {
		logger.Fatal("Error initializing db and repositories: ", zap.Error(err))
	}

	//TxTracker createTxHash client
	createTxHashFunc, err := newCreateTxHashFunc(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize VAA parser", zap.Error(err))
	}

	// create a new processor
	dupVaaProcessor, govStatusProcessor, govConfigProcessor, err := newProcessors(cfg,
		guardianApiProviderPool, s, createTxHashFunc, metrics, logger)
	if err != nil {
		logger.Fatal("failed to initialize processor", zap.Error(err))
	}

	// start serving /health and /ready endpoints
	healthChecks, err := makeHealthChecks(rootCtx, cfg, s.mongoDB.Database, s.postgresDB)
	if err != nil {
		logger.Fatal("Failed to create health checks", zap.Error(err))
	}
	// TODO: handle s.mongoRepository to use postgres also.
	vaaCtrl := vaa.NewController(dupVaaProcessor, s.mongoRepository, logger)
	server := infrastructure.NewServer(logger, cfg.Port, vaaCtrl, cfg.PprofEnabled,
		healthChecks...)
	server.Start()

	// create and start a duplicate VAA consumer.
	duplicateVaaConsumeFunc := newDuplicateVaaConsumeFunc(rootCtx, cfg,
		metrics, logger)
	duplicateVaa := vaaConsumer.New(duplicateVaaConsumeFunc, dupVaaProcessor,
		logger, metrics, cfg.P2pNetwork, cfg.ConsumerWorkerSize)
	duplicateVaa.Start(rootCtx)

	// create and start a governor status consumer.
	governorStatusConsumerFunc := newGovernorStatusConsumeFunc(rootCtx, cfg,
		metrics, logger)

	governorStatus := governorConsumer.New(governorStatusConsumerFunc,
		govStatusProcessor, govConfigProcessor, logger, metrics,
		cfg.P2pNetwork, cfg.GovernorConsumerWorkerSize)
	governorStatus.Start(rootCtx)

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

	// close mongo db connection
	// TODO: remove after switch to use only postgres.
	if s.mongoDB != nil {
		logger.Info("Closing MongoDB connection...")
		s.mongoDB.DisconnectWithTimeout(10 * time.Second)
	}

	// close postgres db connection
	if s.postgresDB != nil {
		logger.Info("Closing Postgres connection...")
		s.postgresDB.Close()
	}

	logger.Info("Terminated wormholescan-fly-event-processor")

}

type storageLayer struct {
	// TODO: remove after switch to use only postgres.
	mongoDB            *dbutil.Session
	mongoRepository    *storage.Repository
	postgresDB         *db.DB
	postgresRepository *storage.PostgresRepository
}

func newStorageLayer(ctx context.Context,
	cfg *config.ServiceConfiguration,
	logger *zap.Logger) (*storageLayer, error) {

	var mongoDb *dbutil.Session
	var mongoRepository *storage.Repository
	var postgresDb *db.DB
	var postgresRepository *storage.PostgresRepository
	var err error
	switch cfg.DbLayer {
	// TODO: remove after switch to use only postgres.
	case config.DbLayerMongo:
		mongoDb, err = dbutil.Connect(ctx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
		if err != nil {
			return nil, err
		}
		mongoRepository = storage.NewRepository(logger, mongoDb.Database)
	case config.DbLayerPostgres:
		postgresDb, err = newPostgresDatabase(ctx, cfg, logger)
		if err != nil {
			return nil, err
		}
		postgresRepository = storage.NewPostgresRepository(postgresDb, logger)
	case config.DbLayerBoth:
		// TODO: remove after switch to use only postgres.
		mongoDb, err = dbutil.Connect(ctx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
		if err != nil {
			return nil, err
		}
		mongoRepository = storage.NewRepository(logger, mongoDb.Database)
		postgresDb, err = newPostgresDatabase(ctx, cfg, logger)
		if err != nil {
			return nil, err
		}
		postgresRepository = storage.NewPostgresRepository(postgresDb, logger)
	default:
		return nil, fmt.Errorf("invalid db layer: %s", cfg.DbLayer)
	}

	return &storageLayer{
		mongoDB:            mongoDb,
		mongoRepository:    mongoRepository,
		postgresDB:         postgresDb,
		postgresRepository: postgresRepository,
	}, nil
}

// newProcessors creates a new processor based on the configuration.
func newProcessors(cfg *config.ServiceConfiguration,
	guardianApiProviderPool *pool.Pool, s *storageLayer, createTxHashFunc txTracker.CreateTxHashFunc,
	metrics metrics.Metrics, logger *zap.Logger) (vaaprocessor.ProcessorFunc, governorStatusProcessor.ProcessorFunc,
	governorConfigProcessor.ProcessorFunc, error) {

	switch cfg.DbLayer {
	case config.DbLayerMongo:
		// TODO: remove after switch to use only postgres.
		dupVaaProcessor := vaaprocessor.NewDuplicateVaaProcessor(guardianApiProviderPool,
			s.mongoRepository, logger, metrics)
		govStatusProcessor := governorStatusProcessor.NewProcessor(s.mongoRepository,
			createTxHashFunc, logger, metrics)
		govConfigProcessor := governorConfigProcessor.NewNoopProcessor()
		return dupVaaProcessor.Process, govStatusProcessor.Process, govConfigProcessor.Process, nil
	case config.DbLayerPostgres:
		dupVaaProcessor := vaaprocessor.NewProcessor(guardianApiProviderPool,
			s.postgresRepository, logger, metrics)
		govStatusProcessor := governorStatusProcessor.NewProcessor(s.postgresRepository,
			createTxHashFunc, logger, metrics)
		govConfigProcessor := governorConfigProcessor.NewProcessor(s.postgresRepository,
			logger, metrics)
		return dupVaaProcessor.Process, govStatusProcessor.Process, govConfigProcessor.Process, nil
	case config.DbLayerBoth:
		// TODO: add vaaProcessor with postgres.
		dupVaaProcessorMongo := vaaprocessor.NewDuplicateVaaProcessor(guardianApiProviderPool,
			s.mongoRepository, logger, metrics)
		dupVaaProcessorPostgres := vaaprocessor.NewProcessor(guardianApiProviderPool,
			s.postgresRepository, logger, metrics)
		dupVaaProcessor := vaaprocessor.NewCompositeProcessor(
			dupVaaProcessorMongo.Process, dupVaaProcessorPostgres.Process)
		govStatusProcessorMongo := governorStatusProcessor.NewProcessor(s.mongoRepository,
			createTxHashFunc, logger, metrics)
		govStatusProcessorPostgres := governorStatusProcessor.NewProcessor(s.postgresRepository,
			createTxHashFunc, logger, metrics)
		govStatusProcessor := governorStatusProcessor.NewCompositeProcessor(
			govStatusProcessorMongo.Process, govStatusProcessorPostgres.Process)
		govConfigProcessor := governorConfigProcessor.NewProcessor(s.postgresRepository,
			logger, metrics)
		return dupVaaProcessor.Process, govStatusProcessor.Process, govConfigProcessor.Process, nil
	}

	return nil, nil, nil, fmt.Errorf("invalid db layer: %s", cfg.DbLayer)
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
	mongoDb *mongo.Database,
	db *db.DB,
) ([]health.Check, error) {

	awsConfig, err := newAwsConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	plugins := []health.Check{health.SQS(awsConfig, cfg.DuplicateVaaSQSUrl)}

	switch cfg.DbLayer {
	case config.DbLayerMongo:
		// TODO: remove after switch to use only postgres.
		plugins = append(plugins, health.Mongo(mongoDb))
	case config.DbLayerPostgres:
		plugins = append(plugins, health.Postgres(db))
	case config.DbLayerBoth:
		plugins = append(plugins, health.Mongo(mongoDb), health.Postgres(db))
	default:
		return nil, fmt.Errorf("invalid db layer: %s", cfg.DbLayer)
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
) queue.ConsumeFunc[queue.EventDuplicateVaa] {

	sqsConsumer, err := newSqsConsumer(ctx, cfg, cfg.DuplicateVaaSQSUrl)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewEventSqs[queue.EventDuplicateVaa](sqsConsumer,
		metrics.IncDuplicatedVaaConsumedQueue, logger)
	return vaaQueue.Consume
}

func newGovernorStatusConsumeFunc(
	ctx context.Context,
	cfg *config.ServiceConfiguration,
	metrics metrics.Metrics,
	logger *zap.Logger,
) queue.ConsumeFunc[queue.EventGovernor] {

	sqsConsumer, err := newSqsConsumer(ctx, cfg, cfg.GovernorSQSUrl)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	governorStatusQueue := queue.NewEventSqs[queue.EventGovernor](sqsConsumer,
		metrics.IncGovernorStatusConsumedQueue, logger)
	return governorStatusQueue.Consume
}

func newCreateTxHashFunc(
	cfg *config.ServiceConfiguration,
	logger *zap.Logger,
) (txTracker.CreateTxHashFunc, error) {
	if cfg.Environment == config.EnvironmentLocal {
		return func(vaaID, txHash string) (*txTracker.TxHashResponse, error) {
			return &txTracker.TxHashResponse{
				NativeTxHash: txHash,
			}, nil
		}, nil
	}
	createTxHashClient, err := txTracker.NewTxTrackerAPIClient(cfg.TxTrackerTimeout, cfg.TxTrackerUrl, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize TxTracker client: %w", err)
	}
	return createTxHashClient.CreateTxHash, nil
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
