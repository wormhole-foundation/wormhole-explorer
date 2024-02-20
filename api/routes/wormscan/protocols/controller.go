package protocols

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/protocols"
	"go.uber.org/zap"
)

type Controller struct {
	srv    service
	logger *zap.Logger
}

type service interface {
	GetProtocolsTotalValues(ctx context.Context) []protocols.ProtocolTotalValuesDTO
}

func NewController(logger *zap.Logger, service service) *Controller {
	return &Controller{
		logger: logger.With(zap.String("module", "ContributorsController")),
		srv:    service,
	}
}

// GetProtocolsTotalValues godoc
// @Description Returns the representative stats for the top protocols
// @Tags wormholescan
// @ID get-top-protocols-stats
// @Success 200 {object} []protocols.ProtocolTotalValuesDTO
// @Failure 500 {object} []protocols.ProtocolTotalValuesDTO
// @Router /api/v1/protocols/stats [get]
func (c *Controller) GetProtocolsTotalValues(ctx *fiber.Ctx) error {
	values := c.srv.GetProtocolsTotalValues(ctx.Context())
	allFailed := true
	for i := range values {
		allFailed = allFailed && len(values[i].Error) > 0
	}

	err := ctx.JSON(values)
	if allFailed && len(values) > 0 {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	return err
}
