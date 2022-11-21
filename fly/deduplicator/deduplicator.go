package deduplicator

import (
	"context"
	"time"

	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"go.uber.org/zap"
)

// Option represents a deduplicator option function.
type Option func(*Deduplicator)

// Deduplicator represents a filter to avoid duplicate messages
type Deduplicator struct {
	cache      cache.CacheInterface[bool]
	logger     *zap.Logger
	expiration time.Duration
}

// New creates a deduplicator instance
func New(cache cache.CacheInterface[bool], logger *zap.Logger, opts ...Option) *Deduplicator {
	d := &Deduplicator{
		cache:      cache,
		expiration: 30 * time.Second,
		logger:     logger}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// WithExpiration allows to specify an expiration time when setting a value.
func WithExpiration(expiration time.Duration) Option {
	return func(d *Deduplicator) {
		d.expiration = expiration
	}
}

// Apply executes the fn function in case the message has not been received previously
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
