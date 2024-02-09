package address

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/address"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware" // required by swaggo
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	_ "github.com/wormhole-foundation/wormhole-explorer/api/response" // required by swaggo
	"go.uber.org/zap"
)

type Controller struct {
	srv    *address.Service
	logger *zap.Logger
}

func NewController(srv *address.Service, logger *zap.Logger) *Controller {

	c := Controller{
		srv:    srv,
		logger: logger.With(zap.String("module", "AddressController")),
	}

	return &c
}

// FindById godoc
// @Description Lookup an address
// @Tags wormholescan
// @ID find-address-by-id
// @Param address path string true "address"
// @Param page query integer false "Page number. Starts at 0."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} response.Response[address.AddressOverview]
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /api/v1/address/:address [get]
func (c *Controller) FindById(ctx *fiber.Ctx) error {

	address := middleware.ExtractAddressFromPath(ctx, c.logger)

	pagination, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if pagination.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	response, err := c.srv.GetAddressOverview(ctx.Context(), address, pagination)
	if err != nil {
		return err
	}
	if len(response.Data.Vaas) == 0 {
		return errors.ErrNotFound
	}

	return ctx.JSON(response)
}
