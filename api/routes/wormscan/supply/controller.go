package supply

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/supply"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *supply.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(serv *supply.Service, logger *zap.Logger) *Controller {
	return &Controller{srv: serv, logger: logger.With(zap.String("module", "CirculatingSupplyService"))}
}

func (c *Controller) GetCirculatingSupply(ctx *fiber.Ctx) error {
	supply := c.srv.GetCurrentCirculatingSupply(ctx.Context())
	ctx.SendString(strconv.Itoa(supply.Unlocked))
	return nil
}

func (c *Controller) GetTotalSupply(ctx *fiber.Ctx) error {
	supply := c.srv.GetTotalSupply(ctx.Context())
	ctx.SendString(strconv.Itoa(supply))
	return nil
}
