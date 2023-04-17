// Package notional contains the logic to get the notional value of assets
package notional

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/internal/coingecko"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// NotionalCacheKey is the cache key for notional value by chainID
const NotionalCacheKey = "WORMSCAN:NOTIONAL:CHAIN_ID:%d"

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
	err = j.cacheClient.Publish(j.cacheChannel, "NOTIONA_UPDATED").Err()
	if err != nil {
		j.logger.Error("failed to publish notional update message to redis pubsub",
			zap.Error(err))
		return err
	}

	return nil
}

// updateNotionalCache updates the notional value of assets in cache.
func (j *NotionalJob) updateNotionalCache(notionals map[vaa.ChainID]NotionalCacheField) error {
	for chainID, notional := range notionals {
		key := fmt.Sprintf(NotionalCacheKey, chainID)
		err := j.cacheClient.Set(key, notional, 0).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// NotionalCacheField is the notional value of assets in cache.
type NotionalCacheField struct {
	NotionalUsd float64   `json:"notional_usd"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (n NotionalCacheField) MarshalBinary() ([]byte, error) {
	return json.Marshal(n)
}

// convertToWormholeChainIDs converts the coingecko chain ids to wormhole chain ids.
func convertToWormholeChainIDs(m map[string]coingecko.NotionalUSD) map[vaa.ChainID]NotionalCacheField {
	w := make(map[vaa.ChainID]NotionalCacheField, len(m))
	now := time.Now()
	for k, v := range m {
		switch k {
		case "solana":
			if v.Price != nil {
				w[vaa.ChainIDSolana] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "ethereum":
			if v.Price != nil {
				w[vaa.ChainIDEthereum] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "terra-luna":
			if v.Price != nil {
				w[vaa.ChainIDTerra] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "binancecoin":
			if v.Price != nil {
				w[vaa.ChainIDBSC] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "matic-network":
			if v.Price != nil {
				w[vaa.ChainIDPolygon] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "avalanche-2":
			if v.Price != nil {
				w[vaa.ChainIDAvalanche] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "oasis-network":
			if v.Price != nil {
				w[vaa.ChainIDOasis] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "algorand":
			if v.Price != nil {
				w[vaa.ChainIDAlgorand] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "aurora":
			if v.Price != nil {
				w[vaa.ChainIDAurora] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "fantom":
			if v.Price != nil {
				w[vaa.ChainIDFantom] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "karura":
			if v.Price != nil {
				w[vaa.ChainIDKarura] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "acala":
			if v.Price != nil {
				w[vaa.ChainIDAcala] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "klay-token":
			if v.Price != nil {
				w[vaa.ChainIDKlaytn] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "celo":
			if v.Price != nil {
				w[vaa.ChainIDCelo] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "near":
			if v.Price != nil {
				w[vaa.ChainIDNear] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "moonbeam":
			if v.Price != nil {
				w[vaa.ChainIDMoonbeam] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "neon":
			if v.Price != nil {
				w[vaa.ChainIDNeon] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "terra-luna-2":
			if v.Price != nil {
				w[vaa.ChainIDTerra2] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "injective-protocol":
			if v.Price != nil {
				w[vaa.ChainIDInjective] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "aptos":
			if v.Price != nil {
				w[vaa.ChainIDAptos] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "sui":
			if v.Price != nil {
				w[vaa.ChainIDSui] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "arbitrum":
			if v.Price != nil {
				w[vaa.ChainIDArbitrum] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "optimism":
			if v.Price != nil {
				w[vaa.ChainIDOptimism] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "xpla":
			if v.Price != nil {
				w[vaa.ChainIDXpla] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "bitcoin":
			if v.Price != nil {
				w[vaa.ChainIDBtc] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		case "base-protocol":
			if v.Price != nil {
				w[vaa.ChainIDBase] = NotionalCacheField{NotionalUsd: *v.Price, UpdatedAt: now}
			}
		}
	}
	return w
}
