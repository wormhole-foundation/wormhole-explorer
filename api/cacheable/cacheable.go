package cacheable

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"
)

func GetOrLoad[T any](
	ctx context.Context,
	logger *zap.Logger,
	cacheClient cache.Cache,
	expirations time.Duration,
	key string,
	load func() (T, error),
) (T, error) {
	log := logger.With(zap.String("key", key))
	value, err := cacheClient.Get(ctx, key)
	foundCache := true
	if err != nil {
		foundCache = false
		if err != cache.ErrNotFound {
			log.Warn("getting result from cache", zap.Error(err))
		}
	}
	var cached CachedResult[T]
	if foundCache {
		err = json.Unmarshal([]byte(value), &cached)
		if err != nil {
			log.Warn("unmarshal cache", zap.Error(err))
		} else if cached.Timestamp.Add(expirations).After(time.Now()) {
			return cached.Result, nil
		}
	}
	result, err := load()
	if err != nil {
		if foundCache {
			log.Warn("load function fails but returns cached result",
				zap.Error(err), zap.String("cacheTime", cached.Timestamp.String()))
			return cached.Result, nil
		}
		return result, err
	}
	newValue := CachedResult[T]{Timestamp: time.Now(), Result: result}
	err = cacheClient.Set(ctx, key, newValue, 10*expirations)
	if err != nil {
		log.Warn("saving the result in the cache", zap.Error(err))
	}
	return result, nil
}

type CachedResult[T any] struct {
	Timestamp time.Time `json:"timestamp"`
	Result    T         `json:"result"`
}

func (c CachedResult[T]) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}
