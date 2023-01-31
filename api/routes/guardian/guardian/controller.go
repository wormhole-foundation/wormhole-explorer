package guardian

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	gs     guardian.GuardianSet
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(logger *zap.Logger, p2pNetwork string) *Controller {
	return &Controller{gs: guardian.GetByEnv(p2pNetwork),
		logger: logger.With(zap.String("module", "GuardianController"))}
}

// GuardianSetResponse response definition.
type GuardianSetResponse struct {
	GuardianSet GuardianSet `json:"guardianSet"`
}

// GuardianSet response definition.
type GuardianSet struct {
	Index     uint32   `json:"index"`
	Addresses []string `json:"addresses"`
}

// GetGuardianSet godoc
// @Description Get current guardian set.
// @Tags Guardian
// @ID guardian-set
// @Success 200 {object} GuardianSetResponse
// @Failure 400
// @Failure 500
// @Router /v1/guardianset/current [get]
func (c *Controller) GetGuardianSet(ctx *fiber.Ctx) error {
	// check guardianSet exists.
	if len(c.gs.GstByIndex) == 0 {
		return response.NewApiError(ctx, fiber.StatusServiceUnavailable, response.Unavailable,
			"guardian set not fetched from chain yet", nil)
	}

	// get lasted guardianSet.
	guardinSet := c.gs.GetLatest()

	// get guardian addresses.
	addresses := make([]string, len(guardinSet.Keys))
	for i, v := range guardinSet.Keys {
		addresses[i] = v.Hex()
	}

	// create response.
	response := GuardianSetResponse{
		GuardianSet: GuardianSet{
			Index:     guardinSet.Index,
			Addresses: addresses,
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}
