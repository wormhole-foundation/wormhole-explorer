// Package notional contains the logic to get the notional value of assets
package notional

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/internal/coingecko"
	"go.uber.org/zap"
)

type Symbol string

// NotionalJob is the job to get the notional value of assets.
type NotionalJob struct {
	coingeckoAPI *coingecko.CoingeckoAPI
	cacheClient  *redis.Client
	cacheChannel string
	p2pNetwork   string
	logger       *zap.Logger
}

// NewNotionalJob creates a new notional job.
func NewNotionalJob(api *coingecko.CoingeckoAPI, cacheClient *redis.Client, p2pNetwork, cacheChannel string, logger *zap.Logger) *NotionalJob {
	return &NotionalJob{
		coingeckoAPI: api,
		cacheClient:  cacheClient,
		cacheChannel: cacheChannel,
		p2pNetwork:   p2pNetwork,
		logger:       logger,
	}
}

// Run runs the notional job.
func (j *NotionalJob) Run() error {
	// get chains coingecko ids by p2p network.
	chainIDs := coingecko.GetChainIDs(j.p2pNetwork)
	if len(chainIDs) == 0 {
		return fmt.Errorf("no chain ids found for p2p network %s", j.p2pNetwork)
	}

	// get notional value of assets.
	coingeckoNotionals, err := j.coingeckoAPI.GetNotionalUSD(chainIDs)
	if err != nil {
		j.logger.Error("failed to get notional value of assets",
			zap.Error(err))
		return err
	}

	// convert notionals with coingecko assets ids to notionals with wormhole chainIDs.
	notionals := convertToWormholeChainIDs(coingeckoNotionals)

	// save notional value of assets in cache.
	err = j.updateNotionalCache(notionals)
	if err != nil {
		j.logger.Error("failed to update notional value of assets in cache",
			zap.Error(err),
			zap.Any("notionals", notionals))
		return err
	}

	// publish notional value of assets to redis pubsub.
	err = j.cacheClient.Publish(j.cacheChannel, "NOTIONAL_UPDATED").Err()
	if err != nil {
		j.logger.Error("failed to publish notional update message to redis pubsub",
			zap.Error(err))
		return err
	}

	return nil
}

// updateNotionalCache updates the notional value of assets in cache.
func (j *NotionalJob) updateNotionalCache(notionals map[Symbol]notional.PriceData) error {

	for chainID, n := range notionals {
		key := fmt.Sprintf(notional.KeyFormatString, chainID)
		err := j.cacheClient.Set(key, n, 0).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// convertToWormholeChainIDs converts the coingecko chain ids to wormhole chain ids.
func convertToWormholeChainIDs(m map[string]coingecko.NotionalUSD) map[Symbol]notional.PriceData {

	w := make(map[Symbol]notional.PriceData, len(m))
	now := time.Now()

	for k, v := range m {

		// Do not update the dictionary when the token price is nil
		if v.Price == nil {
			continue
		}

		var symbol Symbol

		switch k {
		case "solana":
			symbol = "SOL"
		case "ethereum":
			symbol = "ETH"
		case "terra-luna":
			symbol = "LUNC"
		case "binancecoin":
			symbol = "BNB"
		case "matic-network":
			symbol = "MATIC"
		case "avalanche-2":
			symbol = "AVAX"
		case "oasis-network":
			symbol = "ROSE"
		case "algorand":
			symbol = "ALGO"
		case "aurora":
			symbol = "AURORA"
		case "fantom":
			symbol = "FTM"
		case "karura":
			symbol = "KAR"
		case "acala":
			symbol = "ACA"
		case "klay-token":
			symbol = "KLAY"
		case "celo":
			symbol = "CELO"
		case "near":
			symbol = "NEAR"
		case "moonbeam":
			symbol = "GLMR"
		case "neon":
			symbol = "NEON"
		case "terra-luna-2":
			symbol = "LUNA"
		case "injective-protocol":
			symbol = "INJ"
		case "aptos":
			symbol = "APT"
		case "sui":
			symbol = "SUI"
		case "arbitrum":
			symbol = "ARB"
		case "optimism":
			symbol = "OP"
		case "xpla":
			symbol = "XPLA"
		case "bitcoin":
			symbol = "BTC"
		case "base-protocol":
			symbol = "BASE"
		case "tether":
			symbol = "USDT"
		case "usd-coin":
			symbol = "USDC"
		case "binance-usd":
			symbol = "BUSD"
		case "terrausd-wormhole":
			symbol = "UST"
		}

		if symbol != "" {
			w[symbol] = notional.PriceData{NotionalUsd: *v.Price, UpdatedAt: now}
		}
	}
	return w
}
