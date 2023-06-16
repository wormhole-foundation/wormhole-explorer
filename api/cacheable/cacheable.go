package cacheable

import (
	"context"
	"encoding/json"
	"time"

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
	load func() (T, error),
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
		} else if cached.Timestamp.Add(expirations).After(time.Now()) {
			return cached.Result, nil
		}
	}

	//If the result is not found in the cache or it is expired, then load the result.
	result, err := load()
	if err != nil {
		//If the load function fails and the cache was found and is expired, the cache value is returned anyway.
		if foundCache {
			log.Warn("load function fails but returns cached result",
				zap.Error(err), zap.String("cacheTime", cached.Timestamp.String()))
			return cached.Result, nil
		}
		return result, err
	}

	//Saves the result of the execution of the load function in cache.
	newValue := CachedResult[T]{Timestamp: time.Now(), Result: result}
	err = cacheClient.Set(ctx, key, newValue, 10*expirations)
	if err != nil {
		log.Warn("saving the result in the cache", zap.Error(err))
	}

	//Returns the result of the execution of the function load
	return result, nil
}

type CachedResult[T any] struct {
	Timestamp time.Time `json:"timestamp"`
	Result    T         `json:"result"`
}

func (c CachedResult[T]) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}
