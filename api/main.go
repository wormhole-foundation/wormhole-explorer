package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/heartbeats"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/infraestructure"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/observations"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	wormscanCache "github.com/wormhole-foundation/wormhole-explorer/api/internal/cache"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"go.uber.org/zap"
)

var cacheConfig = cache.Config{
	Next: func(c *fiber.Ctx) bool {
		return c.Query("refresh") == "true"
	},
	Expiration:           1 * time.Second,
	CacheControl:         true,
	StoreResponseHeaders: true,
}

func healthOk(ctx *fiber.Ctx) error {
	ctx.Status(200)
	return ctx.SendString("Ok")
}

// @title Wormhole Guardian API
// @version 1.0
// @description Wormhole Guardian API
// @description To get information from the Wormhole Network.
// @description Check each endpoint documentation for more information.
// @termsOfService https://wormhole.com/
// @contact.name API Support
// @contact.url http://wormhole.com/support
// @contact.email info@wormhole.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /v1
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
	lvl, err := cfg.GetLogLevel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid logging level set: %v", cfg.LogLevel)
		panic(err)
	}

	rootLogger := ipfslog.Logger("wormhole-api").Desugar()
	ipfslog.SetAllLoggers(lvl)

	// Setup DB
	cli, err := db.Connect(appCtx, cfg.DB.URL)
	if err != nil {
		panic(err)
	}
	db := cli.Database(cfg.DB.Name)

	// Get cache get function
	cacheGetFunc := NewCache(cfg, rootLogger)

	// Setup repositories
	vaaRepo := vaa.NewRepository(db, rootLogger)
	obsRepo := observations.NewRepository(db, rootLogger)
	governorRepo := governor.NewRepository(db, rootLogger)
	infraestructureRepo := infraestructure.NewRepository(db, rootLogger)
	heartbeatsRepo := heartbeats.NewRepository(db, rootLogger)

	// Setup services
	vaaService := vaa.NewService(vaaRepo, cacheGetFunc, rootLogger)
	obsService := observations.NewService(obsRepo, rootLogger)
	governorService := governor.NewService(governorRepo, rootLogger)
	infraestructureService := infraestructure.NewService(infraestructureRepo, rootLogger)
	heartbeatsService := heartbeats.NewService(heartbeatsRepo, rootLogger)

	// Setup controllers
	vaaCtrl := vaa.NewController(vaaService, rootLogger)
	observationsCtrl := observations.NewController(obsService, rootLogger)
	governorCtrl := governor.NewController(governorService, rootLogger)
	infraestructureCtrl := infraestructure.NewController(infraestructureService)
	guardianCtrl := guardian.NewController(rootLogger)
	heartbeatsCtrl := heartbeats.NewController(heartbeatsService, rootLogger)

	// Setup app with custom error handling.
	response.SetEnableStackTrace(*cfg)
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})

	// Middleware
	prometheus := fiberprometheus.New("wormscan")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "level=info timestamp=${time} method=${method} path=${path} status${status} request_id=${locals:requestid}\n",
	}))

	api := app.Group("/api/v1")
	api.Use(cors.New()) // TODO CORS restrictions?
	api.Use(middleware.ExtractPagination)

	api.Get("/health", infraestructureCtrl.HealthCheck)
	api.Get("/ready", infraestructureCtrl.ReadyCheck)

	// vaas resource
	vaas := api.Group("/vaas")
	vaas.Use(cache.New(cacheConfig))
	vaas.Get("/vaa-counts", vaaCtrl.GetVaaCount)
	vaas.Get("/", vaaCtrl.FindAll)
	vaas.Get("/:chain", vaaCtrl.FindByChain)
	vaas.Get("/:chain/:emitter", vaaCtrl.FindByEmitter)
	vaas.Get("/:chain/:emitter/:sequence", vaaCtrl.FindById)

	// oservations resource
	observations := api.Group("/observations")
	observations.Get("/", observationsCtrl.FindAll)
	observations.Get("/:chain", observationsCtrl.FindAllByChain)
	observations.Get("/:chain/:emitter", observationsCtrl.FindAllByEmitter)
	observations.Get("/:chain/:emitter/:sequence", observationsCtrl.FindAllByVAA)
	observations.Get("/:chain/:emitter/:sequence/:signer/:hash", observationsCtrl.FindOne)

	// governor resources
	governor := api.Group("/governor")
	governorLimit := governor.Group("/limit")
	governorLimit.Get("/", governorCtrl.GetGovernorLimit)

	governorConfigs := governor.Group("/config")
	governorConfigs.Get("/", governorCtrl.FindGovernorConfigurations)
	governorConfigs.Get("/:guardian_address", governorCtrl.FindGovernorConfigurationByGuardianAddress)

	governorStatus := governor.Group("/status")
	governorStatus.Get("/", governorCtrl.FindGovernorStatus)
	governorStatus.Get("/:guardian_address", governorCtrl.FindGovernorStatusByGuardianAddress)

	governorNotional := governor.Group("/notional")
	governorNotional.Get("/limit/", governorCtrl.FindNotionalLimit)
	governorNotional.Get("/limit/:chain", governorCtrl.GetNotionalLimitByChainID)
	governorNotional.Get("/available/", governorCtrl.GetAvailableNotional)
	governorNotional.Get("/available/:chain", governorCtrl.GetAvailableNotionalByChainID)
	governorNotional.Get("/max_available/:chain", governorCtrl.GetMaxNotionalAvailableByChainID)

	enqueueVaas := governor.Group("/enqueued_vaas")
	enqueueVaas.Get("/", governorCtrl.GetEnqueueVass)
	enqueueVaas.Get("/:chain", governorCtrl.GetEnqueueVassByChainID)

	// v1 guardian public api.
	publicAPIV1 := app.Group("/v1")
	// signedVAA resource.
	signedVAA := publicAPIV1.Group("/signed_vaa")
	signedVAA.Get("/:chain/:emitter/:sequence", vaaCtrl.FindSignedVAAByID)
	signedBatchVAA := publicAPIV1.Group("/signed_batch_vaa")
	signedBatchVAA.Get("/:chain/:emitter/:sequence", vaaCtrl.FindSignedBatchVAAByID)
	// guardianSet resource.
	guardianSet := publicAPIV1.Group("/guardianset")
	guardianSet.Get("/current", guardianCtrl.GetGuardianSet)
	// heartbeats resource.
	heartbeats := publicAPIV1.Group("/heartbeats")
	heartbeats.Get("", heartbeatsCtrl.GetLastHeartbeats)
	// governor resource.
	gov := publicAPIV1.Group("/governor")
	gov.Get("/available_notional_by_chain", governorCtrl.GetAvailNotionByChain)
	gov.Get("/enqueued_vaas", governorCtrl.GetEnqueuedVaas)
	gov.Get("/is_vaa_enqueued/:chain/:emitter/:sequence", governorCtrl.IsVaaEnqueued)
	gov.Get("/token_list", governorCtrl.GetTokenList)

	app.Listen(":" + strconv.Itoa(cfg.PORT))
}

// NewCache return a CacheGetFunc to get a value by a Key from cache.
func NewCache(cfg *config.AppConfig, looger *zap.Logger) wormscanCache.CacheGetFunc {
	if cfg.RunMode == config.RunModeDevelopmernt && !cfg.Cache.Enabled {
		dummyCacheClient := wormscanCache.NewDummyCacheClient()
		return dummyCacheClient.Get
	}
	cacheClient := wormscanCache.NewCacheClient(cfg.Cache.URL, cfg.Cache.Enabled, looger)
	return cacheClient.Get
}
