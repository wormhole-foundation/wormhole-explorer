package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/consumer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const (
	// numWorker determines the number of goroutines that fetch tx data from RPC/API services.
	numWorkers = 10

	// queueSize determines the maximum number of global transactions that the producer can enqueue.
	queueSize = 100
)

func makeLogger(logger *zap.Logger, name string) *zap.Logger {

	rightPadding := fmt.Sprintf("%-10s", name)

	l := logger.Named(rightPadding)

	return l
}

func main() {

	// Create the top-level context
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	// Load config
	cfg, err := config.LoadFromEnv[config.BackfillerSettings]()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Initialize rate limiters
	chains.Initialize(&cfg.RpcProviderSettings)

	// Initialize logger
	level, err := ipfslog.LevelFromString(cfg.LogLevel)
	if err != nil {
		log.Fatal("Invalid log level: ", err)
	}
	rootLogger := ipfslog.Logger("backfiller").Desugar()
	ipfslog.SetAllLoggers(level)
	mainLogger := makeLogger(rootLogger, "main")
	mainLogger.Info("Starting")

	// Spawn a goroutine that will call `cancelFunc` if a signal is received.
	go func() {
		l := makeLogger(rootLogger, "watcher")
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-rootCtx.Done():
			l.Info("Closing due to cancelled context")
		case _ = <-sigterm:
			l.Info("Cancelling root context")
			rootCtxCancel()
		}
	}()

	// Initialize the database client
	cli, err := mongo.Connect(rootCtx, options.Client().ApplyURI(cfg.MongodbUri))
	if err != nil {
		mainLogger.Error("Failed to initialize MongoDB client", zap.Error(err))
		return
	}
	defer cli.Disconnect(rootCtx)
	db := cli.Database(cfg.MongodbDatabase)
	globalTransactions := db.Collection("globalTransactions")

	// Count the number of documents to process
	totalDocuments, err := countGlobalTransactions(rootCtx, mainLogger, globalTransactions)
	if err != nil {
		mainLogger.Error("Closing: failed to count number of global transactions", zap.Error(err))
		return
	}
	mainLogger.Info("Starting", zap.Uint64("documentsToProcess", totalDocuments))

	// Spawn the producer goroutine.
	//
	// The producer sends tasks to the workers via a buffered channel.
	queue := make(chan globalTransaction, queueSize)
	p := producerParams{
		logger:             makeLogger(rootLogger, "producer"),
		globalTransactions: globalTransactions,
		queueTx:            queue,
	}
	go produce(rootCtx, &p)

	// Spawn a goroutine for each worker
	var wg sync.WaitGroup
	var processedDocuments atomic.Uint64
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		name := fmt.Sprintf("worker-%d", i)
		p := consumerParams{
			logger:                   makeLogger(rootLogger, name),
			vaaPayloadParserSettings: &cfg.VaaPayloadParserSettings,
			rpcProviderSettings:      &cfg.RpcProviderSettings,
			db:                       db,
			queueRx:                  queue,
			wg:                       &wg,
			totalDocuments:           totalDocuments,
			processedDocuments:       &processedDocuments,
		}
		go consume(rootCtx, &p)
	}

	// Wait for all workers to finish before closing
	wg.Wait()
	mainLogger.Info("Closing main goroutine")
}

// producerParams contains the parameters for the producer goroutine.
type producerParams struct {
	logger             *zap.Logger
	globalTransactions *mongo.Collection
	queueTx            chan<- globalTransaction
}

// produce reads VAA IDs from the database, and sends them through a channel for the workers to consume.
//
// The function will return when:
// - the context is cancelled
// - a fatal error is encountered
// - there are no more items to process
func produce(ctx context.Context, params *producerParams) {
	defer close(params.queueTx)

	// Producer main loop
	var maxId = ""
	for {

		// Get a batch of VAA IDs from the database
		globalTxs, err := queryGlobalTransactions(ctx, params.logger, params.globalTransactions, maxId)
		if err != nil {
			params.logger.Error("Closing: failed to read from cursor", zap.Error(err))
			return
		}

		// If there are no more documents to process, close the goroutine
		if len(globalTxs) == 0 {
			params.logger.Info("Closing: no documents left to process")
			return
		}

		// Enqueue the VAA IDs, and update the pagination cursor
		params.logger.Debug("queueing batch for consumers", zap.Int("elements", len(globalTxs)))
		for _, globalTx := range globalTxs {
			select {
			case params.queueTx <- globalTx:
				maxId = globalTx.Id
			case <-ctx.Done():
				params.logger.Info("Closing: context was cancelled")
				return
			}
		}
	}

}

func countGlobalTransactions(
	ctx context.Context,
	logger *zap.Logger,
	globalTransactions *mongo.Collection,
) (uint64, error) {

	// Build the aggregation pipeline
	var pipeline mongo.Pipeline
	{
		// Look up transactions that have not been processed by the tx-tracker
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{{"originTx", bson.M{"$exists": false}}}},
		})

		// Count the number of results
		pipeline = append(pipeline, bson.D{
			{"$count", "numGlobalTransactions"},
		})
	}

	// Execute the aggregation pipeline
	cur, err := globalTransactions.Aggregate(ctx, pipeline)
	if err != nil {
		logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return 0, err
	}

	// Read results from cursor
	var results []struct {
		NumGlobalTransactions uint64 `bson:"numGlobalTransactions"`
	}
	err = cur.All(ctx, &results)
	if err != nil {
		logger.Error("failed to decode cursor", zap.Error(err))
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	if len(results) > 1 {
		logger.Error("too many results", zap.Int("numResults", len(results)))
		return 0, err
	}

	return results[0].NumGlobalTransactions, nil
}

type globalTransaction struct {
	Id   string       `bson:"_id"`
	Vaas []vaa.VaaDoc `bson:"vaas"`
}

// queryGlobalTransactions gets a batch of VAA IDs from the database.
func queryGlobalTransactions(
	ctx context.Context,
	logger *zap.Logger,
	globalTransactions *mongo.Collection,
	maxId string,
) ([]globalTransaction, error) {

	// Build the aggregation pipeline
	var pipeline mongo.Pipeline
	{
		// Specify sorting criteria
		pipeline = append(pipeline, bson.D{
			{"$sort", bson.D{bson.E{"_id", 1}}},
		})

		// filter out already processed documents
		//
		// We use the _id field as a pagination cursor
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{{"_id", bson.M{"$gt": maxId}}}},
		})

		// Look up transactions that have not been processed by the tx-tracker
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{{"originTx", bson.M{"$exists": false}}}},
		})

		// Left join on the VAA collection
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "vaas"},
				{"localField", "_id"},
				{"foreignField", "_id"},
				{"as", "vaas"},
			}},
		})

		// Limit size of results
		pipeline = append(pipeline, bson.D{
			{"$limit", queueSize},
		})
	}

	// Execute the aggregation pipeline
	cur, err := globalTransactions.Aggregate(ctx, pipeline)
	if err != nil {
		logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return nil, errors.WithStack(err)
	}

	// Read results from cursor
	var documents []globalTransaction
	err = cur.All(ctx, &documents)
	if err != nil {
		logger.Error("failed to decode cursor", zap.Error(err))
		return nil, errors.WithStack(err)
	}

	return documents, nil
}

// consumerParams contains the parameters for the consumer goroutine.
type consumerParams struct {
	logger                   *zap.Logger
	vaaPayloadParserSettings *config.VaaPayloadParserSettings
	rpcProviderSettings      *config.RpcProviderSettings
	db                       *mongo.Database
	queueRx                  <-chan globalTransaction
	wg                       *sync.WaitGroup
	totalDocuments           uint64
	processedDocuments       *atomic.Uint64
}

// consume reads VAA IDs from a channel, processes them, and updates the database accordingly.
//
// The function will return when:
// - the context is cancelled
// - a fatal error is encountered
// - the channel is closed (i.e.: no more items to process)
func consume(ctx context.Context, params *consumerParams) {

	// Initialize the client, which processes source Txs.
	client, err := consumer.New(
		nil,
		params.vaaPayloadParserSettings,
		params.rpcProviderSettings,
		params.logger,
		params.db,
	)
	if err != nil {
		params.logger.Error("Failed to initialize consumer", zap.Error(err))
		params.wg.Done()
		return
	}

	// Main loop: fetch global txs and process them
	for {
		select {

		// Try to pop a globalTransaction from the queue
		case globalTx, ok := <-params.queueRx:

			// If the channel was closed, exit immediately
			if !ok {
				params.logger.Info("Closing, no more documents to process")
				params.wg.Done()
				return
			}

			// Sanity check
			if len(globalTx.Vaas) != 1 {
				params.logger.Warn("globalTransaction doesn't match exactly one VAA, skipping",
					zap.String("vaaId", globalTx.Id),
					zap.Int("matches", len(globalTx.Vaas)),
				)
				params.processedDocuments.Add(1)
				continue
			}

			params.logger.Debug("Processing source tx",
				zap.String("vaaId", globalTx.Id),
				zap.String("txid", *globalTx.Vaas[0].TxHash),
			)

			// Process the transaction
			//
			// This involves:
			// 1. Querying an API/RPC service for the source tx details
			// 2. Persisting source tx details in the database.
			v := globalTx.Vaas[0]
			p := consumer.ProcessSourceTxParams{
				VaaId:    v.ID,
				ChainId:  v.EmitterChain,
				Emitter:  v.EmitterAddr,
				Sequence: v.Sequence,
				TxHash:   *v.TxHash,
			}
			err = client.ProcessSourceTx(ctx, &p)
			if err != nil {
				params.logger.Error("Failed to track source tx",
					zap.String("vaaId", globalTx.Id),
					zap.Error(err),
				)
				params.processedDocuments.Add(1)
				continue
			}

			params.processedDocuments.Add(1)
			params.logger.Debug("Updated source tx",
				zap.String("vaaId", globalTx.Id),
				zap.String("txid", *globalTx.Vaas[0].TxHash),
				zap.String("progress", fmt.Sprintf("%d/%d", params.processedDocuments.Load(), params.totalDocuments)),
			)

		// If the context was cancelled, exit immediately
		case <-ctx.Done():
			params.logger.Info("Closing due to cancelled context")
			params.wg.Done()
			return
		}

	}

}
