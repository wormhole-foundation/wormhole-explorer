package backfiller

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/builder"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/config"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/watcher"
	"go.uber.org/zap"
)

func Run(config *config.BackfillerConfiguration) {

	rootCtx := context.Background()

	logger := logger.New("wormhole-explorer-contract-watcher", logger.WithLevel(config.LogLevel))

	logger.Info("Starting wormhole-explorer-contract-watcher as backfiller ...")

	//setup DB connection
	db, err := dbutil.Connect(rootCtx, logger, config.MongoURI, config.MongoDatabase, false)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	// create metrics client
	metrics := metrics.NewNoopMetrics()

	// create alert client
	alerts := alert.NewDummyClient()

	// create repositories
	repo := storage.NewRepository(db.Database, metrics, alerts, logger)

	var watcher watcher.ContractWatcher

	switch config.Network {
	case domain.P2pMainNet:
		watcher = newWatcherForMainnet(config, repo, metrics, logger)
	case domain.P2pTestNet:
		watcher = newWatcherForTestnet(config, repo, metrics, logger)
	default:
		logger.Fatal("P2P network not supported")
	}

	logger.Info("Processing backfill ...",
		zap.String("network", config.Network),
		zap.String("chain", config.ChainName),
		zap.Bool("persistBlock", config.PersistBlock),
		zap.Uint64("from", config.FromBlock),
		zap.Uint64("to", config.ToBlock))

	watcher.Backfill(rootCtx, config.FromBlock, config.ToBlock, config.PageSize, config.PersistBlock)

	logger.Info("closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	logger.Info("Finish wormhole-explorer-contract-watcher as backfiller")

}

func newWatcherForMainnet(cfg *config.BackfillerConfiguration, repo *storage.Repository, metrics metrics.Metrics, logger *zap.Logger) watcher.ContractWatcher {
	var watcher watcher.ContractWatcher
	switch cfg.ChainName {
	case config.ETHEREUM_MAINNET.ChainID.String():
		watcher = builder.CreateAnkrEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.ETHEREUM_MAINNET, repo, metrics, logger)
	case config.BSC_MAINNET.ChainID.String():
		watcher = builder.CreateAnkrEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.BSC_MAINNET, repo, metrics, logger)
	case config.FANTOM_MAINNET.ChainID.String():
		watcher = builder.CreateAnkrEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.FANTOM_MAINNET, repo, metrics, logger)
	case config.AVALANCHE_MAINNET.ChainID.String():
		watcher = builder.CreateAnkrEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.AVALANCHE_MAINNET, repo, metrics, logger)
	case config.TERRA_MAINNET.ChainID.String():
		watcher = builder.CreateTerraWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.TERRA_MAINNET, logger, repo, metrics)
	case config.MOONBEAM_MAINNET.ChainID.String():
		watcher = builder.CreateEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.MOONBEAM_MAINNET, logger, repo, metrics)
	case config.CELO_MAINNET.ChainID.String():
		watcher = builder.CreateEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.CELO_MAINNET, logger, repo, metrics)
	case config.BASE_MAINNET.ChainID.String():
		watcher = builder.CreateEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.BASE_MAINNET, logger, repo, metrics)
	default:
		logger.Fatal("chain not supported")
	}
	return watcher
}

func newWatcherForTestnet(cfg *config.BackfillerConfiguration, repo *storage.Repository, metrics metrics.Metrics, logger *zap.Logger) watcher.ContractWatcher {
	var watcher watcher.ContractWatcher
	switch cfg.ChainName {
	case config.ETHEREUM_TESTNET.ChainID.String():
		watcher = builder.CreateAnkrEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.ETHEREUM_TESTNET, repo, metrics, logger)
	case config.BSC_TESTNET.ChainID.String():
		watcher = builder.CreateAnkrEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.BSC_TESTNET, repo, metrics, logger)
	case config.FANTOM_TESTNET.ChainID.String():
		watcher = builder.CreateAnkrEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.FANTOM_TESTNET, repo, metrics, logger)
	case config.AVALANCHE_TESTNET.ChainID.String():
		watcher = builder.CreateAnkrEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.AVALANCHE_TESTNET, repo, metrics, logger)
	case config.MOONBEAM_TESTNET.ChainID.String():
		watcher = builder.CreateEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.MOONBEAM_TESTNET, logger, repo, metrics)
	case config.CELO_TESTNET.ChainID.String():
		watcher = builder.CreateEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.CELO_TESTNET, logger, repo, metrics)
	case config.BASE_TESTNET.ChainID.String():
		watcher = builder.CreateEvmWatcher(cfg.RateLimitPerSecond, cfg.ChainUrl, config.BASE_TESTNET, logger, repo, metrics)
	default:
		logger.Fatal("chain not supported")
	}
	return watcher
}
