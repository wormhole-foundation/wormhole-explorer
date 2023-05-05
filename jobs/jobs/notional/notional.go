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
	notionals := convertToSymbols(coingeckoNotionals)

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

// convertToSymbols converts the coingecko response into a symbol map
//
// The returned map has symbols as keys, and price data as the values.
func convertToSymbols(m map[string]coingecko.NotionalUSD) map[Symbol]notional.PriceData {

	w := make(map[Symbol]notional.PriceData, len(m))
	now := time.Now()

	for k, v := range m {

		// Do not update the dictionary when the token price is nil
		if v.Price == nil {
			continue
		}

		var symbol Symbol

		// Translate coingecko IDs into their associated ticker symbols, sorted alphabetically
		switch k {
		case "acala":
			symbol = "ACA"
		case "algorand":
			symbol = "ALGO"
		case "aptos":
			symbol = "APT"
		case "arbitrum":
			symbol = "ARB"
		case "aurora":
			symbol = "AURORA"
		case "avalanche-2":
			symbol = "AVAX"
		case "base-protocol":
			symbol = "BASE"
		case "binance-usd":
			symbol = "BUSD"
		case "binancecoin":
			symbol = "BNB"
		case "bitcoin":
			symbol = "BTC"
		case "celo":
			symbol = "CELO"
		case "dust-protocol":
			symbol = "DUST"
		case "ethereum":
			symbol = "ETH"
		case "injective-protocol":
			symbol = "INJ"
		case "fantom":
			symbol = "FTM"
		case "karura":
			symbol = "KAR"
		case "klay-token":
			symbol = "KLAY"
		case "matic-network":
			symbol = "MATIC"
		case "moonbeam":
			symbol = "GLMR"
		case "near":
			symbol = "NEAR"
		case "neon":
			symbol = "NEON"
		case "oasis-network":
			symbol = "ROSE"
		case "optimism":
			symbol = "OP"
		case "solana":
			symbol = "SOL"
		case "sui":
			symbol = "SUI"
		case "terra-luna":
			symbol = "LUNC"
		case "terra-luna-2":
			symbol = "LUNA"
		case "terrausd-wormhole":
			symbol = "UST"
		case "tether":
			symbol = "USDT"
		case "usd-coin":
			symbol = "USDC"
		case "wombat-exchange":
			symbol = "WOM"
		case "xpla":
			symbol = "XPLA"
		}

		if symbol != "" {
			w[symbol] = notional.PriceData{NotionalUsd: *v.Price, UpdatedAt: now}
		}
	}
	return w
}
