package guardian

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(logger *zap.Logger) *Controller {
	return &Controller{logger: logger.With(zap.String("module", "GuardianController"))}
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

// GetGuardianSet handler for the endpoint /guardian_public_api/v1/guardianset/current
// This endpoint has been migrated from the guardian grpc api.
func (c *Controller) GetGuardianSet(ctx *fiber.Ctx) error {
	// check guardianSet exists.
	if len(ByIndex) == 0 {
		return response.NewApiError(ctx, fiber.StatusServiceUnavailable, response.Unavailable,
			"guardian set not fetched from chain yet", nil)
	}
	// get lasted guardianSet.
	guardinSet := GetLatest()

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
