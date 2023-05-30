// Package observations handle the request of observations data from governor endpoint defined in the api.
package observations

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/observations"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *observations.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(srv *observations.Service, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    srv,
		logger: logger.With(zap.String("module", "ObservationsController")),
	}
}

// FindAll godoc
// @Description Returns all observations, sorted by descending timestamp.
// @Tags Wormscan
// @ID find-observations
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} []ObservationDoc
// @Failure 400
// @Failure 500
// @Router /api/v1/observations [get]
func (c *Controller) FindAll(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	obs, err := c.srv.FindAll(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(obs)
}

// FindAllByChain godoc
// @Description Returns all observations for a given blockchain, sorted by descending timestamp.
// @Tags Wormscan
// @ID find-observations-by-chain
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} []ObservationDoc
// @Failure 400
// @Failure 500
// @Router /api/v1/observations/:chain [get]
func (c *Controller) FindAllByChain(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

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

// FindAllByEmitter godoc
// @Description Returns all observations for a specific emitter address, sorted by descending timestamp.
// @Tags Wormscan
// @ID find-observations-by-emitter
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} []ObservationDoc
// @Failure 400
// @Failure 500
// @Router /api/v1/observations/:chain/:emitter [get]
func (c *Controller) FindAllByEmitter(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

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

// FindAllByVAA godoc
// @Description Find observations identified by emitter chain, emitter address and sequence.
// @Tags Wormscan
// @ID find-observations-by-sequence
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} []ObservationDoc
// @Failure 400
// @Failure 500
// @Router /api/v1/observations/:chain/:emitter/:sequence [get]
func (c *Controller) FindAllByVAA(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

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

// FindOne godoc
// @Description Find a specific observation.
// @Tags Wormscan
// @ID find-observations-by-id
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} []ObservationDoc
// @Failure 400
// @Failure 500
// @Router /api/v1/observations/:chain/:emitter/:sequence/:signer/:hash [get]
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
