package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/wormhole-foundation/wormhole-explorer/parser/config"
	"github.com/wormhole-foundation/wormhole-explorer/parser/http/infraestructure"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole-explorer/parser/pipeline"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
	"github.com/wormhole-foundation/wormhole-explorer/parser/watcher"
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

	// get publish function.
	sqsConsumer, vaaPushFunc, vaaConsumeFunc := newVAAPublishAndConsume(config, logger)
	repository := parser.NewRepository(db.Database, logger)

	// // create a new publisher.
	publisher := pipeline.NewPublisher(logger, repository, vaaPushFunc)
	watcher := watcher.NewWatcher(db.Database, config.MongoDatabase, publisher.Publish, logger)
	err = watcher.Start(rootCtx)
	if err != nil {
		logger.Fatal("failed to watch MongoDB", zap.Error(err))
	}

	// create a consumer
	consumer := pipeline.NewConsumer(vaaConsumeFunc, repository, parserVAAAPIClient, logger)
	consumer.Start(rootCtx)

	server := infraestructure.NewServer(logger, config.Port, config.IsQueueConsumer(), sqsConsumer, db.Database)
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

func newAwsSession(cfg *config.Configuration) (*session.Session, error) {
	region := cfg.AwsRegion
	config := aws.NewConfig().WithRegion(region)
	if cfg.AwsAccessKeyID != "" && cfg.AwsSecretAccessKey != "" {
		config.WithCredentials(credentials.NewStaticCredentials(cfg.AwsAccessKeyID, cfg.AwsSecretAccessKey, ""))
	}
	if cfg.AwsEndpoint != "" {
		config.WithEndpoint(cfg.AwsEndpoint)
	}
	return session.NewSession(config)
}

// Creates two callbacks depending on whether the execution is local (memory queue) or not (SQS queue)
func newVAAPublishAndConsume(config *config.Configuration, logger *zap.Logger) (*sqs.Consumer, queue.VAAPushFunc, queue.VAAConsumeFunc) {
	// check is consumer type.
	if !config.IsQueueConsumer() {
		vaaQueue := queue.NewVAAInMemory()
		return nil, vaaQueue.Publish, vaaQueue.Consume
	}

	sqsConsumer, err := newSQSConsumer(config)
	if err != nil {
		logger.Fatal("failed to create sqs consumer", zap.Error(err))
	}

	sqsProducer, err := newSQSProducer(config)
	if err != nil {
		logger.Fatal("failed to create sqs producer", zap.Error(err))
	}

	vaaQueue := queue.NewVAASQS(sqsProducer, sqsConsumer, logger)
	return sqsConsumer, vaaQueue.Publish, vaaQueue.Consume
}

func newSQSProducer(config *config.Configuration) (*sqs.Producer, error) {
	session, err := newAwsSession(config)
	if err != nil {
		return nil, err
	}

	return sqs.NewProducer(session, config.SQSUrl)
}

func newSQSConsumer(config *config.Configuration) (*sqs.Consumer, error) {
	session, err := newAwsSession(config)
	if err != nil {
		return nil, err
	}

	return sqs.NewConsumer(session, config.SQSUrl,
		sqs.WithMaxMessages(10),
		sqs.WithVisibilityTimeout(120))
}
