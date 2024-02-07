package builder

import (
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/wormhole-foundation/wormhole-explorer/fly/deduplicator"
	"go.uber.org/zap"
)

func NewCache[T any]() (cache.CacheInterface[T], error) {
	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000,          // Num keys to track frequency of (1000).
		MaxCost:     10 * (1 << 20), // Maximum cost of cache (10 MB).
		BufferItems: 64,             // Number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}
	store := store.NewRistretto(c)
	return cache.New[T](store), nil
}

func NewDeduplicator(logger *zap.Logger) (*deduplicator.Deduplicator, error) {
	// Creates a deduplicator to discard VAA messages that were processed previously
	deduplicatorCache, err := NewCache[bool]()
	if err != nil {
		return nil, err
	}
	return deduplicator.New(deduplicatorCache, logger), nil
}
