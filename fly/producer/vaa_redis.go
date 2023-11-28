package producer

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

// RedisProducer represents a redis producer.
type RedisProducer struct {
	client  *redis.Client
	channel string
}

// NewRedisProducer returns a PushFunc that pushes NotificationEvent to redis.
func NewRedisProducer(c *redis.Client, channel string) *RedisProducer {

	return &RedisProducer{
		client:  c,
		channel: channel,
	}
}

// Push pushes a NotificationEvent to redis.
func (p *RedisProducer) Push(ctx context.Context, n *Notification) error {
	body, err := json.Marshal(n.Event)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, p.channel, string(body)).Err()
}
