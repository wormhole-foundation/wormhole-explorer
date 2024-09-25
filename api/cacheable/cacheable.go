package cacheable

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"
)

// GetOrLoad is a function that tries to get the result from the cache, if it is not found or it is expired, then it loads the result.
func GetOrLoad[T any](
	ctx context.Context,
	logger *zap.Logger,
	cacheClient cache.Cache,
	expirations time.Duration,
	key string,
	metrics metrics.Metrics,
	load func() (T, error),
	automaticRenew bool,
) (T, error) {
	log := logger.With(zap.String("key", key))

	// Try to get the result from the cache.
	value, err := cacheClient.Get(ctx, key)
	foundCache := true

	//If the result is not found in the cache or fails, then load the result.
	if err != nil {
		foundCache = false
		if err != cache.ErrNotFound {
			log.Warn("getting result from cache", zap.Error(err))
		}
	}

	var cached CachedResult[T]
	//If the result is found in the cache and it is not expired, then return the result.
	if foundCache {
		err = json.Unmarshal([]byte(value), &cached)
		if err != nil {
			log.Warn("unmarshal cache", zap.Error(err))
		} else if !cached.IsExpired(expirations) {
			return cached.Result, nil
		} else if automaticRenew && !cached.RenewInProgress {
			go renewCachedValue(context.WithoutCancel(ctx), cacheClient, log, key, &cached, load)
			return cached.Result, nil
		}
	}

	//If the result is not found in the cache or it is expired, then load the result.
	result, err := load()
	if err != nil {
		//If the load function fails and the cache was found and is expired, the cache value is returned anyway.
		if foundCache {
			metrics.IncExpiredCacheResponse(key)
			log.Warn("load function fails but returns cached result",
				zap.Error(err), zap.String("cacheTime", cached.Timestamp.String()))
			return cached.Result, nil
		}
		return result, err
	}

	//Saves the result of the execution of the load function in cache.
	newValue := CachedResult[T]{Timestamp: time.Now(), Result: result}
	err = cacheClient.Set(ctx, key, newValue, 0)
	if err != nil {
		log.Warn("saving the result in the cache", zap.Error(err))
	}

	//Returns the result of the execution of the function load
	return result, nil
}

func renewCachedValue[T any](
	ctx context.Context,
	cacheClient cache.Cache,
	log *zap.Logger,
	key string,
	cached *CachedResult[T],
	load func() (T, error),
) {
	cached.RenewInProgress = true
	err := cacheClient.Set(ctx, key, cached, 0) // save the renewInProgress
	if err != nil {
		log.Error("error updating cache value state renewInProgress", zap.Error(err))
		return
	}
	newValue, errNewValue := load()
	if errNewValue != nil {
		log.Error("error renewing cache value", zap.Error(errNewValue))
		cleanRenewInProgress(cacheClient, key, cached)
		return
	}
	renewedCacheValue := CachedResult[T]{Timestamp: time.Now(), Result: newValue, RenewInProgress: false}
	err = cacheClient.Set(ctx, key, renewedCacheValue, 0)
	if err != nil {
		log.Error("error updating cache value", zap.Error(err))
		cleanRenewInProgress(cacheClient, key, cached)
		return
	}
}

func cleanRenewInProgress[T any](cacheClient cache.Cache, key string, cached *CachedResult[T]) {

	cached.RenewInProgress = false
	err := cacheClient.Set(context.Background(), key, cached, 0)

	for err != nil {
		select {
		case <-time.After(100 * time.Millisecond):
			err = cacheClient.Set(context.Background(), key, cached, 0)
		}
	}

}

type CachedResult[T any] struct {
	Timestamp       time.Time `json:"timestamp"`
	Result          T         `json:"result"`
	RenewInProgress bool      `json:"renew_in_progress"`
}

func (c CachedResult[T]) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

func (c CachedResult[T]) IsExpired(ttl time.Duration) bool {
	return c.Timestamp.Add(ttl).After(time.Now())
}
