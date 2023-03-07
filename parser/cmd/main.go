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
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/wormhole-foundation/wormhole-explorer/parser/config"
	"github.com/wormhole-foundation/wormhole-explorer/parser/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/parser/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/parser/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole-explorer/parser/processor"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
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

	level, err := ipfslog.LevelFromString(config.LogLevel)
	if err != nil {
		log.Fatal("Invalid log level", err)
	}

	logger := ipfslog.Logger("wormhole-explorer-parser").Desugar()
	ipfslog.SetAllLoggers(level)

	logger.Info("Starting wormhole-explorer-parser ...")

	//setup DB connection
	db, err := db.New(rootCtx, logger, config.MongoURI, config.MongoDatabase)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	parserVAAAPIClient, err := parser.NewParserVAAAPIClient(config.VaaPayloadParserTimeout,
		config.VaaPayloadParserURL, logger)
	if err != nil {
		logger.Fatal("failed to create parse vaa api client")
	}

	// get consumer function.
	sqsConsumer, vaaConsumeFunc := newVAAConsume(rootCtx, config, logger)
	repository := parser.NewRepository(db.Database, logger)

	//create a processor
	influxCli := newInfluxClient(config.InfluxUrl, config.InfluxToken)
	metrics := metrics.New(influxCli, config.InfluxOrg, config.InfluxBucket, logger)
	processor := processor.New(repository, metrics, logger)

	// create and start a consumer
	consumer := consumer.New(vaaConsumeFunc, processor.Process, parserVAAAPIClient, logger)
	consumer.Start(rootCtx)

	server := infrastructure.NewServer(logger, config.Port, config.PprofEnabled, config.IsQueueConsumer(), sqsConsumer, db.Database)
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

	logger.Info("Closing database connections ...")
	db.Close()
	logger.Info("Closing Http server ...")
	server.Stop()
	logger.Info("Finished wormhole-explorer-parser")
}

func newAwsConfig(appCtx context.Context, cfg *config.Configuration) (aws.Config, error) {
	region := cfg.AwsRegion
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

// Creates a callbacks depending on whether the execution is local (memory queue) or not (SQS queue)
func newVAAConsume(appCtx context.Context, config *config.Configuration, logger *zap.Logger) (*sqs.Consumer, queue.VAAConsumeFunc) {
	sqsConsumer, err := newSQSConsumer(appCtx, config)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	filterConsumeFunc := newFilterFunc(config)
	vaaQueue := queue.NewVAASQS(sqsConsumer, filterConsumeFunc, logger)
	return sqsConsumer, vaaQueue.Consume
}

func newSQSConsumer(appCtx context.Context, config *config.Configuration) (*sqs.Consumer, error) {
	awsconfig, err := newAwsConfig(appCtx, config)
	if err != nil {
		return nil, err
	}

	return sqs.NewConsumer(awsconfig, config.SQSUrl,
		sqs.WithMaxMessages(10),
		sqs.WithVisibilityTimeout(120))
}

func newFilterFunc(cfg *config.Configuration) queue.FilterConsumeFunc {
	if cfg.P2pNetwork == config.P2pMainNet {
		return queue.PythFilter
	}
	return queue.NonFilter
}

func newInfluxClient(url, token string) influxdb2.Client {
	return influxdb2.NewClient(url, token)
}
