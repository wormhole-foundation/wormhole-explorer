package builder

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/metrics"
	"github.com/eko/gocache/v3/store"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/deduplicator"
	"go.uber.org/zap"
)

func NewCache[T any](name string, numKeys, maxCost int64) (cache.CacheInterface[T], error) {
	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: numKeys,             // Num keys to track frequency of (1000).
		MaxCost:     maxCost * (1 << 20), // Maximum cost of cache (10 MB).
		BufferItems: 64,                  // Number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}
	store := store.NewRistretto(c)
	return cache.NewMetric[T](
		metrics.NewPrometheus(name),
		cache.New[T](store),
	), nil
}

func NewDeduplicator(name string, cfg config.Cache, logger *zap.Logger) (*deduplicator.Deduplicator, error) {
	// Creates a deduplicator to discard VAA messages that were processed previously
	deduplicatorCache, err := NewCache[bool](name, cfg.NumKeys, cfg.MaxCostsInMB)
	if err != nil {
		return nil, err
	}
	expiration := time.Duration(cfg.ExpirationInSeconds) * time.Second
	return deduplicator.New(deduplicatorCache, logger, deduplicator.WithExpiration(expiration)), nil
}
