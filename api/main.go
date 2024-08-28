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

	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/protocols"

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
	guardianHandlers "github.com/wormhole-foundation/wormhole-explorer/api/handlers/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/heartbeats"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/observations"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/operations"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/relays"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/stats"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/tvl"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan"
	rpcApi "github.com/wormhole-foundation/wormhole-explorer/api/rpc"
	wormscanCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	vaaPayloadParser "github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/coingecko"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	xlogger "github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	stats2 "github.com/wormhole-foundation/wormhole-explorer/common/stats"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

//go:embed docs/swagger.json
var swagger []byte

// GetSwagger godoc
// @Description Returns the swagger specification for this API.
// @Tags wormholescan
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

// @title Wormholescan API
// @version 1.0
// @description Wormhole Guardian API
// @description This is the API for the Wormhole Guardian and Explorer.
// @description The API has two namespaces: wormholescan and guardian.
// @description wormholescan is the namespace for the explorer and the new endpoints. The prefix is /api/v1.
// @description guardian is the legacy namespace backguard compatible with guardian node API. The prefix is /v1.
// @description This API is public and does not require authentication although some endpoints are rate limited.
// @description Check each endpoint documentation for more information.
// @termsOfService https://wormhole.com/
// @contact.name API Support
// @contact.url https://discord.com/invite/wormholecrypto
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
	db, err := dbutil.Connect(appCtx, rootLogger, cfg.DB.URL, cfg.DB.Name, false)
	if err != nil {
		rootLogger.Fatal("failed to connect to MongoDB", zap.Error(err))
	}

	// Get cache get function
	rootLogger.Info("initializing cache")
	cache, err := NewCache(appCtx, cfg, rootLogger)
	if err != nil {
		rootLogger.Fatal("failed to initialize cache", zap.Error(err))
	}

	// cfg.Cache.Expiration
	rootLogger.Info("initializing TVL cache")
	tvl := tvl.NewTVL(cfg.P2pNetwork, cache, cfg.Cache.TvlKey, cfg.Cache.TvlExpiration, rootLogger)

	// coingeckoAPI client
	coingeckoAPI := coingecko.NewCoinGeckoAPI(cfg.Coingecko.URL,
		cfg.Coingecko.HeaderKey, cfg.Coingecko.ApiKey)

	//InfluxDB client
	rootLogger.Info("initializing InfluxDB client")
	influxCli := newInfluxClient(cfg.Influx.URL, cfg.Influx.Token)

	//VaaPayloadParser client
	vaaParserFunc, err := NewVaaParserFunc(cfg, rootLogger)
	if err != nil {
		rootLogger.Fatal("failed to initialize VAA parser", zap.Error(err))
	}

	// create token provider
	tokenProvider := domain.NewTokenProvider(cfg.P2pNetwork)

	// Set up repositories
	rootLogger.Info("initializing repositories")
	addressRepo := address.NewRepository(db.Database, rootLogger)
	vaaRepo := vaa.NewRepository(db.Database, rootLogger)
	obsRepo := observations.NewRepository(db.Database, rootLogger)
	governorRepo := governor.NewRepository(db.Database, rootLogger)
	infrastructureRepo := infrastructure.NewRepository(db.Database, rootLogger)
	heartbeatsRepo := heartbeats.NewRepository(db.Database, rootLogger)
	transactionsRepo := transactions.NewRepository(
		tvl,
		cfg.P2pNetwork,
		influxCli,
		cfg.Influx.Organization,
		cfg.Influx.Bucket24Hours,
		cfg.Influx.Bucket30Days,
		cfg.Influx.BucketInfinite,
		db.Database,
		rootLogger,
	)
	relaysRepo := relays.NewRepository(db.Database, rootLogger)
	operationsRepo := operations.NewRepository(db.Database, rootLogger)
	statsRepo := stats.NewRepository(
		influxCli,
		cfg.Influx.Organization,
		cfg.Influx.Bucket24Hours,
		cfg.Influx.BucketInfinite,
		coingeckoAPI,
		tokenProvider,
		rootLogger)
	statsAddressRepo := stats2.NewAddressRepository(
		influxCli,
		cfg.Influx.Organization,
		cfg.Influx.BucketInfinite,
		cache,
		rootLogger)
	statsHolderRepo := stats2.NewHolderRepositoryReadable(cache, rootLogger)

	protocolsRepo := protocols.NewRepository(
		protocols.WrapQueryAPI(influxCli.QueryAPI(cfg.Influx.Organization)),
		cfg.Influx.BucketInfinite,
		cfg.Influx.Bucket30Days,
		rootLogger)
	guardianSetRepository := repository.NewGuardianSetRepository(db.Database, rootLogger)

	metrics := metrics.NewPrometheusMetrics(cfg.Environment)

	// Set up services
	rootLogger.Info("initializing services")
	expirationTime := time.Duration(cfg.Cache.MetricExpiration) * time.Minute
	addressService := address.NewService(addressRepo, rootLogger)
	vaaService := vaa.NewService(vaaRepo, cache.Get, vaaParserFunc, rootLogger)
	obsService := observations.NewService(obsRepo, rootLogger)
	governorService := governor.NewService(governorRepo, cache, metrics, rootLogger)
	infrastructureService := infrastructure.NewService(infrastructureRepo, rootLogger)
	heartbeatsService := heartbeats.NewService(heartbeatsRepo, rootLogger)
	transactionsService := transactions.NewService(transactionsRepo, cache, expirationTime, tokenProvider, metrics, rootLogger)
	relaysService := relays.NewService(relaysRepo, rootLogger)
	operationsService := operations.NewService(operationsRepo, rootLogger)
	statsService := stats.NewService(statsRepo, statsAddressRepo, statsHolderRepo, cache, expirationTime, metrics, rootLogger)
	protocolsService := protocols.NewService(cfg.Protocols, []string{protocols.CCTP, protocols.PortalTokenBridge, protocols.NTT}, protocolsRepo, rootLogger, cache, cfg.Cache.ProtocolsStatsKey, cfg.Cache.ProtocolsStatsExpiration, metrics, tvl)
	guardianService := guardianHandlers.NewService(guardianSetRepository, cfg.P2pNetwork, cache, metrics, rootLogger)

	// Set up a custom error handler
	response.SetEnableStackTrace(*cfg)
	app := fiber.New(fiber.Config{
		ErrorHandler:          middleware.ErrorHandler,
		DisableStartupMessage: true,
		Immutable:             true,
	})

	// Configure middleware
	labels := map[string]string{"service": "wormscan-api", "environment": cfg.Environment}
	prometheus := fiberprometheus.NewWithLabels(labels, "http", "")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)
	app.Use(middleware.OriginMetrics(metrics))

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "level=info timestamp=${time} method=${method} path=${path} latency=${latency} status${status} request_id=${locals:requestid} ip=${ips} queryParams=${queryParams}\n",
		Next: func(c *fiber.Ctx) bool {
			return middleware.IsK8sPath(c.Path())
		},
	}))
	if cfg.PprofEnabled {
		app.Use(pprof.New())
	}
	app.Use(cors.New())

	// Configure rate limiter
	if cfg.RateLimit.Enabled {
		rl, err := NewRateLimiter(appCtx, cfg, rootLogger)
		if err != nil {
			panic(err)
		}
		app.Use(rl)
	}

	// Set up route handlers
	app.Get("/swagger.json", GetSwagger)
	wormscan.RegisterRoutes(app, rootLogger, addressService, vaaService, obsService, governorService, infrastructureService, transactionsService, relaysService, operationsService, statsService, protocolsService)
	guardian.RegisterRoutes(cfg, app, rootLogger, vaaService, governorService, heartbeatsService, guardianService)

	// Set up gRPC handlers
	handler := rpcApi.NewHandler(vaaService, heartbeatsService, governorService, guardianService, rootLogger)
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

	rootLogger.Info("closing cache...")
	cache.Close()

	rootLogger.Info("closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	rootLogger.Info("terminated API service successfully")
}

// NewCache get a CacheGetFunc to get a value by a Key from cache and a CacheReadable to get a value by a Key from notional local cache.
func NewCache(ctx context.Context, cfg *config.AppConfig, logger *zap.Logger) (wormscanCache.Cache, error) {

	// if run mode is development with cache is disabled, return a dummy cache client and a dummy notional cache client.
	if cfg.RunMode == config.RunModeDevelopmernt && !cfg.Cache.Enabled {
		dummyCacheClient := wormscanCache.NewDummyCacheClient()
		return dummyCacheClient, nil
	}

	// if we are not in development mode, use a distributed cache and for notional a pubsub to sync local cache.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Cache.URL})

	// get cache client
	cacheClient, err := wormscanCache.NewCacheClient(redisClient, cfg.Cache.Enabled, cfg.Cache.Prefix, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache client: %w", err)
	}

	return cacheClient, nil
}

func newInfluxClient(url, token string) influxdb2.Client {
	return influxdb2.NewClient(url, token)
}

func NewRateLimiter(ctx context.Context, cfg *config.AppConfig, logger *zap.Logger) (func(*fiber.Ctx) error, error) {

	if cfg.RateLimit.Prefix != "" {
		cfg.RateLimit.Prefix += ":rate-limiter:"
	} else {
		cfg.RateLimit.Prefix = "rate-limiter:"
	}

	enableApiTokens := len(cfg.GetApiTokens()) > 0
	enableByApiToken := make(map[string]bool)
	if enableApiTokens {
		for _, token := range cfg.GetApiTokens() {
			enableByApiToken[token] = true
		}
	}

	// initialize rate limiter
	store, err := frs.New(
		frs.Config{URL: cfg.Cache.URL, Prefix: cfg.RateLimit.Prefix})
	if err != nil {
		logger.Error("failed to initialize rate limiter",
			zap.String("url", cfg.Cache.URL),
			zap.String("prefix", cfg.RateLimit.Prefix),
			zap.Error(err))
		return nil, err
	}

	// default to 60 requests per minute
	if cfg.RateLimit.Max == 0 {
		cfg.RateLimit.Max = 60
	}

	logger.Info("rate limit enabled", zap.Int("max requests per minute", cfg.RateLimit.Max))

	router := limiter.New(limiter.Config{
		Next: func(c *fiber.Ctx) bool {
			if enableApiTokens {
				apiKey := c.Get("X-API-KEY")
				if apiKey != "" {
					_, exists := enableByApiToken[apiKey]
					return exists
				}
			}
			ip := utils.GetRealIp(c)
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
	})

	return router, nil

}

// NewVaaParserFunc returns a function to parse VAA payload.
func NewVaaParserFunc(cfg *config.AppConfig, logger *zap.Logger) (vaaPayloadParser.ParseVaaFunc, error) {
	if cfg.RunMode == config.RunModeDevelopmernt && !cfg.VaaPayloadParser.Enabled {
		return func(vaa *sdk.VAA) (any, error) {
			return nil, nil
		}, nil
	}
	vaaPayloadParserClient, err := vaaPayloadParser.NewParserVAAAPIClient(cfg.VaaPayloadParser.Timeout,
		cfg.VaaPayloadParser.URL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize VAA parser client: %w", err)
	}
	return vaaPayloadParserClient.ParseVaa, nil
}
