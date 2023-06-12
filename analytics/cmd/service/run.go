package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/go-redis/redis/v8"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/config"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/metric"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/queue"
	wormscanNotionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	sqs_client "github.com/wormhole-foundation/wormhole-explorer/common/client/sqs"
	health "github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// load configuration
	config, err := config.New(rootCtx)
	if err != nil {
		log.Fatal("Error creating config", err)
	}

	// build logger
	logger := logger.New("wormhole-explorer-analytics", logger.WithLevel(config.LogLevel))
	logger.Info("starting analytics service...")

	// setup DB connection
	logger.Info("connecting to MongoDB...")
	db, err := NewDatabase(rootCtx, logger, config.MongodbURI, config.MongodbDatabase)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	// create influxdb client.
	logger.Info("initializing InfluxDB client...")
	influxCli := newInfluxClient(config.InfluxUrl, config.InfluxToken)
	influxCli.Options().SetBatchSize(100)

	// get health check functions.
	logger.Info("creating health check functions...")
	healthChecks, err := newHealthChecks(rootCtx, config, influxCli, db.Database)
	if err != nil {
		logger.Fatal("failed to create health checks", zap.Error(err))
	}

	//create notional cache
	logger.Info("initializing notional cache...")
	notionalCache, err := newNotionalCache(rootCtx, config, logger)
	if err != nil {
		logger.Fatal("failed to create notional cache", zap.Error(err))
	}

	// create a metrics instance
	logger.Info("initializing metrics instance...")
	metric, err := metric.New(rootCtx, db.Database, influxCli, config.InfluxOrganization, config.InfluxBucketInfinite,
		config.InfluxBucket30Days, config.InfluxBucket24Hours, notionalCache, logger)
	if err != nil {
		logger.Fatal("failed to create metrics instance", zap.Error(err))
	}

	// create and start a consumer.
	logger.Info("initializing metrics consumer...")
	vaaConsumeFunc := newVAAConsume(rootCtx, config, logger)
	consumer := consumer.New(vaaConsumeFunc, metric.Push, logger, config.P2pNetwork)
	consumer.Start(rootCtx)

	// create and start server.
	logger.Info("initializing infrastructure server...")
	server := infrastructure.NewServer(logger, config.Port, config.PprofEnabled, healthChecks...)
	server.Start()

	// Waiting for signal
	logger.Info("waiting for termination signal or context cancellation...")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-rootCtx.Done():
		logger.Warn("terminating (root context cancelled)")
	case signal := <-sigterm:
		logger.Info("terminating (signal received)", zap.String("signal", signal.String()))
	}

	logger.Info("cancelling root context...")
	rootCtxCancel()
	logger.Info("closing metrics client...")
	metric.Close()
	logger.Info("closing HTTP server...")
	server.Stop()
	logger.Info("closing MongoDB connection...")
	db.Close()
	logger.Info("terminated successfully")
}

// Creates a callbacks depending on whether the execution is local (memory queue) or not (SQS queue)
func newVAAConsume(appCtx context.Context, config *config.Configuration, logger *zap.Logger) queue.VAAConsumeFunc {
	sqsConsumer, err := newSQSConsumer(appCtx, config)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewVaaSqs(sqsConsumer, logger)
	return vaaQueue.Consume
}

func newSQSConsumer(appCtx context.Context, config *config.Configuration) (*sqs_client.Consumer, error) {
	awsconfig, err := newAwsConfig(appCtx, config)
	if err != nil {
		return nil, err
	}

	return sqs_client.NewConsumer(awsconfig, config.SQSUrl,
		sqs_client.WithMaxMessages(10),
		sqs_client.WithVisibilityTimeout(120))
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

func newInfluxClient(url, token string) influxdb2.Client {
	return influxdb2.NewClient(url, token)
}

func newHealthChecks(
	ctx context.Context,
	config *config.Configuration,
	influxCli influxdb2.Client,
	db *mongo.Database,
) ([]health.Check, error) {

	awsConfig, err := newAwsConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	healthChecks := []health.Check{
		health.SQS(awsConfig, config.SQSUrl),
		health.Influx(influxCli),
		health.Mongo(db),
	}
	return healthChecks, nil
}

func newNotionalCache(
	ctx context.Context,
	cfg *config.Configuration,
	logger *zap.Logger,
) (wormscanNotionalCache.NotionalLocalCacheReadable, error) {

	// use a distributed cache and for notional a pubsub to sync local cache.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.CacheURL})

	// get notional cache client and init load to local cache
	notionalCache, err := wormscanNotionalCache.NewNotionalCache(ctx, redisClient, cfg.CacheChannel, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create notional cache client: %w", err)
	}
	notionalCache.Init(ctx)

	return notionalCache, nil
}

// Database contains handles to MongoDB.
type Database struct {
	Database *mongo.Database
	client   *mongo.Client
}

// NewDatabase connects to DB and returns a client that will disconnect when the passed in context is cancelled.
func NewDatabase(appCtx context.Context, log *zap.Logger, uri, databaseName string) (*Database, error) {

	cli, err := mongo.Connect(appCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return &Database{client: cli, Database: cli.Database(databaseName)}, err
}

const databaseCloseDeadline = 30 * time.Second

// Close attempts to gracefully Close the database connection.
func (d *Database) Close() error {

	ctx, cancelFunc := context.WithDeadline(
		context.Background(),
		time.Now().Add(databaseCloseDeadline),
	)

	err := d.client.Disconnect(ctx)

	cancelFunc()
	return err
}
