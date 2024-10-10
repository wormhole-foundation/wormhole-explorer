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

// GetCirculatingSupply godoc
// @Description Get W token circulation supply.
// @Tags wormholescan
// @ID circulating-supply
// @Success 200
// @Router /api/v1/supply/circulating [get]
func (c *Controller) GetCirculatingSupply(ctx *fiber.Ctx) error {
	supply := c.srv.GetCurrentCirculatingSupply(ctx.Context())
	ctx.SendString(strconv.Itoa(supply.Unlocked))
	return nil
}

// GetTotalSupply godoc
// @Description Get W token total supply.
// @Tags wormholescan
// @ID total-supply
// @Success 200
// @Router /api/v1/supply/total [get]
func (c *Controller) GetTotalSupply(ctx *fiber.Ctx) error {
	supply := c.srv.GetTotalSupply(ctx.Context())
	ctx.SendString(strconv.Itoa(supply))
	return nil
}

type SupplyInfoResponse struct {
	CirculatingSupply string `json:"circulating_supply"`
	TotalSupply       string `json:"total_supply"`
}

// GetSupplyInfo godoc
// @Description Get W token supply data (circulation and total supply).
// @Tags wormholescan
// @ID supply-info
// @Success 200 {object} SupplyInfoResponse
// @Router /api/v1/supply [get]
func (c *Controller) GetSupplyInfo(ctx *fiber.Ctx) error {
	supply := c.srv.GetSupplyInfo(ctx.Context())
	circulationSupply := SupplyInfoResponse{
		CirculatingSupply: strconv.Itoa(supply.CirculatingSupply.Unlocked),
		TotalSupply:       strconv.Itoa(supply.TotalSupply),
	}
	return ctx.JSON(circulationSupply)
}
