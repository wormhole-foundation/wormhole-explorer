package notional

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"go.uber.org/zap"
)

const (
	wormscanNotionalUpdated       = "NOTIONAL_UPDATED"
	wormscanNotionalCacheKeyRegex = "WORMSCAN:NOTIONAL:SYMBOL:*"
	KeyFormatString               = "WORMSCAN:NOTIONAL:SYMBOL:%s"
)

var (
	ErrNotFound          = errors.New("NOT FOUND")
	ErrInvalidCacheField = errors.New("INVALID CACHE FIELD")
)

// NotionalLocalCacheReadable is the interface for notional local cache.
type NotionalLocalCacheReadable interface {
	Get(symbol domain.Symbol) (PriceData, error)
	Close() error
}

// PriceData is the notional value of assets in cache.
type PriceData struct {
	NotionalUsd decimal.Decimal `json:"notional_usd"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
//
// This function is used when the notional job writes data to redis.
func (p PriceData) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

// NotionalCacheClient redis cache client.
type NotionalCache struct {
	client      *redis.Client
	pubSub      *redis.PubSub
	channel     string
	notionalMap sync.Map
	prefix      string
	logger      *zap.Logger
}

// NewNotionalCache create a new cache client.
// After create a NotionalCache use the Init method to initialize pubsub and load the cache.
func NewNotionalCache(ctx context.Context, redisClient *redis.Client, prefix string, channel string, log *zap.Logger) (*NotionalCache, error) {
	if redisClient == nil {
		return nil, errors.New("redis client is nil")
	}

	pubsub := redisClient.Subscribe(ctx, channel)
	return &NotionalCache{
		client:      redisClient,
		pubSub:      pubsub,
		channel:     channel,
		notionalMap: sync.Map{},
		prefix:      prefix,
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

	var cursor uint64
	var err error
	for {
		// Get a page of results from the cursor
		var keys []string
		scanCmd := c.client.Scan(ctx, cursor, c.renderRegExp(), 100)
		if scanCmd.Err() != nil {
			c.logger.Error("redis.ScanCmd has errors", zap.Error(err))
			return fmt.Errorf("redis.ScanCmd has errors: %w", err)
		}
		keys, cursor, err = scanCmd.Result()
		if err != nil {
			c.logger.Error("call to redis.ScanCmd.Result() failed", zap.Error(err))
			return fmt.Errorf("call to redis.ScanCmd.Result() failed: %w", err)
		}

		// Get notional value from keys
		for _, key := range keys {
			var field PriceData
			value, err := c.client.Get(ctx, key).Result()
			json.Unmarshal([]byte(value), &field)
			if err != nil {
				c.logger.Error("loadCache", zap.Error(err))
				return err
			}
			// Save notional value to local cache
			c.notionalMap.Store(key, field)
		}

		// If we've reached the end of the cursor, return
		if cursor == 0 {
			return nil
		}
	}
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
func (c *NotionalCache) Get(symbol domain.Symbol) (PriceData, error) {
	var notional PriceData

	// get notional cache key
	key := fmt.Sprintf(KeyFormatString, symbol)
	key = c.renderKey(key)

	// get notional cache value
	field, ok := c.notionalMap.Load(key)
	if !ok {
		return notional, ErrNotFound
	}

	// convert any field to NotionalCacheField
	notional, ok = field.(PriceData)
	if !ok {
		c.logger.Error("invalid notional cache field",
			zap.Any("field", field),
			zap.String("symbol", symbol.String()))
		return notional, ErrInvalidCacheField
	}
	return notional, nil
}

func (c *NotionalCache) renderKey(key string) string {
	if c.prefix != "" {
		return fmt.Sprintf("%s:%s", c.prefix, key)
	} else {
		return key
	}
}

func (c *NotionalCache) renderRegExp() string {
	return "*" + c.renderKey(wormscanNotionalCacheKeyRegex)
}
