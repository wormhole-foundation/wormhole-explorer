package health

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/influx"
)

func Influx(client *influx.Client) Check {
	return func(ctx context.Context) error {
		return client.Healthcheck(ctx)
	}
}
