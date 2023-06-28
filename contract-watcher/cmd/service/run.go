package service

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/builder"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/config"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/ankr"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/processor"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/watcher"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/ratelimit"
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

type watchersConfig struct {
	evms      []config.WatcherBlockchainAddresses
	solana    *config.WatcherBlockchain
	terra     *config.WatcherBlockchain
	aptos     *config.WatcherBlockchain
	oasis     *config.WatcherBlockchainAddresses
	moonbeam  *config.WatcherBlockchainAddresses
	celo      *config.WatcherBlockchainAddresses
	rateLimit rateLimitConfig
}

type rateLimitConfig struct {
	evm      int
	solana   int
	terra    int
	aptos    int
	oasis    int
	moonbeam int
	celo     int
}

func Run() {

	defer handleExit()
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	config, err := config.New(rootCtx)
	if err != nil {
		log.Fatal("Error creating config", err)
	}

	logger := logger.New("wormhole-explorer-contract-watcher", logger.WithLevel(config.LogLevel))

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

	metrics := metrics.NewPrometheusMetrics(config.Environment)

	// create repositories
	repo := storage.NewRepository(db.Database, metrics, logger)

	// create watchers
	watchers := newWatchers(config, repo, metrics, logger)

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

func newWatchers(config *config.ServiceConfiguration, repo *storage.Repository, metrics metrics.Metrics, logger *zap.Logger) []watcher.ContractWatcher {
	var watchers *watchersConfig
	switch config.P2pNetwork {
	case domain.P2pMainNet:
		watchers = newWatchersForMainnet()
	case domain.P2pTestNet:
		watchers = newWatchersForTestnet()
	default:
		watchers = &watchersConfig{}
	}

	result := make([]watcher.ContractWatcher, 0)

	// add evm watchers
	evmLimiter := ratelimit.New(watchers.rateLimit.evm, ratelimit.Per(time.Second))
	ankrClient := ankr.NewAnkrSDK(config.AnkrUrl, evmLimiter, metrics)
	for _, w := range watchers.evms {
		params := watcher.EVMParams{ChainID: w.ChainID, Blockchain: w.Name, SizeBlocks: w.SizeBlocks,
			WaitSeconds: w.WaitSeconds, InitialBlock: w.InitialBlock, MethodsByAddress: w.MethodsByAddress}
		result = append(result, watcher.NewEVMWatcher(ankrClient, repo, params, metrics, logger))
	}

	// add solana watcher
	if watchers.solana != nil {
		solanWatcher := builder.CreateSolanaWatcher(watchers.rateLimit.solana, config.SolanaUrl, *watchers.solana, logger, repo, metrics)
		result = append(result, solanWatcher)
	}

	// add terra watcher
	if watchers.terra != nil {
		terraWatcher := builder.CreateTerraWatcher(watchers.rateLimit.terra, config.TerraUrl, *watchers.terra, logger, repo, metrics)
		result = append(result, terraWatcher)
	}

	// add aptos watcher
	if watchers.aptos != nil {
		aptosWatcher := builder.CreateAptosWatcher(watchers.rateLimit.aptos, config.AptosUrl, *watchers.aptos, logger, repo, metrics)
		result = append(result, aptosWatcher)
	}

	// add oasis watcher
	if watchers.oasis != nil {
		oasisWatcher := builder.CreateOasisWatcher(watchers.rateLimit.oasis, config.OasisUrl, *watchers.oasis, logger, repo, metrics)
		result = append(result, oasisWatcher)
	}

	// add moonbeam watcher
	if watchers.moonbeam != nil {
		moonbeamWatcher := builder.CreateMoonbeamWatcher(watchers.rateLimit.moonbeam, config.MoonbeamUrl, *watchers.moonbeam, logger, repo, metrics)
		result = append(result, moonbeamWatcher)
	}

	if watchers.celo != nil {
		celoWatcher := builder.CreateCeloWatcher(watchers.rateLimit.evm, config.CeloUrl, *watchers.celo, logger, repo, metrics)
		result = append(result, celoWatcher)
	}
	return result
}

func newWatchersForMainnet() *watchersConfig {
	return &watchersConfig{
		evms: []config.WatcherBlockchainAddresses{
			config.ETHEREUM_MAINNET,
			config.POLYGON_MAINNET,
			config.BSC_MAINNET,
			config.FANTOM_MAINNET,
			config.AVALANCHE_MAINNET,
		},
		solana:   &config.SOLANA_MAINNET,
		terra:    &config.TERRA_MAINNET,
		aptos:    &config.APTOS_MAINNET,
		oasis:    &config.OASIS_MAINNET,
		moonbeam: &config.MOONBEAM_MAINNET,
		celo:     &config.CELO_MAINNET,
		rateLimit: rateLimitConfig{
			evm:      1000,
			solana:   3,
			terra:    10,
			aptos:    3,
			oasis:    3,
			moonbeam: 5,
			celo:     3,
		},
	}
}

func newWatchersForTestnet() *watchersConfig {
	return &watchersConfig{
		evms: []config.WatcherBlockchainAddresses{
			config.ETHEREUM_TESTNET,
			config.POLYGON_TESTNET,
			config.BSC_TESTNET,
			config.FANTOM_TESTNET,
			config.AVALANCHE_TESTNET,
		},
		solana:   &config.SOLANA_TESTNET,
		aptos:    &config.APTOS_TESTNET,
		oasis:    &config.OASIS_TESTNET,
		moonbeam: &config.MOONBEAM_TESTNET,
		celo:     &config.CELO_TESTNET,
		rateLimit: rateLimitConfig{
			evm:      10,
			solana:   2,
			terra:    5,
			aptos:    1,
			oasis:    1,
			moonbeam: 2,
			celo:     3,
		},
	}
}
