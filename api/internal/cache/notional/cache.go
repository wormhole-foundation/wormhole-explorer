package notional

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	wormscanNotionalCacheKeyRegex = "*WORMSCAN:NOTIONAL:CHAIN_ID:*"
	wormscanNotionalUpdated       = "NOTIONAL_UPDATED"
)

var (
	ErrNotFound          = errors.New("NOT FOUND")
	ErrInvalidCacheField = errors.New("INVALID CACHE FIELD")
)

// NotionalLocalCacheReadable is the interface for notional local cache.
type NotionalLocalCacheReadable interface {
	Get(chainID vaa.ChainID) (NotionalCacheField, error)
	Close() error
}

// NotionalCacheField is the notional value of assets in cache.
type NotionalCacheField struct {
	NotionalUsd float64   `json:"notional_usd"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NotionalCacheClient redis cache client.
type NotionalCache struct {
	client      *redis.Client
	pubSub      *redis.PubSub
	channel     string
	notionalMap sync.Map
	logger      *zap.Logger
}

// NewNotionalCache create a new cache client.
// After create a NotionalCache use the Init method to initialize pubsub and load the cache.
func NewNotionalCache(ctx context.Context, redisClient *redis.Client, channel string, log *zap.Logger) (*NotionalCache, error) {
	if redisClient == nil {
		return nil, errors.New("redis client is nil")
	}

	pubsub := redisClient.Subscribe(ctx, channel)
	return &NotionalCache{
		client:      redisClient,
		pubSub:      pubsub,
		channel:     channel,
		notionalMap: sync.Map{},
		logger:      log}, nil
}

// Init subscribe to notional pubsub and load the cache.
func (c *NotionalCache) Init(ctx context.Context) error {
	// load notional cache
	err := c.loadCache(ctx)
	if err != nil {
		return err
	}

	// notional cache updated channel subscribe
	c.subscribe(ctx)

	return nil
}

// loadCache load notional cache from redis.
func (c *NotionalCache) loadCache(ctx context.Context) error {
	scanCom := c.client.Scan(ctx, 0, wormscanNotionalCacheKeyRegex, 100)
	for {
		// Scan for notional keys
		keys, cursor, err := scanCom.Result()
		if err != nil {
			c.logger.Error("loadCache", zap.Error(err))
			return err
		}

		// Get notional value from keys
		for _, key := range keys {
			var field NotionalCacheField
			value, err := c.client.Get(ctx, key).Result()
			json.Unmarshal([]byte(value), &field)
			if err != nil {
				c.logger.Error("loadCache", zap.Error(err))
				return err
			}
			// Save notional value to local cache
			c.notionalMap.Store(key, field)
		}

		if cursor == 0 {
			break
		}
	}
	return nil
}

// Subscribe to a notional update channel and load new values for the notional cache.
func (c *NotionalCache) subscribe(ctx context.Context) {
	ch := c.pubSub.Channel()

	go func() {
		for msg := range ch {
			if wormscanNotionalUpdated == msg.Payload {
				// update notional cache
				c.loadCache(ctx)
			}
		}
	}()
}

// Close the pubsub channel.
func (c *NotionalCache) Close() error {
	return c.pubSub.Close()
}

// Get notional cache value.
func (c *NotionalCache) Get(chainID vaa.ChainID) (NotionalCacheField, error) {
	var notional NotionalCacheField

	// get notional cache key
	key := fmt.Sprintf("WORMSCAN:NOTIONAL:CHAIN_ID:%d", chainID)

	// get notional cache value
	field, ok := c.notionalMap.Load(key)
	if !ok {
		return notional, ErrNotFound
	}

	// convert any field to NotionalCacheField
	notional, ok = field.(NotionalCacheField)
	if !ok {
		c.logger.Error("invalid notional cache field",
			zap.Any("field", field),
			zap.Any("chainID", chainID))
		return notional, ErrInvalidCacheField
	}
	return notional, nil
}
