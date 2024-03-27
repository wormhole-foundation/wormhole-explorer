package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/repository"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/configuration"

	"github.com/go-redis/redis"
	txtrackerProcessVaa "github.com/wormhole-foundation/wormhole-explorer/common/client/txtracker"
	common "github.com/wormhole-foundation/wormhole-explorer/common/coingecko"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	filePrices "github.com/wormhole-foundation/wormhole-explorer/common/prices"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/config"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/internal/coingecko"
	apiPrices "github.com/wormhole-foundation/wormhole-explorer/jobs/internal/prices"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/migration"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/notional"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/report"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type exitCode int

func main() {
	defer handleExit()
	ctx := context.Background()

	// get the config
	cfg, errConf := configuration.LoadFromEnv[config.Configuration](ctx)
	if errConf != nil {
		log.Fatal("error creating config", errConf)
	}

	logger := logger.New("wormhole-explorer-jobs", logger.WithLevel(cfg.LogLevel))
	logger.Info("started job execution", zap.String("job_id", cfg.JobID))

	var err error
	switch cfg.JobID {
	case jobs.JobIDNotional:
		nCfg, errCfg := configuration.LoadFromEnv[config.NotionalConfiguration](ctx)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}
		notionalJob := initNotionalJob(ctx, nCfg, logger)
		err = notionalJob.Run()

	case jobs.JobIDTransferReport:
		aCfg, errCfg := configuration.LoadFromEnv[config.TransferReportConfiguration](ctx)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}
		transferReport := initTransferReportJob(ctx, aCfg, logger)
		err = transferReport.Run(ctx)

	case jobs.JobIDHistoricalPrices:
		hCfg, errCfg := configuration.LoadFromEnv[config.HistoricalPricesConfiguration](ctx)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}
		historyPrices := initHistoricalPricesJob(ctx, hCfg, logger)
		err = historyPrices.Run(ctx)

	case jobs.JobIDMigrationSourceTx:
		mCfg, errCfg := configuration.LoadFromEnv[config.MigrateSourceTxConfiguration](ctx)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}

		chainID := sdk.ChainID(mCfg.ChainID)
		migrationJob := initMigrateSourceTxJob(ctx, mCfg, chainID, logger)
		err = migrationJob.Run(ctx)

	case jobs.JobIDProtocolsStatsHourly:
		statsJob := initProtocolStatsHourlyJob(ctx, logger)
		err = statsJob.Run(ctx)
	case jobs.JobIDProtocolsStatsDaily:
		statsJob := initProtocolStatsDailyJob(ctx, logger)
		err = statsJob.Run(ctx)
	default:
		logger.Fatal("Invalid job id", zap.String("job_id", cfg.JobID))
	}

	if err != nil {
		logger.Error("failed job execution", zap.String("job_id", cfg.JobID), zap.Error(err))
	} else {
		logger.Info("finish job execution successfully", zap.String("job_id", cfg.JobID))
	}

}

// initNotionalJob initializes notional job.
func initNotionalJob(_ context.Context, cfg *config.NotionalConfiguration, logger *zap.Logger) *notional.NotionalJob {
	// init coingecko api client.
	api := coingecko.NewCoingeckoAPI(cfg.CoingeckoURL, cfg.CoingeckoHeaderKey, cfg.CoingeckoApiKey, logger)
	// init redis client.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.CacheURL})
	// init token provider.
	tokenProvider := domain.NewTokenProvider(cfg.P2pNetwork)
	notify := notional.NoopNotifier()
	// create notional job.
	notionalJob := notional.NewNotionalJob(api, redisClient, cfg.CachePrefix, cfg.NotionalChannel, tokenProvider, notify, logger)
	return notionalJob
}

// initTransferReportJob initializes transfer report job.
func initTransferReportJob(ctx context.Context, cfg *config.TransferReportConfiguration, logger *zap.Logger) *report.TransferReportJob {
	//setup DB connection
	db, err := dbutil.Connect(ctx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
	if err != nil {
		logger.Fatal("Failed to connect MongoDB", zap.Error(err))
	}
	var getPriceByTime report.GetPriceByTimeFn
	switch strings.ToLower(cfg.PricesType) {
	case "file":
		pricesCache := filePrices.NewCoinPricesCache(cfg.PricesUri)
		pricesCache.InitCache()
		getPriceByTime = pricesCache.GetPriceByTime
	case "api":
		api := apiPrices.NewPricesApi(cfg.PricesUri, logger)
		getPriceByTime = api.GetPriceByTime

	default:
		logger.Fatal("Invalid prices type", zap.String("prices_type", cfg.PricesType))
	}

	// init token provider.
	tokenProvider := domain.NewTokenProvider(cfg.P2pNetwork)
	return report.NewTransferReportJob(db.Database, cfg.PageSize, getPriceByTime, cfg.OutputPath, tokenProvider, logger)
}

func initHistoricalPricesJob(ctx context.Context, cfg *config.HistoricalPricesConfiguration, logger *zap.Logger) *notional.HistoryNotionalJob {
	//setup DB connection
	db, err := dbutil.Connect(ctx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
	if err != nil {
		logger.Fatal("Failed to connect MongoDB", zap.Error(err))
	}
	// init coingecko api client.
	api := common.NewCoinGeckoAPI(cfg.CoingeckoURL, cfg.CoingeckoHeaderKey, cfg.CoingeckoApiKey)
	// create history notional job.
	notionalJob := notional.NewHistoryNotionalJob(api, db.Database, cfg.P2pNetwork, cfg.RequestLimitTimeSeconds, cfg.PriceDays, logger)
	return notionalJob
}

func initMigrateSourceTxJob(ctx context.Context, cfg *config.MigrateSourceTxConfiguration, chainID sdk.ChainID, logger *zap.Logger) *migration.MigrateSourceChainTx {
	//setup DB connection
	db, err := dbutil.Connect(ctx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
	if err != nil {
		logger.Fatal("Failed to connect MongoDB", zap.Error(err))
	}

	// init tx tracker api client.
	txTrackerAPIClient, err := txtrackerProcessVaa.NewTxTrackerAPIClient(cfg.TxTrackerTimeout, cfg.TxTrackerURL, logger)
	if err != nil {
		logger.Fatal("Failed to create txtracker api client", zap.Error(err))
	}
	sleepTime := time.Duration(cfg.SleepTimeSeconds) * time.Second
	fromDate, _ := time.Parse(time.RFC3339, cfg.FromDate)
	toDate, _ := time.Parse(time.RFC3339, cfg.ToDate)

	return migration.NewMigrationSourceChainTx(db.Database, cfg.PageSize, sdk.ChainID(cfg.ChainID), fromDate, toDate, txTrackerAPIClient, sleepTime, logger)
}

func initProtocolStatsHourlyJob(ctx context.Context, logger *zap.Logger) *protocols.StatsJob {
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
	to := time.Now().UTC().Truncate(1 * time.Hour)
	from := to.Add(-1 * time.Hour)
	return protocols.NewStatsJob(dbWriter,
		from,
		to,
		dbconsts.ProtocolsActivityMeasurementHourly,
		dbconsts.ProtocolsStatsMeasurementHourly,
		protocolRepos,
		logger)
}

func initProtocolStatsDailyJob(ctx context.Context, logger *zap.Logger) *protocols.StatsJob {
	cfgJob, errCfg := configuration.LoadFromEnv[config.ProtocolsStatsConfiguration](ctx)
	if errCfg != nil {
		log.Fatal("error creating config", errCfg)
	}
	errUnmarshal := json.Unmarshal([]byte(cfgJob.ProtocolsJson), &cfgJob.Protocols)
	if errUnmarshal != nil {
		log.Fatal("error unmarshalling protocols config", errUnmarshal)
	}
	dbClient := influxdb2.NewClient(cfgJob.InfluxUrl, cfgJob.InfluxToken)
	dbWriter := dbClient.WriteAPIBlocking(cfgJob.InfluxOrganization, cfgJob.InfluxBucketInfinite)

	protocolRepos := make([]repository.ProtocolRepository, 0, len(cfgJob.Protocols))
	for _, c := range cfgJob.Protocols {
		builder, ok := repository.ProtocolsRepositoryFactory[c.Name]
		if !ok {
			log.Fatal("error creating protocol stats client. Unknown protocol:", c.Name, errCfg)
		}
		cs := builder(c.Url, logger.With(zap.String("protocol", c.Name), zap.String("url", c.Url)))
		protocolRepos = append(protocolRepos, cs)
	}
	to := time.Now().UTC().Truncate(24 * time.Hour)
	from := to.Add(-24 * time.Hour)
	return protocols.NewStatsJob(dbWriter,
		from,
		to,
		dbconsts.ProtocolsActivityMeasurementDaily,
		dbconsts.ProtocolsStatsMeasurementDaily,
		protocolRepos,
		logger)
}

func handleExit() {
	if r := recover(); r != nil {
		if e, ok := r.(exitCode); ok {
			os.Exit(int(e))
		}
		panic(r) // not an Exit, bubble up
	}
}
