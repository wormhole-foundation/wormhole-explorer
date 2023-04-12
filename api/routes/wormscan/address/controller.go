package address

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/address"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware" // required by swaggo
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
// @Tags Wormscan
// @ID find-address-by-id
// @Param address path string true "address"
// @Param page query integer false "Page number. Starts at 0."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} response.Response[address.AddressOverview]
// @Failure 400
// @Failure 404
// @Failure 500
// @Router /api/v1/address/{address} [get]
func (c *Controller) FindById(ctx *fiber.Ctx) error {

	address, err := middleware.ExtractAddress(ctx, c.logger)
	if err != nil {
		return err
	}

	pagination, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
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
