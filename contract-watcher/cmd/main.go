package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/config"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/ankr"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/processor"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/watcher"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
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

	logger := ipfslog.Logger("wormhole-explorer-contract-watcher").Desugar()
	ipfslog.SetAllLoggers(level)

	logger.Info("Starting wormhole-explorer-contract-watcher ...")

	//setup DB connection
	db, err := db.New(rootCtx, logger, config.MongoURI, config.MongoDatabase)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	// get health check functions.
	healthChecks, err := newHealthChecks(rootCtx, db.Database)
	if err != nil {
		logger.Fatal("failed to create health checks", zap.Error(err))
	}

	// create repositories
	repo := storage.NewRepository(db.Database, logger)

	// create watchers
	watchers := newWatchers(config.P2pNetwork, config.AnkrUrl, repo, logger)

	//create processor
	processor := processor.NewProcessor(watchers, logger)
	processor.Start(rootCtx)

	// create and start server.
	server := infrastructure.NewServer(logger, config.Port, config.PprofEnabled, healthChecks...)
	server.Start()

	logger.Info("Started wormhole-explorer-contract-watcher")

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

	logger.Info("Closing processor ...")
	processor.Close()
	logger.Info("Closing database connections ...")
	db.Close()
	logger.Info("Closing Http server ...")
	server.Stop()
	logger.Info("Finished wormhole-explorer-contract-watcher")
}

func newHealthChecks(ctx context.Context, db *mongo.Database) ([]health.Check, error) {
	return []health.Check{health.Mongo(db)}, nil
}

type watcherBlockchain struct {
	chainID     vaa.ChainID
	name        string
	address     string
	sizeBlocks  uint8
	waitSeconds uint16
}

func newWatchers(p2pNetwork, ankUrl string, repo *storage.Repository, logger *zap.Logger) []watcher.ContractWatcher {
	var watchers []watcherBlockchain
	switch p2pNetwork {
	case domain.P2pMainNet:
		watchers = newWatchersForMainnet()
	default:
		watchers = []watcherBlockchain{}
	}
	client := ankr.NewAnkrSDK(ankUrl)
	result := make([]watcher.ContractWatcher, 0)
	for _, w := range watchers {
		params := watcher.EVMParams{ChainID: w.chainID, Blockchain: w.name, ContractAddress: w.address,
			SizeBlocks: w.sizeBlocks, WaitSeconds: w.waitSeconds}
		result = append(result, watcher.NewEVMWatcher(client, repo, params, logger))
	}
	return result
}

func newWatchersForMainnet() []watcherBlockchain {
	return []watcherBlockchain{
		{vaa.ChainIDEthereum, "eth", "0x3ee18B2214AFF97000D974cf647E7C347E8fa585", 100, 10},
		{vaa.ChainIDPolygon, "polygon", "0x5a58505a96D1dbf8dF91cB21B54419FC36e93fdE", 100, 10},
		{vaa.ChainIDBSC, "bsc", "0xB6F6D86a8f9879A9c87f643768d9efc38c1Da6E7", 100, 10},
		{vaa.ChainIDFantom, "fantom", "0x7C9Fc5741288cDFdD83CeB07f3ea7e22618D79D2", 100, 10},
	}
}
