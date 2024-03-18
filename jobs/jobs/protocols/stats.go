package protocols

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/repository"
	"go.uber.org/zap"
	"sync"
	"time"
)

type StatsJob struct {
	writerDB               api.WriteAPIBlocking
	logger                 *zap.Logger
	repositories           []repository.ProtocolRepository
	from                   time.Time
	to                     time.Time
	destinationMeasurement string
	statsMeasurement       string
}

// NewStatsJob creates an instance of the job implementation.
func NewStatsJob(writerDB api.WriteAPIBlocking, from, to time.Time, activityMeasurement, statsMeasurement string, repositories []repository.ProtocolRepository, logger *zap.Logger) *StatsJob {
	return &StatsJob{
		writerDB:               writerDB,
		logger:                 logger.With(zap.String("module", "ProtocolsStatsJob")),
		repositories:           repositories,
		from:                   from,
		to:                     to,
		destinationMeasurement: activityMeasurement,
		statsMeasurement:       statsMeasurement,
	}
}

func (s *StatsJob) Run(ctx context.Context) error {

	wg := sync.WaitGroup{}
	wg.Add(len(s.repositories))

	s.logger.Info("running protocols stats job", zap.Time("from", s.from), zap.Time("to", s.to))

	for _, repo := range s.repositories {
		go s.processProtocol(ctx, repo, &wg)
	}
	wg.Wait()
	return nil
}

func (s *StatsJob) processProtocol(ctx context.Context, protocolRepo repository.ProtocolRepository, wg *sync.WaitGroup) {
	defer wg.Done()

	var stats repository.Stats
	var errStats error
	wgStats := sync.WaitGroup{}
	wgStats.Add(1)
	go func() {
		defer wgStats.Done()
		stats, errStats = protocolRepo.GetStats(ctx)
	}()

	activity, errAct := protocolRepo.GetActivity(ctx, s.from, s.to)
	if errAct != nil {
		s.logger.Error("failed to get protocol activity", zap.Error(errAct), zap.String("protocol", protocolRepo.ProtocolName()), zap.Time("from", s.from), zap.Time("to", s.to))
		return
	}

	wgStats.Wait()
	if errStats != nil {
		s.logger.Error("failed to get protocol stats", zap.Error(errStats), zap.String("protocol", protocolRepo.ProtocolName()))
		return
	}

	data := protocolData{
		Stats:    stats,
		Activity: activity,
	}

	errAct = s.updateActivity(ctx, protocolRepo.ProtocolName(), data.Activity, s.from)
	if errAct != nil {
		s.logger.Error("failed updating protocol activities in influxdb", zap.Error(errAct), zap.String("protocol", protocolRepo.ProtocolName()))
	}

	errStats = s.updateStats(ctx, protocolRepo.ProtocolName(), data.Stats, s.to)
	if errStats != nil {
		s.logger.Error("failed updating protocol stats in influxdb", zap.Error(errStats), zap.String("protocol", protocolRepo.ProtocolName()))
	}

}

type protocolData struct {
	Stats    repository.Stats
	Activity repository.ProtocolActivity
}

func (s *StatsJob) updateActivity(ctx context.Context, protocol string, data repository.ProtocolActivity, ts time.Time) error {

	txs := uint64(0)
	totalUsd := float64(0)
	for _, act := range data.Activities {
		txs += act.Txs
		totalUsd += act.TotalUSD
	}

	point := influxdb2.NewPointWithMeasurement(s.destinationMeasurement).
		AddTag("protocol", protocol).
		AddField("total_value_secure", data.TotalValueSecure).
		AddField("total_value_transferred", data.TotalValueTransferred).
		AddField("volume", data.Volume).
		AddField("txs", txs).
		AddField("total_usd", totalUsd).
		SetTime(ts)

	err := s.writerDB.WritePoint(ctx, point)
	if err != nil {
		s.logger.Error("failed updating protocol Activities in influxdb", zap.Error(err), zap.String("protocol", protocol))
	}
	return err
}

func (s *StatsJob) updateStats(ctx context.Context, protocol string, data repository.Stats, ts time.Time) error {

	point := influxdb2.
		NewPointWithMeasurement(s.statsMeasurement).
		AddTag("protocol", protocol).
		AddField("total_messages", data.TotalMessages).
		AddField("total_value_locked", data.TotalValueLocked).
		AddField("volume", data.Volume).
		SetTime(ts)

	err := s.writerDB.WritePoint(ctx, point)
	if err != nil {
		s.logger.Error("failed updating protocol stats in influxdb", zap.Error(err), zap.String("protocol", protocol))
	}
	return err
}
