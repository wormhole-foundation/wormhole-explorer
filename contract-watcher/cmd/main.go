package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	solana_go "github.com/gagliardetto/solana-go"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/config"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/ankr"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/aptos"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/evm"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/solana"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/terra"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/processor"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/watcher"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
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

func main() {
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

	// create repositories
	repo := storage.NewRepository(db.Database, logger)

	// create watchers
	watchers := newWatchers(config, repo, logger)

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
	chainID      vaa.ChainID
	name         string
	address      string
	sizeBlocks   uint8
	waitSeconds  uint16
	initialBlock int64
}

type watchersConfig struct {
	evms      []watcherBlockchain
	solana    *watcherBlockchain
	terra     *watcherBlockchain
	aptos     *watcherBlockchain
	oasis     *watcherBlockchain
	rateLimit rateLimitConfig
}

type rateLimitConfig struct {
	evm    int
	solana int
	terra  int
	aptos  int
	oasis  int
}

func newWatchers(config *config.Configuration, repo *storage.Repository, logger *zap.Logger) []watcher.ContractWatcher {
	var watchers *watchersConfig
	switch config.P2pNetwork {
	case domain.P2pMainNet:
		watchers = newEVMWatchersForMainnet()
	case domain.P2pTestNet:
		watchers = newEVMWatchersForTestnet()
	default:
		watchers = &watchersConfig{}
	}

	result := make([]watcher.ContractWatcher, 0)

	// add evm watchers
	evmLimiter := ratelimit.New(watchers.rateLimit.evm, ratelimit.Per(time.Second))
	ankrClient := ankr.NewAnkrSDK(config.AnkrUrl, evmLimiter)
	for _, w := range watchers.evms {
		params := watcher.EVMParams{ChainID: w.chainID, Blockchain: w.name, ContractAddress: w.address,
			SizeBlocks: w.sizeBlocks, WaitSeconds: w.waitSeconds, InitialBlock: w.initialBlock}
		result = append(result, watcher.NewEVMWatcher(ankrClient, repo, params, logger))
	}

	// add solana watcher
	if watchers.solana != nil {
		contractAddress, err := solana_go.PublicKeyFromBase58(watchers.solana.address)
		if err != nil {
			logger.Fatal("failed to parse solana contract address", zap.Error(err))
		}
		solanaLimiter := ratelimit.New(watchers.rateLimit.solana, ratelimit.Per(time.Second))
		solanaClient := solana.NewSolanaSDK(config.SolanaUrl, solanaLimiter)
		params := watcher.SolanaParams{Blockchain: watchers.solana.name, ContractAddress: contractAddress,
			SizeBlocks: watchers.solana.sizeBlocks, WaitSeconds: watchers.solana.waitSeconds, InitialBlock: watchers.solana.initialBlock}
		result = append(result, watcher.NewSolanaWatcher(solanaClient, repo, params, logger))
	}

	// add terra watcher
	if watchers.terra != nil {
		terraLimiter := ratelimit.New(watchers.rateLimit.terra, ratelimit.Per(time.Second))
		terraClient := terra.NewTerraSDK(config.TerraUrl, terraLimiter)
		params := watcher.TerraParams{ChainID: watchers.terra.chainID, Blockchain: watchers.terra.name,
			ContractAddress: watchers.terra.address, WaitSeconds: watchers.terra.waitSeconds, InitialBlock: watchers.terra.initialBlock}
		result = append(result, watcher.NewTerraWatcher(terraClient, params, repo, logger))
	}

	// add aptos watcher
	if watchers.aptos != nil {
		aptosLimiter := ratelimit.New(watchers.rateLimit.aptos, ratelimit.Per(time.Second))
		aptosClient := aptos.NewAptosSDK(config.AptosUrl, aptosLimiter)
		params := watcher.AptosParams{
			Blockchain:      watchers.aptos.name,
			ContractAddress: watchers.aptos.address,
			SizeBlocks:      watchers.aptos.sizeBlocks,
			WaitSeconds:     watchers.aptos.waitSeconds,
			InitialBlock:    watchers.aptos.initialBlock}
		result = append(result, watcher.NewAptosWatcher(aptosClient, params, repo, logger))
	}

	// add oasis watcher
	if watchers.oasis != nil {
		oasisLimiter := ratelimit.New(watchers.rateLimit.oasis, ratelimit.Per(time.Second))
		oasisClient := evm.NewEvmSDK(config.OasisUrl, oasisLimiter)
		params := watcher.EVMParams{
			ChainID:         watchers.oasis.chainID,
			Blockchain:      watchers.oasis.name,
			ContractAddress: watchers.oasis.address,
			SizeBlocks:      watchers.oasis.sizeBlocks,
			WaitSeconds:     watchers.oasis.waitSeconds,
			InitialBlock:    watchers.oasis.initialBlock}
		result = append(result, watcher.NewEvmStandarWatcher(oasisClient, params, repo, logger))
	}

	return result
}

func newEVMWatchersForMainnet() *watchersConfig {
	return &watchersConfig{
		evms: []watcherBlockchain{
			ETHEREUM_MAINNET,
			POLYGON_MAINNET,
			BSC_MAINNET,
			FANTOM_MAINNET,
			AVALANCHE_MAINNET,
		},
		solana: &SOLANA_MAINNET,
		terra:  &TERRA_MAINNET,
		aptos:  &APTOS_MAINNET,
		oasis:  &OASIS_MAINNET,
		rateLimit: rateLimitConfig{
			evm:    1000,
			solana: 3,
			terra:  10,
			aptos:  3,
			oasis:  3,
		},
	}
}

func newEVMWatchersForTestnet() *watchersConfig {
	return &watchersConfig{
		evms: []watcherBlockchain{
			ETHEREUM_TESTNET,
			POLYGON_TESTNET,
			BSC_TESTNET,
			FANTOM_TESTNET,
			AVALANCHE_TESTNET,
		},
		solana: &SOLANA_TESTNET,
		aptos:  &APTOS_TESTNET,
		oasis:  &OASIS_TESTNET,
		rateLimit: rateLimitConfig{
			evm:    10,
			solana: 2,
			terra:  5,
			aptos:  1,
			oasis:  1,
		},
	}
}
