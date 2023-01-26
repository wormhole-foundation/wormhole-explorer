package wormscan

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/observations"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
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

// RegisterRoutes sets up the handlers for the Wormscan API.
func RegisterRoutes(
	app *fiber.App,
	rootLogger *zap.Logger,
	vaaService *vaa.Service,
	obsService *observations.Service,
	governorService *governor.Service,
	infrastructureService *infrastructure.Service,
) {

	// Set up controllers
	vaaCtrl := vaa.NewController(vaaService, rootLogger)
	observationsCtrl := observations.NewController(obsService, rootLogger)
	governorCtrl := governor.NewController(governorService, rootLogger)
	infrastructureCtrl := infrastructure.NewController(infrastructureService)

	// Set up route handlers
	api := app.Group("/api/v1")
	api.Use(cors.New()) // TODO CORS restrictions?
	api.Use(middleware.ExtractPagination)

	// monitoring
	api.Get("/health", infrastructureCtrl.HealthCheck)
	api.Get("/ready", infrastructureCtrl.ReadyCheck)

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
	enqueueVaas.Get("/", governorCtrl.GetEnqueueVaas)
	enqueueVaas.Get("/:chain", governorCtrl.GetEnqueuedVaasByChainID)
}
