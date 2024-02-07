package builder

import (
	"github.com/go-redis/redis/v8"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
)

func NewRedisClient(cfg *config.Configuration) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: cfg.Redis.RedisUri})
}
