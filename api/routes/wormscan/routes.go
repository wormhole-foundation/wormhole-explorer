package wormscan

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	addrsvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/address"
	govsvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	obssvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/observations"
	opsvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/operations"
	protocolssvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/protocols"
	relayssvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/relays"
	statssvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/stats"
	supplySvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/supply"
	trxsvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	vaasvc "github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/address"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/governor"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/observations"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/operations"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/protocols"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/relays"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/stats"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/supply"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"

	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan/vaa"

	"go.uber.org/zap"
)

// RegisterRoutes sets up the handlers for the Wormscan API.
func RegisterRoutes(
	notSupportedByEnv fiber.Handler,
	app *fiber.App,
	rootLogger *zap.Logger,
	addressService *addrsvc.Service,
	vaaService *vaasvc.Service,
	obsService *obssvc.Service,
	governorService *govsvc.Service,
	transactionsService *trxsvc.Service,
	relaysService *relayssvc.Service,
	operationsService *opsvc.Service,
	statsService *statssvc.Service,
	protocolsService *protocolssvc.Service,
	supplyService *supplySvc.Service,
	checks ...health.Check,
) {

	// Set up controllers
	addressCtrl := address.NewController(addressService, rootLogger)
	vaaCtrl := vaa.NewController(vaaService, rootLogger)
	observationsCtrl := observations.NewController(obsService, rootLogger)
	governorCtrl := governor.NewController(governorService, rootLogger)
	infrastructureCtrl := infrastructure.NewController(checks, rootLogger)
	transactionCtrl := transactions.NewController(transactionsService, rootLogger)
	relaysCtrl := relays.NewController(relaysService, rootLogger)
	opsCtrl := operations.NewController(operationsService, rootLogger)
	statsCtrl := stats.NewController(statsService, rootLogger)
	contributorsCtrl := protocols.NewController(rootLogger, protocolsService)
	supplyCtrl := supply.NewController(supplyService, rootLogger)

	// Set up route handlers
	api := app.Group("/api/v1")
	api.Use(cors.New()) // TODO CORS restrictions?
	api.Use(compress.New(compress.Config{
		Next: func(c *fiber.Ctx) bool {
			endpointsToCompress := []string{"/tokens-symbol-activity", "/application-activity"}
			path := c.Path()
			for _, endpoint := range endpointsToCompress {
				if strings.HasSuffix(path, endpoint) {
					return false
				}
			}
			return true // Don't execute middleware if Next returns true
		},
		Level: compress.LevelBestSpeed,
	}))

	// monitoring
	api.Get("/health", infrastructureCtrl.HealthCheck)
	api.Get("/ready", infrastructureCtrl.ReadyCheck)
	api.Get("/version", infrastructureCtrl.Version)

	// Circulating Supply
	api.Get("/supply/circulating", supplyCtrl.GetCirculatingSupply)
	api.Get("/supply/total", supplyCtrl.GetTotalSupply)
	api.Get("/supply", supplyCtrl.GetSupplyInfo)

	// accounts resource
	api.Get("/address/:id", addressCtrl.FindById)

	// analytics, transactions, custom endpoints
	api.Get("/global-tx/:chain/:emitter/:sequence", transactionCtrl.FindGlobalTransactionByID)
	api.Get("/last-txs", transactionCtrl.GetLastTransactions)
	api.Get("/scorecards", transactionCtrl.GetScorecards)
	api.Get("/x-chain-activity", transactionCtrl.GetChainActivity)
	api.Get("/x-chain-activity/tops", transactionCtrl.GetChainActivityTops)
	api.Get("/top-assets-by-volume", transactionCtrl.GetTopAssets)
	api.Get("/top-chain-pairs-by-num-transfers", transactionCtrl.GetTopChainPairs)
	api.Get("token/:chain/:token_address", transactionCtrl.GetTokenByChainAndAddress)
	api.Get("/transactions", transactionCtrl.ListTransactions)
	api.Get("/transactions/:chain/:emitter/:sequence", transactionCtrl.GetTransactionByID)
	api.Get("/application-activity", transactionCtrl.GetApplicationActivity)
	api.Get("/tokens-symbol-volume", transactionCtrl.GetTokensVolume)
	api.Get("/tokens-symbol-activity", transactionCtrl.GetTokenSymbolActivity)

	// stats custom endpoints
	api.Get("/top-symbols-by-volume", statsCtrl.GetTopSymbolsByVolume)
	api.Get("/top-100-corridors", statsCtrl.GetTopCorridors)
	api.Get("/protocols/stats", contributorsCtrl.GetProtocolsTotalValues)
	api.Get("/native-token-transfer/summary", notSupportedByEnv, statsCtrl.GetNativeTokenTransferSummary)
	api.Get("/native-token-transfer/activity", notSupportedByEnv, statsCtrl.GetNativeTokenTransferActivity)
	api.Get("/native-token-transfer/transfer-by-time", notSupportedByEnv, statsCtrl.GetNativeTokenTransferByTime)
	api.Get("/native-token-transfer/top-address", notSupportedByEnv, statsCtrl.GetNativeTokenTransferAddressTop)
	api.Get("/native-token-transfer/top-holder", notSupportedByEnv, statsCtrl.GetNativeTokenTransferTopHolder)

	// operations resource
	operations := api.Group("/operations")
	operations.Get("/", opsCtrl.FindAll)
	operations.Get("/:chain/:emitter/:sequence", opsCtrl.FindById)

	// vaas resource
	vaas := api.Group("/vaas")
	vaas.Get("/vaa-counts", vaaCtrl.GetVaaCount)
	vaas.Get("/", vaaCtrl.FindAll)
	vaas.Get("/:chain", vaaCtrl.FindByChain)
	vaas.Get("/:chain/:emitter", vaaCtrl.FindByEmitter)
	vaas.Get("/:chain/:emitter/:sequence", vaaCtrl.FindById)
	vaas.Get("/:chain/:emitter/:sequence/duplicated", vaaCtrl.FindDuplicatedById)
	vaas.Post("/parse", vaaCtrl.ParseVaa)

	// observations resource
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
	enqueueVaas.Get("/", governorCtrl.GetEnqueuedVaas)
	enqueueVaas.Get("/:chain", governorCtrl.GetEnqueuedVaasByChainID)
	governor.Get("/vaas", governorCtrl.GetGovernorVaas)

	relays := api.Group("/relays")
	relays.Get("/:chain/:emitter/:sequence", relaysCtrl.FindOne)
}
