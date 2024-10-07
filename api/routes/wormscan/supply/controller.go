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

type CirculatingSupplyResponse struct {
	CirculatingSupply string `json:"circulating_supply"`
}

// GetCirculatingSupply godoc
// @Description Get W token circulation supply.
// @Tags wormholescan
// @ID supply
// @Success 200 {object} CirculatingSupplyResponse
// @Router /api/v1/supply [get]
func (c *Controller) GetCirculatingSupply(ctx *fiber.Ctx) error {
	supply := c.srv.GetCurrentCirculatingSupply(ctx.Context())
	return ctx.JSON(CirculatingSupplyResponse{
		CirculatingSupply: strconv.Itoa(supply.Unlocked),
	})
}

type TotalSupplyResponse struct {
	TotalSupply string `json:"total_supply"`
}

// GetTotalSupply godoc
// @Description Get W token total supply.
// @Tags wormholescan
// @ID total-supply
// @Success 200 {object} TotalSupplyResponse
// @Router /api/v1/total-supply [get]
func (c *Controller) GetTotalSupply(ctx *fiber.Ctx) error {
	supply := c.srv.GetTotalSupply(ctx.Context())

	return ctx.JSON(TotalSupplyResponse{
		TotalSupply: strconv.Itoa(supply),
	})
}
