package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/go-redis/redis"
	common "github.com/wormhole-foundation/wormhole-explorer/common/coingecko"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	filePrices "github.com/wormhole-foundation/wormhole-explorer/common/prices"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/config"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/internal/coingecko"
	apiPrices "github.com/wormhole-foundation/wormhole-explorer/jobs/internal/prices"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/notional"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/report"
	"go.uber.org/zap"
)

type exitCode int

func main() {
	defer handleExit()
	context := context.Background()

	// get the config
	cfg, errConf := config.New(context)
	if errConf != nil {
		log.Fatal("error creating config", errConf)
	}

	logger := logger.New("wormhole-explorer-jobs", logger.WithLevel(cfg.LogLevel))
	logger.Info("started job execution", zap.String("job_id", cfg.JobID))

	var err error
	switch cfg.JobID {
	case jobs.JobIDNotional:
		nCfg, errCfg := config.NewNotionalConfiguration(context)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}
		notionalJob := initNotionalJob(context, nCfg, logger)
		err = notionalJob.Run()
	case jobs.JobIDTransferReport:
		aCfg, errCfg := config.NewTransferReportConfiguration(context)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}
		transferReport := initTransferReportJob(context, aCfg, logger)
		err = transferReport.Run(context)
	case jobs.JobIDHistoricalPrices:
		hCfg, errCfg := config.NewHistoricalPricesConfiguration(context)
		if errCfg != nil {
			log.Fatal("error creating config", errCfg)
		}
		historyPrices := initHistoricalPricesJob(context, hCfg, logger)
		err = historyPrices.Run(context)

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

func handleExit() {
	if r := recover(); r != nil {
		if e, ok := r.(exitCode); ok {
			os.Exit(int(e))
		}
		panic(r) // not an Exit, bubble up
	}
}
