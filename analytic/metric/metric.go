package metric

import (
	"context"

	influx "github.com/wormhole-foundation/wormhole-explorer/common/client/influx"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// Metric definition.
type Metric struct {
	influxCli *influx.Client
}

// New create a new *Metric
func New(influxCli *influx.Client) *Metric {
	return &Metric{influxCli: influxCli}
}

// Push implement MetricPushFunc definition
func (m *Metric) Push(ctx context.Context, vaa *vaa.VAA) {
	// TODO
	return
}
