package main

import (
	"context"
	"encoding/json"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/configuration"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/stats"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	cfg, errConf := config.New(ctx)
	if errConf != nil {
		log.Fatal("error creating config", errConf)
	}

	logger := logger.New("wormhole-explorer-jobs", logger.WithLevel(cfg.LogLevel))
	logger.Info("started job execution", zap.String("job_id", cfg.JobID))

	var err error
	switch cfg.JobID {
	case jobs.JobIDNotional:
		nCfg, errCfg := config.NewNotionalConfiguration(ctx)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}
		notionalJob := initNotionalJob(ctx, nCfg, logger)
		err = notionalJob.Run()

	case jobs.JobIDTransferReport:
		aCfg, errCfg := config.NewTransferReportConfiguration(ctx)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}
		transferReport := initTransferReportJob(ctx, aCfg, logger)
		err = transferReport.Run(ctx)

	case jobs.JobIDHistoricalPrices:
		hCfg, errCfg := config.NewHistoricalPricesConfiguration(ctx)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}
		historyPrices := initHistoricalPricesJob(ctx, hCfg, logger)
		err = historyPrices.Run(ctx)

	case jobs.JobIDMigrationSourceTx:
		mCfg, errCfg := config.NewMigrateSourceTxConfiguration(ctx)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}

		chainID := sdk.ChainID(mCfg.ChainID)
		migrationJob := initMigrateSourceTxJob(ctx, mCfg, chainID, logger)
		err = migrationJob.Run(ctx)

	case jobs.JobIDContributorsStats:
		statsJob := initContributorsStatsJob(ctx, logger)
		err = statsJob.Run(ctx)
	case jobs.JobIDContributorsActivity:
		activityJob := initContributorsActivityJob(ctx, logger)
		err = activityJob.Run(ctx)
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
func initNotionalJob(ctx context.Context, cfg *config.NotionalConfiguration, logger *zap.Logger) *notional.NotionalJob {
	// init coingecko api client.
	api := coingecko.NewCoingeckoAPI(cfg.CoingeckoURL)
	// init redis client.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.CacheURL})
	// init token provider.
	tokenProvider := domain.NewTokenProvider(cfg.P2pNetwork)
	// create notional job.
	notionalJob := notional.NewNotionalJob(api, redisClient, cfg.CachePrefix, cfg.NotionalChannel, tokenProvider, logger)
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

func initContributorsStatsJob(ctx context.Context, logger *zap.Logger) *stats.ContributorsStatsJob {
	cfgJob, errCfg := configuration.LoadFromEnv[config.ContributorsStatsConfiguration](ctx)
	if errCfg != nil {
		log.Fatal("error creating config", errCfg)
	}
	errUnmarshal := json.Unmarshal([]byte(cfgJob.ContributorsJson), &cfgJob.Contributors)
	if errUnmarshal != nil {
		log.Fatal("error unmarshalling contributors", errUnmarshal)
	}
	dbClient := influxdb2.NewClient(cfgJob.InfluxUrl, cfgJob.InfluxToken)
	dbWriter := dbClient.WriteAPIBlocking(cfgJob.InfluxOrganization, cfgJob.InfluxBucket30Days)
	statsFetchers := make([]stats.ClientStats, 0, len(cfgJob.Contributors))
	for _, c := range cfgJob.Contributors {
		cs := stats.NewHttpRestClientStats(c.Name,
			c.Url,
			logger.With(zap.String("sevice", c.Name), zap.String("url", c.Url)),
			&http.Client{},
		)
		statsFetchers = append(statsFetchers, cs)
	}
	return stats.NewContributorsStatsJob(dbWriter, logger, cfgJob.StatsVersion, statsFetchers...)
}

func initContributorsActivityJob(ctx context.Context, logger *zap.Logger) *stats.ContributorsActivityJob {
	cfgJob, errCfg := configuration.LoadFromEnv[config.ContributorsStatsConfiguration](ctx)
	if errCfg != nil {
		log.Fatal("error creating config", errCfg)
	}
	errUnmarshal := json.Unmarshal([]byte(cfgJob.ContributorsJson), &cfgJob.Contributors)
	if errUnmarshal != nil {
		log.Fatal("error unmarshalling contributors", errUnmarshal)
	}
	dbClient := influxdb2.NewClient(cfgJob.InfluxUrl, cfgJob.InfluxToken)
	dbWriter := dbClient.WriteAPIBlocking(cfgJob.InfluxOrganization, cfgJob.InfluxBucket30Days)
	statsFetchers := make([]stats.ClientActivity, 0, len(cfgJob.Contributors))
	for _, c := range cfgJob.Contributors {
		cs := stats.NewHttpRestClientActivity(c.Name,
			c.Url,
			logger.With(zap.String("sevice", c.Name), zap.String("url", c.Url)),
			&http.Client{},
		)
		statsFetchers = append(statsFetchers, cs)
	}
	return stats.NewContributorActivityJob(dbWriter, logger, cfgJob.ActivityVersion, statsFetchers...)
}

func handleExit() {
	if r := recover(); r != nil {
		if e, ok := r.(exitCode); ok {
			os.Exit(int(e))
		}
		panic(r) // not an Exit, bubble up
	}
}
