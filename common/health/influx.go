package health

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/influx"
)

// InfluxOSS do a healtheckIN influx OSS version.
func InfluxOSS(client *influx.Client) Check {
	return func(ctx context.Context) error {
		return client.HealthcheckOSS(ctx)
	}
}

// Influx do a healtheck in influx Cloud version.
func Influx(client *influx.Client) Check {
	return func(ctx context.Context) error {
		return client.Healthcheck(ctx)
	}
}
