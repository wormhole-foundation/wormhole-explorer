package main

import (
	"context"
	"encoding/json"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/configuration"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/config"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/repository"
	"go.uber.org/zap"
	"log"
	"time"
)

func main() {

	// get the config
	cfg, errConf := configuration.LoadFromEnv[config.Configuration](context.Background())
	if errConf != nil {
		log.Fatal("error creating config", errConf)
	}

	logger := logger.New("wormhole-explorer-jobs", logger.WithLevel(cfg.LogLevel))
	logger.Info("started job execution backfiller", zap.String("job_id", cfg.JobID))

	if cfg.JobID == jobs.JobIDProtocolsStatsHourly {
		to := time.Now().UTC().Truncate(1 * time.Hour)
		totals := 24 * 29
		for i := 1; i <= totals; i++ {
			fmt.Println("execution:", i)
			from := to.Add(time.Duration(-1) * time.Hour)
			job := initProtocolStatsHourlyJob(context.Background(), logger, from, to)
			job.Run(context.Background())
			to = from
			time.Sleep(3 * time.Second)
		}
	}

	if cfg.JobID == jobs.JobIDProtocolsStatsDaily {
		to := time.Now().UTC().Truncate(24 * time.Hour)
		totals := 29
		for i := 1; i <= totals; i++ {
			fmt.Println("execution:", i)
			from := to.Add(time.Duration(-24) * time.Hour)
			job := initProtocolStatsDailyJob(context.Background(), logger, from, to)
			job.Run(context.Background())
			to = from
			time.Sleep(3 * time.Second)
		}
	}

}

func initProtocolStatsHourlyJob(ctx context.Context, logger *zap.Logger, from, to time.Time) *protocols.StatsJob {
	cfgJob, errCfg := configuration.LoadFromEnv[config.ProtocolsStatsConfiguration](ctx)
	if errCfg != nil {
		log.Fatal("error creating config", errCfg)
	}
	errUnmarshal := json.Unmarshal([]byte(cfgJob.ProtocolsJson), &cfgJob.Protocols)
	if errUnmarshal != nil {
		log.Fatal("error unmarshalling protocols config", errUnmarshal)
	}
	dbClient := influxdb2.NewClient(cfgJob.InfluxUrl, cfgJob.InfluxToken)
	dbWriter := dbClient.WriteAPIBlocking(cfgJob.InfluxOrganization, cfgJob.InfluxBucket30Days)

	protocolRepos := make([]repository.ProtocolRepository, 0, len(cfgJob.Protocols))
	for _, c := range cfgJob.Protocols {
		builder, ok := repository.ProtocolsRepositoryFactory[c.Name]
		if !ok {
			log.Fatal("error creating protocol stats client. Unknown protocol:", c.Name, errCfg)
		}
		cs := builder(c.Url, logger.With(zap.String("protocol", c.Name), zap.String("url", c.Url)))
		protocolRepos = append(protocolRepos, cs)
	}
	return protocols.NewStatsJob(dbWriter,
		from,
		to,
		dbconsts.ProtocolsActivityMeasurementHourly,
		dbconsts.ProtocolsStatsMeasurementHourly,
		protocolRepos,
		logger)
}

func initProtocolStatsDailyJob(ctx context.Context, logger *zap.Logger, from, to time.Time) *protocols.StatsJob {
	cfgJob, errCfg := configuration.LoadFromEnv[config.ProtocolsStatsConfiguration](ctx)
	if errCfg != nil {
		log.Fatal("error creating config", errCfg)
	}
	errUnmarshal := json.Unmarshal([]byte(cfgJob.ProtocolsJson), &cfgJob.Protocols)
	if errUnmarshal != nil {
		log.Fatal("error unmarshalling protocols config", errUnmarshal)
	}
	dbClient := influxdb2.NewClient(cfgJob.InfluxUrl, cfgJob.InfluxToken)
	dbWriter := dbClient.WriteAPIBlocking(cfgJob.InfluxOrganization, cfgJob.InfluxBucket30Days)

	protocolRepos := make([]repository.ProtocolRepository, 0, len(cfgJob.Protocols))
	for _, c := range cfgJob.Protocols {
		builder, ok := repository.ProtocolsRepositoryFactory[c.Name]
		if !ok {
			log.Fatal("error creating protocol stats client. Unknown protocol:", c.Name, errCfg)
		}
		cs := builder(c.Url, logger.With(zap.String("protocol", c.Name), zap.String("url", c.Url)))
		protocolRepos = append(protocolRepos, cs)
	}
	return protocols.NewStatsJob(dbWriter,
		from,
		to,
		dbconsts.ProtocolsActivityMeasurementDaily,
		dbconsts.ProtocolsStatsMeasurementDaily,
		protocolRepos,
		logger)
}
