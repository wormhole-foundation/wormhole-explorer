package metric

import (
	"context"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Metric definition.
type Metric struct {
	influxCli influxdb2.Client
	writeApi  api.WriteAPIBlocking
	logger    *zap.Logger
}

// New create a new *Metric.
func New(influxCli influxdb2.Client, organization, bucket string, logger *zap.Logger) *Metric {
	writeAPI := influxCli.WriteAPIBlocking(organization, bucket)
	return &Metric{influxCli: influxCli, writeApi: writeAPI, logger: logger}
}

// Push implement MetricPushFunc definition.
func (m *Metric) Push(ctx context.Context, vaa *vaa.VAA) error {
	return m.vaaCountMeasurement(ctx, vaa)
}

// Close influx client.
func (m *Metric) Close() {
	m.influxCli.Close()
}

// vaaCountMeasurement handle the push of metric point for measurement vaa_count.
func (m *Metric) vaaCountMeasurement(ctx context.Context, vaa *vaa.VAA) error {

	// Create a measurement for all messages (including Pyth)
	{
		measurement := "vaa_count_including_pyth"

		// Create a new point for the `vaa_count` measurement.
		point := influxdb2.
			NewPointWithMeasurement(measurement).
			AddTag("chain_id", strconv.Itoa(int(vaa.EmitterChain))).
			AddField("count", 1).
			SetTime(vaa.Timestamp.Add(time.Nanosecond * time.Duration(vaa.Sequence)))

		// Write the point to influx.
		err := m.writeApi.WritePoint(ctx, point)
		if err != nil {
			m.logger.Error("failed to write metric",
				zap.String("measurement", measurement),
				zap.Uint16("chain_id", uint16(vaa.EmitterChain)),
				zap.Error(err),
			)
			return err
		}
	}

	// Create a measurement for non-pyth messages only
	if vaa.EmitterChain != sdk.ChainIDPythNet {
		measurement := "vaa_count"

		// Create a new point for the `vaa_count` measurement.
		point := influxdb2.
			NewPointWithMeasurement(measurement).
			AddTag("chain_id", strconv.Itoa(int(vaa.EmitterChain))).
			AddField("count", 1).
			SetTime(vaa.Timestamp.Add(time.Nanosecond * time.Duration(vaa.Sequence)))

		// Write the point to influx.
		err := m.writeApi.WritePoint(ctx, point)
		if err != nil {
			m.logger.Error("failed to write metric",
				zap.String("measurement", measurement),
				zap.Uint16("chain_id", uint16(vaa.EmitterChain)),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}
