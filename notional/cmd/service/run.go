package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	wormscanNotionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	health "github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/notional/config"
	"github.com/wormhole-foundation/wormhole-explorer/notional/http/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/notional/prices"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type exitCode int

func handleExit() {
	if r := recover(); r != nil {
		if e, ok := r.(exitCode); ok {
			os.Exit(int(e))
		}
		panic(r) // not an Exit, bubble up
	}
}

func Run() {
	defer handleExit()
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	// load configuration
	config, err := config.New(rootCtx)
	if err != nil {
		log.Fatal("Error creating config", err)
	}

	// build logger
	logger := logger.New("wormhole-notional", logger.WithLevel(config.LogLevel))
	logger.Info("starting notional service...")

	// setup DB connection
	logger.Info("connecting to MongoDB...")
	db, err := dbutil.Connect(rootCtx, logger, config.MongodbURI, config.MongodbDatabase, false)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	// get health check functions.
	logger.Info("creating health check functions...")
	healthChecks, err := newHealthChecks(rootCtx, config, db.Database)
	if err != nil {
		logger.Fatal("failed to create health checks", zap.Error(err))
	}

	//create notional cache
	logger.Info("initializing notional cache...")
	notionalCache, err := newNotionalCache(rootCtx, config, logger)
	if err != nil {
		logger.Fatal("failed to create notional cache", zap.Error(err))
	}

	// create token provider
	tokenProvider := domain.NewTokenProvider(config.P2pNetwork)

	//create repositories
	repository := prices.NewPriceRepository(db.Database, logger)

	//create services
	priceService := prices.NewPriceService(repository, tokenProvider, notionalCache, logger)

	//create controllers
	priceController := prices.NewController(priceService, logger)

	// create and start server.
	logger.Info("initializing infrastructure server...")

	server := infrastructure.NewServer(logger, config.Port, config.PprofEnabled, priceController, healthChecks...)
	server.Start()

	// Waiting for signal
	logger.Info("waiting for termination signal or context cancellation...")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-rootCtx.Done():
		logger.Warn("terminating (root context cancelled)")
	case signal := <-sigterm:
		logger.Info("terminating (signal received)", zap.String("signal", signal.String()))
	}

	logger.Info("cancelling root context...")
	rootCtxCancel()

	logger.Info("closing HTTP server...")
	server.Stop()

	logger.Info("closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	logger.Info("terminated successfully")
}

func newHealthChecks(
	ctx context.Context,
	config *config.Configuration,
	db *mongo.Database,
) ([]health.Check, error) {

	healthChecks := []health.Check{
		health.Mongo(db),
	}
	return healthChecks, nil
}

func newNotionalCache(
	ctx context.Context,
	cfg *config.Configuration,
	logger *zap.Logger,
) (wormscanNotionalCache.NotionalLocalCacheReadable, error) {

	// use a distributed cache and for notional a pubsub to sync local cache.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.CacheURL})

	// get notional cache client and init load to local cache
	notionalCache, err := wormscanNotionalCache.NewNotionalCache(ctx, redisClient, cfg.CachePrefix, cfg.CacheChannel, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create notional cache client: %w", err)
	}
	notionalCache.Init(ctx)

	return notionalCache, nil
}
