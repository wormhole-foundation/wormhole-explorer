package queue

import (
	"context"
	"fly/storage"
	"time"

	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type VAADeduplication struct {
	repository *storage.Repository
	logger     *zap.Logger
	cache      cache.CacheInterface[bool]
	expiration time.Duration
}

func NewVAADeduplication(repository *storage.Repository, cache cache.CacheInterface[bool], logger *zap.Logger) *VAADeduplication {
	return &VAADeduplication{
		repository: repository,
		cache:      cache,
		expiration: 30 * time.Second,
		logger:     logger}
}

func (d *VAADeduplication) Publish(ctx context.Context, v *vaa.VAA, data []byte) error {
	if v, _ := d.cache.Get(ctx, v.MessageID()); v {
		return nil
	}

	if err := d.repository.UpsertVaa(v, data); err != nil {
		return err
	}

	_ = d.cache.Set(ctx, v.MessageID(), true, store.WithCost(16), store.WithExpiration(d.expiration))

	return nil
}
