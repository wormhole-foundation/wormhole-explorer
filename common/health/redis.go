package health

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// Redis does a ping.
func Redis(client *redis.Client) Check {
	return func(ctx context.Context) error {
		result := client.Ping(ctx)
		return result.Err()
	}
}
