// Package observations handle the request of observations data from governor endpoint defined in the api.
package observations

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(srv *Service, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    srv,
		logger: logger.With(zap.String("module", "ObservationsController")),
	}
}

// FindAll handler for the endpoint /observations/.
func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
	obs, err := c.srv.FindAll(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}

// FindAllByChain handler for the endpoint /observations/:chain.
func (c *Controller) FindAllByChain(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
	chainID, err := middleware.ExtractChainID(ctx, c.logger)
	if err != nil {
		return err
	}
	obs, err := c.srv.FindByChain(ctx.Context(), chainID, p)
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}

// FindAllByEmitter handler for the endpoint /observations/:chain/:emitter.
func (c *Controller) FindAllByEmitter(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
	chainID, addr, err := middleware.ExtractVAAChainIDEmitter(ctx, c.logger)
	if err != nil {
		return err
	}

	obs, err := c.srv.FindByEmitter(ctx.Context(), chainID, addr, p)
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}

// FindAllByVAA handler for the endpoint  /observations/:chain/:emitter/:sequence
func (c *Controller) FindAllByVAA(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
	chainID, addr, seq, err := middleware.ExtractVAAParams(ctx, c.logger)
	if err != nil {
		return err
	}

	obs, err := c.srv.FindByVAA(ctx.Context(), chainID, addr, strconv.FormatUint(seq, 10), p)
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}

// FindOne handler for the endpoint /observations/:chain/:emitter/:sequence/:signer/:hash
func (c *Controller) FindOne(ctx *fiber.Ctx) error {
	chainID, addr, seq, err := middleware.ExtractVAAParams(ctx, c.logger)
	if err != nil {
		return err
	}
	signerAddr, err := middleware.ExtractObservationSigner(ctx, c.logger)
	if err != nil {
		return err
	}
	hash, err := middleware.ExtractObservationHash(ctx, c.logger)
	if err != nil {
		return err
	}
	obs, err := c.srv.FindOne(ctx.Context(), chainID, addr, strconv.FormatUint(seq, 10), signerAddr, []byte(hash))
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}
