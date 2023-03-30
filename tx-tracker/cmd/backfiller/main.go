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

	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/consumer"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
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
	rootLogger := logger.New("backfiller", logger.WithLevel(cfg.LogLevel))
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
		log.Fatal("Failed to initialize MongoDB client: ", err)
	}
	defer cli.Disconnect(rootCtx)
	repository := consumer.NewRepository(rootLogger, cli.Database(cfg.MongodbDatabase))

	// Count the number of documents to process
	totalDocuments, err := repository.CountIncompleteDocuments(rootCtx)
	if err != nil {
		log.Fatal("Closing - failed to count number of global transactions: ", err)
	}
	mainLogger.Info("Starting", zap.Uint64("documentsToProcess", totalDocuments))

	// Spawn the producer goroutine.
	//
	// The producer sends tasks to the workers via a buffered channel.
	queue := make(chan consumer.GlobalTransaction, cfg.BulkSize)
	p := producerParams{
		logger:     makeLogger(rootLogger, "producer"),
		repository: repository,
		queueTx:    queue,
		bulkSize:   cfg.BulkSize,
	}
	go produce(rootCtx, &p)

	// Spawn a goroutine for each worker
	var wg sync.WaitGroup
	var processedDocuments atomic.Uint64
	wg.Add(int(cfg.NumWorkers))
	for i := uint(0); i < cfg.NumWorkers; i++ {
		name := fmt.Sprintf("worker-%d", i)
		p := consumerParams{
			logger:                   makeLogger(rootLogger, name),
			vaaPayloadParserSettings: &cfg.VaaPayloadParserSettings,
			rpcProviderSettings:      &cfg.RpcProviderSettings,
			repository:               repository,
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
	logger     *zap.Logger
	repository *consumer.Repository
	queueTx    chan<- consumer.GlobalTransaction
	bulkSize   uint
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
		globalTxs, err := params.repository.GetIncompleteDocuments(ctx, maxId, params.bulkSize)
		if err != nil {
			params.logger.Error("Closing: failed to read from cursor", zap.Error(err))
			return
		}

		// If there are no more documents to process, close the goroutine
		if len(globalTxs) == 0 {
			params.logger.Info("Closing: no documents left in database")
			return
		}

		// Enqueue the VAA IDs, and update the pagination cursor
		params.logger.Debug("queueing batch for consumers", zap.Int("numElements", len(globalTxs)))
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

// consumerParams contains the parameters for the consumer goroutine.
type consumerParams struct {
	logger                   *zap.Logger
	vaaPayloadParserSettings *config.VaaPayloadParserSettings
	rpcProviderSettings      *config.RpcProviderSettings
	repository               *consumer.Repository
	queueRx                  <-chan consumer.GlobalTransaction
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
		params.repository,
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
				params.logger.Info("Closing, channel was closed")
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
			if globalTx.Vaas[0].TxHash == nil {
				params.logger.Warn("VAA doesn't have a TxHash, skipping",
					zap.String("vaaId", globalTx.Id),
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
