package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/repository"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/stats"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/configuration"

	"github.com/go-redis/redis/v8"
	wormscanNotionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
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
	case jobs.JobIDMigrationNativeTxHash:
		job := initMigrateNativeTxHashJob(ctx, logger)
		err = job.Run(ctx)
	case jobs.JobIDNTTTopAddressStats:
		job := initNTTTopAddressStatsJob(ctx, logger)
		err = job.Run(ctx)
	case jobs.JobIDNTTTopHolderStats:
		job := initNTTTopHolderStatsJob(ctx, logger)
		err = job.Run(ctx)
	case jobs.JobIDNTTMedianStats:
		job := initNTTMedianStatsJob(ctx, logger)
		err = job.Run(ctx)
	default:
		logger.Error("Invalid job id", zap.String("job_id", cfg.JobID))
	}

	if err != nil {
		logger.Fatal("failed job execution", zap.String("job_id", cfg.JobID), zap.Error(err))
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

func initMigrateSourceTxJob(ctx context.Context, cfg *config.MigrateSourceTxConfiguration, _ sdk.ChainID, logger *zap.Logger) *migration.MigrateSourceChainTx {
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

func initMigrateNativeTxHashJob(ctx context.Context, logger *zap.Logger) *migration.MigrateNativeTxHash {
	cfgJob, errCfg := configuration.LoadFromEnv[config.MigrateNativeTxHashConfiguration](ctx)
	if errCfg != nil {
		log.Fatal("error creating config", errCfg)
	}
	db, err := dbutil.Connect(ctx, logger, cfgJob.MongoURI, cfgJob.MongoDatabase, false)
	if err != nil {
		logger.Fatal("Failed to connect MongoDB", zap.Error(err))
	}
	return migration.NewMigrationNativeTxHash(db.Database, cfgJob.PageSize, logger)
}

func initNTTTopAddressStatsJob(ctx context.Context, logger *zap.Logger) *stats.NTTTopAddressJob {
	cfgJob, errCfg := configuration.LoadFromEnv[config.NTTTopAddressStatsConfiguration](ctx)
	if errCfg != nil {
		log.Fatal("error creating config", errCfg)
	}

	// init influx client.
	influxClient := influxdb2.NewClient(cfgJob.InfluxUrl, cfgJob.InfluxToken)

	// init redis client.
	redisClient := redis.NewClient(&redis.Options{Addr: cfgJob.CacheUrl})

	// init cache client.
	cache, err := cache.NewCacheClient(redisClient, true, cfgJob.CachePrefix, logger)
	if err != nil {
		log.Fatal("error creating cache client", err)
	}

	return stats.NewNTTTopAddressJob(influxClient, cfgJob.InfluxOrganization, cfgJob.InfluxBucketInfinite, cache, logger)
}

func initNTTTopHolderStatsJob(ctx context.Context, logger *zap.Logger) *stats.NTTTopHolderJob {
	cfgJob, errCfg := configuration.LoadFromEnv[config.NTTTopHolderStatsConfiguration](ctx)
	if errCfg != nil {
		log.Fatal("error creating config", errCfg)
	}

	// init redis client.
	redisClient := redis.NewClient(&redis.Options{Addr: cfgJob.CacheUrl})

	// init cache client.
	cache, err := cache.NewCacheClient(redisClient, true, cfgJob.CachePrefix, logger)
	if err != nil {
		log.Fatal("error creating cache client", err)
	}

	tokenProvider := domain.NewTokenProvider(cfgJob.P2pNetwork)

	// get notional cache client and init load to local cache
	notionalCache, err := wormscanNotionalCache.NewNotionalCache(ctx, redisClient, cfgJob.CachePrefix, cfgJob.CacheNotionalChannel, logger)
	if err != nil {
		log.Fatal("failed to create notional cache client", err)
	}
	notionalCache.Init(ctx)

	return stats.NewNTTTopHolderJob(resty.New(), cfgJob.ArkhamUrl, cfgJob.ArkhamApiKey, cfgJob.SolanaUrl, cache, tokenProvider, notionalCache, logger)
}

func initNTTMedianStatsJob(ctx context.Context, logger *zap.Logger) *stats.NTTMedian {
	cfgJob, errCfg := configuration.LoadFromEnv[config.NTTMedianStatsConfiguration](ctx)
	if errCfg != nil {
		log.Fatal("error creating config", errCfg)
	}

	// init influx client.
	influxClient := influxdb2.NewClient(cfgJob.InfluxUrl, cfgJob.InfluxToken)

	// init redis client.
	redisClient := redis.NewClient(&redis.Options{Addr: cfgJob.CacheUrl})

	// init cache client.
	cache, err := cache.NewCacheClient(redisClient, true, cfgJob.CachePrefix, logger)
	if err != nil {
		log.Fatal("error creating cache client", err)
	}

	return stats.NewNTTMedian(influxClient, cfgJob.InfluxOrganization, cfgJob.InfluxBucketInfinite, cache, logger)
}

func handleExit() {
	if r := recover(); r != nil {
		if e, ok := r.(exitCode); ok {
			os.Exit(int(e))
		}
		panic(r) // not an Exit, bubble up
	}
}
