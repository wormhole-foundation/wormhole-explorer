package relays

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/relays"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *relays.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(srv *relays.Service, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    srv,
		logger: logger.With(zap.String("module", "RelaysController")),
	}
}

// FindByVAA godoc
// @Description Get a specific relay information by chainID, emitter address and sequence.
// @Tags wormholescan
// @ID find-relay-by-vaa-id
// @Success 200 {object} []observations.ObservationDoc
// @Failure 400
// @Failure 500
// @Router /api/v1/relays/:chain/:emitter/:sequence [get]
func (c *Controller) FindOne(ctx *fiber.Ctx) error {
	chainID, addr, seq, err := middleware.ExtractVAAParams(ctx, c.logger)
	if err != nil {
		return err
	}
	obs, err := c.srv.FindByVAA(ctx.Context(), chainID, addr, strconv.FormatUint(seq, 10))
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}
