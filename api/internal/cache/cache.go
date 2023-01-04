// Package cache implement a simple cache redis client.
// It define a type [Cache] that represent the cache client and
// It define the methods Get to get a valur from a cache key.
package cache

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"go.uber.org/zap"
)

var ErrCacheNotEnabled = errors.New("CACHE NOT ENABLED")

// CacheClient redis cache client.
type CacheClient struct {
	Client  *redis.Client
	Prefix  string
	Enabled bool
	logger  *zap.Logger
}

// NewCacheClient init a new cache client.
func NewCacheClient(url string, enabled bool, log *zap.Logger) *CacheClient {
	client := redis.NewClient(
		&redis.Options{
			Addr: url,
		})
	return &CacheClient{Client: client, Prefix: "wormscan", Enabled: enabled, logger: log}
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
		if err != redis.Nil {
			requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
			c.logger.Error("ket does not exist in cache",
				zap.Error(err), zap.String("key", key), zap.String("requestID", requestID))
			return "", errs.ErrNotFound
		}
		return "", errs.ErrInternalError
	}
	return value, nil
}
