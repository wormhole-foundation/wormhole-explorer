package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/wormhole-foundation/wormhole-explorer/common/events"
	"go.uber.org/zap"
)

// RedisSubscriber is a redis subscriber.
type RedisSubscriber struct {
	client  *redis.Client
	pubSub  *redis.PubSub
	logger  *zap.Logger
	handler WatcherFunc
}

// NewRedisSubscriber create a new redis subscriber.
func NewRedisSubscriber(ctx context.Context, redisClient *redis.Client, prefix, channel string, handler WatcherFunc, log *zap.Logger) (*RedisSubscriber, error) {
	if redisClient == nil {
		return nil, errors.New("redis client is nil")
	}

	channel = fmt.Sprintf("%s:%s", prefix, channel)
	pubsub := redisClient.Subscribe(ctx, channel)
	return &RedisSubscriber{
		client:  redisClient,
		pubSub:  pubsub,
		handler: handler,
		logger:  log}, nil
}

// Start executes database event consumption.
func (w *RedisSubscriber) Start(ctx context.Context) error {
	w.subscribe(ctx)
	return nil
}

// Close closes the redis event consumption.
func (w *RedisSubscriber) Close(ctx context.Context) error {
	return w.pubSub.Close()
}

func (r *RedisSubscriber) subscribe(ctx context.Context) {
	ch := r.pubSub.Channel()
	go func() {
		for msg := range ch {
			var notification events.NotificationEvent
			err := json.Unmarshal([]byte(msg.Payload), &notification)
			if err != nil {
				r.logger.Error("Error decoding vaaEvent message from SQSEvent", zap.Error(err))
				continue
			}

			switch notification.Event {
			case events.SignedVaaType:
				signedVaa, err := events.GetEventData[events.SignedVaa](&notification)
				if err != nil {
					r.logger.Error("Error decoding signedVAA from notification event", zap.String("trackId", notification.TrackID), zap.Error(err))
					continue
				}
				r.handler(&Event{
					ID:   signedVaa.ID,
					Vaas: signedVaa.Vaa,
				})
			default:
				continue
			}
		}
	}()
}
