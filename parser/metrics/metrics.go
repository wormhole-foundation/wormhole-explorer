package metrics

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"go.uber.org/zap"
)

const measurement = "vaa_volume"

// Metric definition.
type Metrics struct {
	influxCli influxdb2.Client
	writeApi  api.WriteAPIBlocking
	logger    *zap.Logger
}

type Volume struct {
	ChainSourceID      uint16
	ChainDestinationID uint16
	Value              uint64
	Timestamp          time.Time
	AppID              string
}

// New create a new *Metric
func New(influxCli influxdb2.Client, organization, bucket string, logger *zap.Logger) *Metrics {
	writeAPI := influxCli.WriteAPIBlocking(organization, bucket)
	return &Metrics{influxCli: influxCli, writeApi: writeAPI, logger: logger}
}

func (m *Metrics) PushVolume(ctx context.Context, v *Volume) error {
	point := influxdb2.NewPointWithMeasurement(measurement).
		AddTag("chain_source_id", fmt.Sprintf("%d", v.ChainSourceID)).
		AddTag("chain_destination_id", fmt.Sprintf("%d", v.ChainDestinationID)).
		AddField("volume", v.Value).
		AddField("app_id", v.AppID).
		SetTime(v.Timestamp)

	// write point to influx
	err := m.writeApi.WritePoint(ctx, point)
	if err != nil {
		return err
	}
	return nil
}
