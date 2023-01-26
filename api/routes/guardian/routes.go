package guardian

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/heartbeats"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"go.uber.org/zap"
)

// RegisterRoutes sets up the handlers for the Guardian API.
func RegisterRoutes(
	app *fiber.App,
	rootLogger *zap.Logger,
	vaaService *vaa.Service,
	governorService *governor.Service,
	heartbeatsService *heartbeats.Service,
) {

	// Set up controllers
	vaaCtrl := vaa.NewController(vaaService, rootLogger)
	governorCtrl := governor.NewController(governorService, rootLogger)
	guardianCtrl := guardian.NewController(rootLogger)
	heartbeatsCtrl := heartbeats.NewController(heartbeatsService, rootLogger)

	// Set up route handlers
	apiV1 := app.Group("/v1")

	// signedVAA resource
	signedVAA := apiV1.Group("/signed_vaa")
	signedVAA.Get("/:chain/:emitter/:sequence", vaaCtrl.FindSignedVAAByID)
	signedBatchVAA := apiV1.Group("/signed_batch_vaa")
	signedBatchVAA.Get("/:chain/:trxID/:nonce", vaaCtrl.FindSignedBatchVAAByID)

	// guardianSet resource
	guardianSet := apiV1.Group("/guardianset")
	guardianSet.Get("/current", guardianCtrl.GetGuardianSet)

	// heartbeats resource
	heartbeats := apiV1.Group("/heartbeats")
	heartbeats.Get("", heartbeatsCtrl.GetLastHeartbeats)

	// governor resource
	gov := apiV1.Group("/governor")
	gov.Get("/available_notional_by_chain", governorCtrl.GetAvailNotionByChain)
	gov.Get("/enqueued_vaas", governorCtrl.GetEnqueuedVaas)
	gov.Get("/is_vaa_enqueued/:chain/:emitter/:sequence", governorCtrl.IsVaaEnqueued)
	gov.Get("/token_list", governorCtrl.GetTokenList)
}
