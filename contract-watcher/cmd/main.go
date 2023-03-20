package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	solana_go "github.com/gagliardetto/solana-go"
	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/config"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/ankr"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/solana"
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
	rateLimit rateLimitConfig
}

type rateLimitConfig struct {
	evm    int
	solana int
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
	return result
}

func newEVMWatchersForMainnet() *watchersConfig {
	return &watchersConfig{
		evms: []watcherBlockchain{
			{vaa.ChainIDEthereum, "eth", "0x3ee18B2214AFF97000D974cf647E7C347E8fa585", 100, 10, 16820790},
			{vaa.ChainIDPolygon, "polygon", "0x5a58505a96D1dbf8dF91cB21B54419FC36e93fdE", 100, 10, 40307020},
			{vaa.ChainIDBSC, "bsc", "0xB6F6D86a8f9879A9c87f643768d9efc38c1Da6E7", 100, 10, 26436320},
			{vaa.ChainIDFantom, "fantom", "0x7C9Fc5741288cDFdD83CeB07f3ea7e22618D79D2", 100, 10, 57525624},
		},
		solana: &watcherBlockchain{vaa.ChainIDSolana, "solana", "wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb", 100, 10, 182730391},
		rateLimit: rateLimitConfig{
			evm:    1000,
			solana: 3,
		},
	}
}

func newEVMWatchersForTestnet() *watchersConfig {
	return &watchersConfig{
		evms: []watcherBlockchain{
			{vaa.ChainIDEthereum, "eth_goerli", "0xF890982f9310df57d00f659cf4fd87e65adEd8d7", 100, 10, 8660321},
			{vaa.ChainIDPolygon, "polygon_mumbai", "0x377D55a7928c046E18eEbb61977e714d2a76472a", 100, 10, 33151522},
			{vaa.ChainIDBSC, "bsc_testnet_chapel", "0x9dcF9D205C9De35334D646BeE44b2D2859712A09", 100, 10, 28071327},
			{vaa.ChainIDFantom, "fantom_testnet", "0x599CEa2204B4FaECd584Ab1F2b6aCA137a0afbE8", 100, 10, 14524466},
		},
		solana: &watcherBlockchain{vaa.ChainIDSolana, "solana", "DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe", 50, 10, 16820790},
		rateLimit: rateLimitConfig{
			evm:    10,
			solana: 2,
		},
	}
}
