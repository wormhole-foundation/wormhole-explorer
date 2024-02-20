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
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/configuration"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/consumer"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const layoutRcf3339 = "2006-01-02T15:04:05.000Z"

func makeLogger(logger *zap.Logger, name string) *zap.Logger {

	rightPadding := fmt.Sprintf("%-10s", name)

	l := logger.Named(rightPadding)

	return l
}

func main() {

	// Create the top-level context
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	// Load config
	cfg, err := configuration.LoadFromEnv[config.BackfillerSettings](rootCtx)
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Load testnet rpc config
	var testRpcConfig *config.TestnetRpcProviderSettings
	if configuration.IsTestnet(cfg.P2pNetwork) {
		testRpcConfig, err = configuration.LoadFromEnv[config.TestnetRpcProviderSettings](rootCtx)
		if err != nil {
			log.Fatal("Error loading testnet rpc config: ", err)
		}
	}

	// create rpc pool
	rpcPool, err := newRpcPool(cfg.RpcProviderSettings, testRpcConfig)
	if err != nil {
		log.Fatal("Failed to initialize rpc pool: ", zap.Error(err))
	}

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
	db, err := dbutil.Connect(rootCtx, mainLogger, cfg.MongodbUri, cfg.MongodbDatabase, false)
	if err != nil {
		log.Fatal("Failed to initialize MongoDB client: ", err)
	}
	repository := consumer.NewRepository(rootLogger, db.Database)

	strategyCallbacks, err := parseStrategyCallbacks(mainLogger, cfg, repository)
	if err != nil {
		log.Fatal("Failed to parse strategy callbacks: ", err)
	}

	// Count the number of documents to process
	totalDocuments, err := strategyCallbacks.countFn(rootCtx)
	if err != nil {
		log.Fatal("Closing - failed to count number of global transactions: ", err)
	}
	mainLogger.Info("Starting", zap.Uint64("documentsToProcess", totalDocuments))

	// Spawn the producer goroutine.
	//
	// The producer sends tasks to the workers via a buffered channel.
	queue := make(chan consumer.GlobalTransaction, cfg.BulkSize)
	p := producerParams{
		logger:            makeLogger(rootLogger, "producer"),
		repository:        repository,
		queueTx:           queue,
		bulkSize:          cfg.BulkSize,
		strategyCallbacks: strategyCallbacks,
	}
	go produce(rootCtx, &p)

	// Spawn a goroutine for each worker
	var wg sync.WaitGroup
	var processedDocuments atomic.Uint64
	wg.Add(int(cfg.NumWorkers))
	for i := uint(0); i < cfg.NumWorkers; i++ {
		name := fmt.Sprintf("worker-%d", i)
		p := consumerParams{
			logger:             makeLogger(rootLogger, name),
			rpcPool:            rpcPool,
			repository:         repository,
			queueRx:            queue,
			wg:                 &wg,
			totalDocuments:     totalDocuments,
			processedDocuments: &processedDocuments,
			p2pNetwork:         cfg.P2pNetwork,
		}
		go consume(rootCtx, &p)
	}

	// Wait for all workers to finish before closing
	mainLogger.Info("Waiting for all workers to finish...")
	wg.Wait()

	mainLogger.Info("Closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	mainLogger.Info("Closing main goroutine")
}

func parseStrategyCallbacks(
	logger *zap.Logger,
	cfg *config.BackfillerSettings,
	r *consumer.Repository,
) (*strategyCallbacks, error) {

	switch cfg.Strategy.Name {

	case config.BackfillerStrategyReprocessFailed:
		cb := strategyCallbacks{
			countFn:    r.CountIncompleteDocuments,
			iteratorFn: r.GetIncompleteDocuments,
		}

		logger.Info("backfilling incomplete documents")

		return &cb, nil

	case config.BackfillerStrategyTimeRange:

		timestampAfter, err := time.Parse(layoutRcf3339, cfg.Strategy.TimestampAfter)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestampAfter: %w", err)
		}
		timestampBefore, err := time.Parse(layoutRcf3339, cfg.Strategy.TimestampBefore)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestampBefore: %w", err)
		}

		cb := strategyCallbacks{
			countFn: func(ctx context.Context) (uint64, error) {
				return r.CountDocumentsByTimeRange(ctx, timestampAfter, timestampBefore)
			},
			iteratorFn: func(ctx context.Context, lastId string, lastTimestamp *time.Time, limit uint) ([]consumer.GlobalTransaction, error) {
				return r.GetDocumentsByTimeRange(ctx, lastId, lastTimestamp, limit, timestampAfter, timestampBefore)
			},
		}

		logger.Info("backfilling by time range",
			zap.Time("after", timestampAfter),
			zap.Time("before", timestampBefore),
		)

		return &cb, nil

	default:
		return nil, fmt.Errorf("unknown strategy: %s", cfg.Strategy.Name)
	}

}

type strategyCallbacks struct {
	countFn    func(ctx context.Context) (uint64, error)
	iteratorFn func(ctx context.Context, lastId string, lastTimestamp *time.Time, limit uint) ([]consumer.GlobalTransaction, error)
}

// producerParams contains the parameters for the producer goroutine.
type producerParams struct {
	logger            *zap.Logger
	repository        *consumer.Repository
	queueTx           chan<- consumer.GlobalTransaction
	bulkSize          uint
	strategyCallbacks *strategyCallbacks
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
	var lastId = ""
	var lastTimestamp *time.Time
	for {

		// Get a batch of VAA IDs from the database
		globalTxs, err := params.strategyCallbacks.iteratorFn(ctx, lastId, lastTimestamp, params.bulkSize)
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
				if len(globalTx.Vaas) != 0 {
					lastId = globalTx.Id
					lastTimestamp = globalTx.Vaas[0].Timestamp
				}
			case <-ctx.Done():
				params.logger.Info("Closing: context was cancelled")
				return
			}
		}
	}

}

// consumerParams contains the parameters for the consumer goroutine.
type consumerParams struct {
	logger             *zap.Logger
	rpcPool            map[sdk.ChainID]*pool.Pool
	repository         *consumer.Repository
	queueRx            <-chan consumer.GlobalTransaction
	wg                 *sync.WaitGroup
	totalDocuments     uint64
	processedDocuments *atomic.Uint64
	p2pNetwork         string
}

// consume reads VAA IDs from a channel, processes them, and updates the database accordingly.
//
// The function will return when:
// - the context is cancelled
// - a fatal error is encountered
// - the channel is closed (i.e.: no more items to process)
func consume(ctx context.Context, params *consumerParams) {

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
				Timestamp: v.Timestamp,
				VaaId:     v.ID,
				ChainId:   v.EmitterChain,
				Emitter:   v.EmitterAddr,
				Sequence:  v.Sequence,
				TxHash:    *v.TxHash,
				Overwrite: true, // Overwrite old contents
			}
			_, err := consumer.ProcessSourceTx(ctx, params.logger, params.rpcPool, params.repository, &p, params.p2pNetwork)
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

func newRpcPool(rpcSettings config.RpcProviderSettings,
	rpcTestnetSettings *config.TestnetRpcProviderSettings) (map[sdk.ChainID]*pool.Pool, error) {

	rpcPool := make(map[sdk.ChainID]*pool.Pool)

	// get rpc settings map
	rpcConfigMap, err := rpcSettings.ToMap()
	if err != nil {
		return nil, err
	}

	// get rpc testnet settings map
	var rpcTestnetMap map[sdk.ChainID][]config.RpcConfig
	if rpcTestnetSettings != nil {
		rpcTestnetMap, err = rpcTestnetSettings.ToMap()
		if err != nil {
			return nil, err
		}
	}

	// merge rpc testnet settings to rpc settings map
	if len(rpcTestnetMap) > 0 {
		for chainID, rpcConfig := range rpcTestnetMap {
			rpcConfigMap[chainID] = append(rpcConfigMap[chainID], rpcConfig...)
		}
	}

	// convert rpc settings map to rpc pool
	convertFn := func(rpcConfig []config.RpcConfig) []pool.Config {
		poolConfigs := make([]pool.Config, 0, len(rpcConfig))
		for _, rpc := range rpcConfig {
			poolConfigs = append(poolConfigs, pool.Config{
				Id:                rpc.Url,
				Priority:          rpc.Priority,
				RequestsPerMinute: rpc.RequestsPerMinute,
			})
		}
		return poolConfigs
	}

	// create rpc pool
	for chainID, rpcConfig := range rpcConfigMap {
		rpcPool[chainID] = pool.NewPool(convertFn(rpcConfig))
	}

	return rpcPool, nil
}
