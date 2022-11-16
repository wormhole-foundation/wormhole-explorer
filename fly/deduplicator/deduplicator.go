package deduplicator

import (
	"context"
	"time"

	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"go.uber.org/zap"
)

type DeduplicatorOption func(*Deduplicator)

type Deduplicator struct {
	cache      cache.CacheInterface[bool]
	logger     *zap.Logger
	expiration time.Duration
}

func New(cache cache.CacheInterface[bool], logger *zap.Logger, opts ...DeduplicatorOption) *Deduplicator {
	d := &Deduplicator{
		cache:      cache,
		expiration: 30 * time.Second,
		logger:     logger}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func WithExpiration(expiration time.Duration) DeduplicatorOption {
	return func(d *Deduplicator) {
		d.expiration = expiration
	}
}

func (d *Deduplicator) Apply(ctx context.Context, key string, fn func() error) error {
	if v, _ := d.cache.Get(ctx, key); v {
		return nil
	}

	if err := fn(); err != nil {
		return err
	}

	_ = d.cache.Set(ctx, key, true, store.WithCost(16), store.WithExpiration(d.expiration))

	return nil
}
