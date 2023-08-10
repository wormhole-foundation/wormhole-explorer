package builder

import (
	"time"

	solana_go "github.com/gagliardetto/solana-go"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/config"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/ankr"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/aptos"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/evm"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/solana"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/internal/terra"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/storage"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/watcher"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

func CreateAnkrEvmWatcher(rateLimit int, chainURL string, wb config.WatcherBlockchainAddresses, repo *storage.Repository,
	metrics metrics.Metrics, logger *zap.Logger) watcher.ContractWatcher {
	evmLimiter := ratelimit.New(rateLimit, ratelimit.Per(time.Second))
	ankrClient := ankr.NewAnkrSDK(chainURL, evmLimiter, metrics)
	params := watcher.EVMParams{ChainID: wb.ChainID, Blockchain: wb.Name, SizeBlocks: wb.SizeBlocks,
		WaitSeconds: wb.WaitSeconds, InitialBlock: wb.InitialBlock, MethodsByAddress: wb.MethodsByAddress}
	return watcher.NewEVMWatcher(ankrClient, repo, params, metrics, logger)
}

func CreateSolanaWatcher(rateLimit int, chainURL string, wb config.WatcherBlockchain, logger *zap.Logger, repo *storage.Repository, metrics metrics.Metrics) watcher.ContractWatcher {
	contractAddress, err := solana_go.PublicKeyFromBase58(wb.Address)
	if err != nil {
		logger.Fatal("failed to parse solana contract address", zap.Error(err))
	}
	solanaLimiter := ratelimit.New(rateLimit, ratelimit.Per(time.Second))
	solanaClient := solana.NewSolanaSDK(chainURL, solanaLimiter, metrics, solana.WithRetries(3, 10*time.Second))
	params := watcher.SolanaParams{Blockchain: wb.Name, ContractAddress: contractAddress,
		SizeBlocks: wb.SizeBlocks, WaitSeconds: wb.WaitSeconds, InitialBlock: wb.InitialBlock}
	return watcher.NewSolanaWatcher(solanaClient, repo, params, metrics, logger)
}

func CreateTerraWatcher(rateLimit int, chainURL string, wb config.WatcherBlockchain, logger *zap.Logger, repo *storage.Repository, metrics metrics.Metrics) watcher.ContractWatcher {
	terraLimiter := ratelimit.New(rateLimit, ratelimit.Per(time.Second))
	terraClient := terra.NewTerraSDK(chainURL, terraLimiter, metrics)
	params := watcher.TerraParams{ChainID: wb.ChainID, Blockchain: wb.Name,
		ContractAddress: wb.Address, WaitSeconds: wb.WaitSeconds, InitialBlock: wb.InitialBlock}
	return watcher.NewTerraWatcher(terraClient, params, repo, metrics, logger)
}

func CreateAptosWatcher(rateLimit int, chainURL string, wb config.WatcherBlockchain, logger *zap.Logger, repo *storage.Repository, metrics metrics.Metrics) watcher.ContractWatcher {
	aptosLimiter := ratelimit.New(rateLimit, ratelimit.Per(time.Second))
	aptosClient := aptos.NewAptosSDK(chainURL, aptosLimiter, metrics)
	params := watcher.AptosParams{
		Blockchain:      wb.Name,
		ContractAddress: wb.Address,
		SizeBlocks:      wb.SizeBlocks,
		WaitSeconds:     wb.WaitSeconds,
		InitialBlock:    wb.InitialBlock}
	return watcher.NewAptosWatcher(aptosClient, params, repo, metrics, logger)
}

func CreateEvmWatcher(
	rateLimit int,
	chainURL string,
	wb config.WatcherBlockchainAddresses,
	logger *zap.Logger,
	repo *storage.Repository,
	metrics metrics.Metrics,
) watcher.ContractWatcher {

	limiter := ratelimit.New(rateLimit, ratelimit.Per(time.Second))
	client := evm.NewEvmSDK(chainURL, limiter, metrics)
	params := watcher.EVMParams{
		ChainID:          wb.ChainID,
		Blockchain:       wb.Name,
		SizeBlocks:       wb.SizeBlocks,
		WaitSeconds:      wb.WaitSeconds,
		InitialBlock:     wb.InitialBlock,
		MethodsByAddress: wb.MethodsByAddress,
	}

	return watcher.NewEvmStandarWatcher(client, params, repo, metrics, logger)
}
