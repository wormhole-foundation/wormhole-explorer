package main

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	frs "github.com/XLabs/fiber-redis-storage"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/improbable-eng/grpc-web/go/grpcweb"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/address"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/heartbeats"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/observations"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/tvl"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan"
	rpcApi "github.com/wormhole-foundation/wormhole-explorer/api/rpc"
	wormscanCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	wormscanNotionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	xlogger "github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"go.uber.org/zap"
)

//go:embed docs/swagger.json
var swagger []byte

// GetSwagger godoc
// @Description Returns the swagger specification for this API.
// @Tags Wormscan
// @ID swagger
// @Success 200 {object} object
// @Failure 400
// @Failure 500
// @Router /swagger.json [get]
func GetSwagger(ctx *fiber.Ctx) error {

	written, err := ctx.
		Response().
		BodyWriter().
		Write(swagger)

	if written != len(swagger) {
		return fmt.Errorf("partial write to response body: wrote %d bytes, expected %d", written, len(swagger))
	}

	return err
}

// @title Wormhole Guardian API
// @version 1.0
// @description Wormhole Guardian API
// @description This is the API for the Wormhole Guardian and Explorer.
// @description The API has two namespaces: wormscan and guardian.
// @description wormscan is the namespace for the explorer and the new endpoints. The prefix is /api/v1.
// @description guardian is the legacy namespace backguard compatible with guardian node API. The prefix is /v1.
// @description This API is public and does not require authentication although some endpoints are rate limited.
// @description Check each endpoint documentation for more information.
// @termsOfService https://wormhole.com/
// @contact.name API Support
// @contact.url http://wormhole.com/support
// @contact.email info@wormhole.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /
func main() {
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Grab config
	cfg, err := config.Get()
	if err != nil {
		fmt.Fprint(os.Stderr, "Error parsing configuration")
		panic(err)
	}

	// Logging
	rootLogger := xlogger.New("wormhole-api", xlogger.WithLevel(cfg.LogLevel))
	defer rootLogger.Sync()

	// Setup DB
	rootLogger.Info("connecting to MongoDB")
	cli, err := db.Connect(appCtx, cfg.DB.URL)
	if err != nil {
		panic(err)
	}
	db := cli.Database(cfg.DB.Name)

	// Get cache get function
	rootLogger.Info("initializing notional cache")
	cache, notionalCache := NewCache(appCtx, cfg, rootLogger)

	// cfg.Cache.Expiration
	rootLogger.Info("initializing TVL cache")
	tvl := tvl.NewTVL(cfg.P2pNetwork, cache, cfg.Cache.TvlKey, cfg.Cache.TvlExpiration, rootLogger)

	//InfluxDB client
	rootLogger.Info("initializing InfluxDB client")
	influxCli := newInfluxClient(cfg.Influx.URL, cfg.Influx.Token)

	// Set up repositories
	rootLogger.Info("initializing repositories")
	addressRepo := address.NewRepository(db, rootLogger)
	vaaRepo := vaa.NewRepository(db, rootLogger)
	obsRepo := observations.NewRepository(db, rootLogger)
	governorRepo := governor.NewRepository(db, rootLogger)
	infrastructureRepo := infrastructure.NewRepository(db, rootLogger)
	heartbeatsRepo := heartbeats.NewRepository(db, rootLogger)
	transactionsRepo := transactions.NewRepository(
		tvl,
		influxCli,
		cfg.Influx.Organization,
		cfg.Influx.Bucket24Hours,
		cfg.Influx.Bucket30Days,
		cfg.Influx.BucketInfinite,
		db,
		rootLogger,
	)

	// Set up services
	rootLogger.Info("initializing services")
	addressService := address.NewService(addressRepo, rootLogger)
	vaaService := vaa.NewService(vaaRepo, cache.Get, rootLogger)
	obsService := observations.NewService(obsRepo, rootLogger)
	governorService := governor.NewService(governorRepo, rootLogger)
	infrastructureService := infrastructure.NewService(infrastructureRepo, rootLogger)
	heartbeatsService := heartbeats.NewService(heartbeatsRepo, rootLogger)
	transactionsService := transactions.NewService(transactionsRepo, cache, rootLogger)

	// Set up a custom error handler
	response.SetEnableStackTrace(*cfg)
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})

	// Configure middleware
	prometheus := fiberprometheus.New("wormscan")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "level=info timestamp=${time} method=${method} path=${path} status${status} request_id=${locals:requestid}\n",
	}))
	if cfg.PprofEnabled {
		app.Use(pprof.New())
	}
	app.Use(cors.New())

	// Configure rate limiter
	if cfg.RateLimit.Enabled {

		store, err := frs.New(
			frs.Config{URL: fmt.Sprintf("redis://%s", cfg.Cache.URL), Prefix: cfg.RateLimit.Prefix})
		if err != nil {
			panic(err)
		}

		// default to 60 requests per minute
		if cfg.RateLimit.Max == 0 {
			cfg.RateLimit.Max = 60
		}

		rootLogger.Info("rate limit enabled", zap.Int("max requests per minute", cfg.RateLimit.Max))
		app.Use(limiter.New(limiter.Config{
			Next: func(c *fiber.Ctx) bool {

				ip := utils.GetRealIp(c)
				rootLogger.Info("rate limit", zap.String("ip", ip))
				return utils.IsPrivateIPAsString(ip)
			},
			Max:        cfg.RateLimit.Max,
			Expiration: 60 * time.Second,
			KeyGenerator: func(c *fiber.Ctx) string {
				return utils.GetRealIp(c)
			},
			LimitReached: func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusTooManyRequests)
			},
			Storage: store,
		}))
	}

	// Set up route handlers
	app.Get("/swagger.json", GetSwagger)
	wormscan.RegisterRoutes(app, rootLogger, addressService, vaaService, obsService, governorService, infrastructureService, transactionsService)
	guardian.RegisterRoutes(cfg, app, rootLogger, vaaService, governorService, heartbeatsService)

	// Set up gRPC handlers
	handler := rpcApi.NewHandler(vaaService, heartbeatsService, governorService, rootLogger, cfg.P2pNetwork)
	grpcServer := rpcApi.NewServer(handler, rootLogger)
	grpcWebServer := grpcweb.WrapServer(grpcServer)
	app.Use(
		adaptor.HTTPMiddleware(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if grpcWebServer.IsGrpcWebRequest(r) {
					grpcWebServer.ServeHTTP(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			})
		}))

	rootLogger.Info("starting HTTP server in a separate goroutine")
	go func() {
		if err := app.Listen(":" + strconv.Itoa(cfg.PORT)); err != nil {
			panic("failed to start HTTP server: " + err.Error())
		}
	}()

	// Waiting for signal
	rootLogger.Info("waiting for signal or context cancellation")
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-appCtx.Done():
		rootLogger.Warn("terminating with root context cancelled")
	case signal := <-sigterm:
		rootLogger.Info("terminating with signal", zap.String("signal", signal.String()))
	}

	rootLogger.Info("cleanup tasks...")
	rootLogger.Info("shutting down server...")
	app.Shutdown()
	rootLogger.Info("closing notional cache...")
	notionalCache.Close()
	rootLogger.Info("closing cache...")
	cache.Close()
	rootLogger.Info("terminated API service successfully")
}

// NewCache get a CacheGetFunc to get a value by a Key from cache and a CacheReadable to get a value by a Key from notional local cache.
func NewCache(ctx context.Context, cfg *config.AppConfig, logger *zap.Logger) (wormscanCache.Cache, wormscanNotionalCache.NotionalLocalCacheReadable) {
	// if run mode is development with cache is disabled, return a dummy cache client and a dummy notional cache client.
	if cfg.RunMode == config.RunModeDevelopmernt && !cfg.Cache.Enabled {
		dummyCacheClient := wormscanCache.NewDummyCacheClient()
		dummyNotionalCache := wormscanNotionalCache.NewDummyNotionalCache()
		return dummyCacheClient, dummyNotionalCache
	}

	// if we are not in development mode, use a distributed cache and for notional a pubsub to sync local cache.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Cache.URL})

	// get cache client
	cacheClient, _ := wormscanCache.NewCacheClient(redisClient, cfg.Cache.Enabled, logger)

	// get notional cache client and init load to local cache
	notionalCache, _ := wormscanNotionalCache.NewNotionalCache(ctx, redisClient, cfg.Cache.Channel, logger)
	notionalCache.Init(ctx)

	return cacheClient, notionalCache
}

func newInfluxClient(url, token string) influxdb2.Client {
	return influxdb2.NewClient(url, token)
}
