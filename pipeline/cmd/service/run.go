package service

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/consumer"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	db2 "github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/config"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/healthcheck"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/http/infrastructure"
	pipelineAlert "github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/queue"
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

func Run() {

	defer handleExit()
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	cfg, err := config.New(rootCtx)
	if err != nil {
		log.Fatal("Error creating config", err)
	}

	logger := logger.New("wormhole-explorer-pipeline", logger.WithLevel(cfg.LogLevel))

	logger.Info("Starting wormhole-explorer-pipeline ...")

	// get alert client.
	alertClient, err := newAlertClient(cfg)
	if err != nil {
		logger.Fatal("failed to create alert client", zap.Error(err))
	}

	// get metrics.
	metrics := newMetrics(cfg)

	awsCfg, err := newAwsConfig(rootCtx, cfg)
	if err != nil {
		logger.Fatal("failed to create aws config", zap.Error(err))
	}
	//setup DB connection
	db, err := dbutil.Connect(rootCtx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	// get publish function.
	pushFunc, err := newTopicProducer(rootCtx, cfg, alertClient, metrics, logger, awsCfg)
	if err != nil {
		logger.Fatal("failed to create publish function", zap.Error(err))
	}

	quit := make(chan bool)
	var postresqlClient *db2.DB

	if cfg.DbLayer == config.DbLayerPostgres {

		postresqlClient, err = db2.NewDB(rootCtx, cfg.PostreSQLUrl)
		if err != nil {
			log.Fatal("Failed to initialize PostgreSQL client: ", err)
		}
		defer postresqlClient.Close()

		postresqlRepository := consumer.NewPostreSqlRepository(postresqlClient)

		consumeFunc := newVaaSqsConsumeFunc(rootCtx, awsCfg, cfg, metrics, logger)
		consumer.New(postresqlRepository,
			logger,
			pushFunc,
			metrics,
			cfg.WorkersSize,
		).Start(rootCtx, consumeFunc)

	} else {
		// create a new pipeline repository.
		repository := pipeline.NewRepository(db.Database, logger)

		// create and start a new tx hash handler.
		txHashHandler := pipeline.NewTxHashHandler(repository, pushFunc, alertClient, metrics, logger, quit)
		go txHashHandler.Run(rootCtx)

		// create a new publisher.
		publisher := pipeline.NewPublisher(pushFunc, metrics, repository, cfg.P2pNetwork, txHashHandler, logger)
		watcher := watcher.NewWatcher(rootCtx, db.Database, cfg.MongoDatabase, publisher.Publish, alertClient, metrics, logger)
		err = watcher.Start(rootCtx)
		if err != nil {
			logger.Fatal("failed to watch MongoDB", zap.Error(err))
		}
	}

	// get health check functions.
	healthChecks := newHealthChecks(cfg, db.Database, awsCfg, postresqlClient)

	server := infrastructure.NewServer(logger, cfg.Port, cfg.PprofEnabled, healthChecks...)
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

	logger.Info("closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

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

func newTopicProducer(appCtx context.Context, config *config.Configuration, alertClient alert.AlertClient, metrics metrics.Metrics, logger *zap.Logger, awsConfig aws.Config) (topic.PushFunc, error) {

	snsProducer, err := sns.NewProducer(awsConfig, config.SNSUrl)
	if err != nil {
		return nil, err
	}

	return topic.NewVAASNS(snsProducer, alertClient, metrics, logger).Publish, nil
}

func newHealthChecks(cfg *config.Configuration, db *mongo.Database, awsConfig aws.Config, sqlClient *db2.DB) []healthcheck.Check {
	checks := []healthcheck.Check{healthcheck.SNS(awsConfig, cfg.SNSUrl)}
	if cfg.DbLayer == config.DbLayerMongo {
		checks = append(checks, healthcheck.Mongo(db))
	}
	if cfg.DbLayer == config.DbLayerPostgres {
		checks = append(checks, healthcheck.Postresql(sqlClient))
	}
	return checks
}

func newMetrics(cfg *config.Configuration) metrics.Metrics {
	metricsEnabled := cfg.MetricsEnabled
	if !metricsEnabled {
		return metrics.NewDummyMetrics()
	}
	return metrics.NewPrometheusMetrics(cfg.Environment)
}

func newAlertClient(cfg *config.Configuration) (alert.AlertClient, error) {
	if !cfg.AlertEnabled {
		return alert.NewDummyClient(), nil
	}

	alertConfig := alert.AlertConfig{
		Environment: cfg.Environment,
		ApiKey:      cfg.AlertApiKey,
		Enabled:     cfg.AlertEnabled,
	}
	return alert.NewAlertService(alertConfig, pipelineAlert.LoadAlerts)
}

func newVaaSqsConsumeFunc(
	ctx context.Context,
	awsCfg aws.Config,
	cfg *config.Configuration,
	metrics metrics.Metrics,
	logger *zap.Logger,
) queue.ConsumeFunc {

	sqsConsumer, err := newSqsConsumer(awsCfg, cfg.VaaSqsUrl)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewEventSqs(sqsConsumer, queue.NewVaaConverter(logger), metrics, logger)
	return vaaQueue.Consume
}

func newSqsConsumer(awsconfig aws.Config, sqsUrl string) (*sqs.Consumer, error) {

	consumer, err := sqs.NewConsumer(
		awsconfig,
		sqsUrl,
		sqs.WithMaxMessages(10),
		sqs.WithVisibilityTimeout(60),
	)
	return consumer, err
}
