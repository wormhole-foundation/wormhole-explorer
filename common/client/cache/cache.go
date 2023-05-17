// Package cache implement a simple cache redis client.
// It define a type [Cache] that represent the cache client and
// It define the methods Get to get a valur from a cache key.
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var (
	ErrCacheNotEnabled = errors.New("CACHE NOT ENABLED")
	ErrNotFound        = errors.New("KEY NOT FOUND IN CACHE")
	ErrInternal        = errors.New("INTERNAL CACHE ERROR")
)

// CacheClient redis cache client.
type CacheClient struct {
	Client  *redis.Client
	Enabled bool
	logger  *zap.Logger
}

// Cache is the interface for cache client.
type Cache interface {
	CacheReadable
	CacheWriteable
}

// CacheWriteable is the interface for write cache.
type CacheWriteable interface {
	Set(ctx context.Context, key string, value interface{}, expirations time.Duration) error
}

// CacheReadable is the interface for read cache.
type CacheReadable interface {
	Get(ctx context.Context, key string) (string, error)
	Close() error
}

type CacheGetFunc func(ctx context.Context, key string) (string, error)

// NewCacheClient init a new cache client.
func NewCacheClient(redisClient *redis.Client, enabled bool, log *zap.Logger) (*CacheClient, error) {
	if redisClient == nil {
		return nil, errors.New("redis client is nil")
	}
	return &CacheClient{Client: redisClient, Enabled: enabled, logger: log}, nil
}

// Get get a cache value or error from a key.
// If the cache is not enabled, the error value
// If the cache not contain a value from a key, the error value errors.ErrNotFound is returned.
// If exist some internal error in the cache, the error value errros.ErrInternalError is returned.
func (c *CacheClient) Get(ctx context.Context, key string) (string, error) {
	if !c.Enabled {
		return "", ErrCacheNotEnabled
	}
	value, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		if errors.Is(err, redis.Nil) {
			c.logger.Debug("key does not exist in cache",
				zap.Error(err), zap.String("key", key), zap.String("requestID", requestID))
			return "", ErrNotFound
		}
		c.logger.Error("error getting key from cache",
			zap.Error(err), zap.String("key", key), zap.String("requestID", requestID))
		return "", ErrInternal
	}
	return value, nil
}

// Close close the cache client.
func (c *CacheClient) Close() error {
	return c.Client.Close()
}

// Set set a value in cache.
func (c *CacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !c.Enabled {
		return ErrCacheNotEnabled
	}
	err := c.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		c.logger.Error("can not set key/value in cache",
			zap.Error(err),
			zap.String("key", key),
			zap.Any("value", value),
			zap.String("requestID", requestID))
		return err
	}
	return nil
}
