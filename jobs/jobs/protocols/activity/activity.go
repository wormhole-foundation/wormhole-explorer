package activity

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/activity/internal/repositories"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

// NewProtocolActivityJob creates an instance of the job implementation.
func NewProtocolActivityJob(statsDB api.WriteAPIBlocking, logger *zap.Logger, version string, activityFetchers ...ClientActivity) *ProtocolsActivityJob {
	return &ProtocolsActivityJob{
		statsDB:          statsDB,
		logger:           logger.With(zap.String("module", "ProtocolsActivityJob")),
		activityFetchers: activityFetchers,
		version:          version,
	}
}

func (m *ProtocolsActivityJob) Run(ctx context.Context) error {

	clientsQty := len(m.activityFetchers)
	wg := sync.WaitGroup{}
	wg.Add(clientsQty)
	errs := make(chan error, clientsQty)
	ts := time.Now().UTC().Truncate(time.Hour) // make minutes and seconds zero, so we only work with date and hour
	from := ts.Add(-1 * time.Hour)
	m.logger.Info("running protocols activity job ", zap.Time("from", from), zap.Time("to", ts))
	for _, cs := range m.activityFetchers {
		go func(c ClientActivity) {
			defer wg.Done()
			activity, err := c.Get(ctx, from, ts)
			if err != nil {
				errs <- err
				return
			}
			errs <- m.updateActivity(ctx, c.ProtocolName(), m.version, activity, from)
		}(cs)
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *ProtocolsActivityJob) updateActivity(ctx context.Context, protocol, version string, activity repositories.ProtocolActivity, ts time.Time) error {

	points := make([]*write.Point, 0, len(activity.Activities))

	for i := range activity.Activities {
		point := influxdb2.
			NewPointWithMeasurement(dbconsts.ProtocolsActivityMeasurement).
			AddTag("protocol", protocol).
			AddTag("emitter_chain_id", strconv.FormatUint(activity.Activities[i].EmitterChainID, 10)).
			AddTag("destination_chain_id", strconv.FormatUint(activity.Activities[i].DestinationChainID, 10)).
			AddTag("version", version).
			AddField("total_value_secure", activity.TotalValueSecure).
			AddField("total_value_transferred", activity.TotalValueTransferred).
			AddField("txs", activity.Activities[i].Txs).
			AddField("total_usd", activity.Activities[i].TotalUSD).
			SetTime(ts)
		points = append(points, point)
	}

	err := m.statsDB.WritePoint(ctx, points...)
	if err != nil {
		m.logger.Error("failed updating protocol Activities in influxdb", zap.Error(err), zap.String("protocol", protocol))
	}
	return err
}
