package vaa

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/pagination"
	"go.uber.org/zap"
)

type Controller struct {
	srv    *Service
	logger *zap.Logger
}

func NewController(serv *Service, logger *zap.Logger) *Controller {
	return &Controller{srv: serv, logger: logger.With(zap.String("module", "VaaController"))}
}

func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	vaas, err := c.srv.FindAll(ctx.Context(), p)
	if err != nil {
		fmt.Printf("error finding vaas: %v", err)
	}
	return ctx.JSON(vaas)
}

func (c *Controller) FindByChain(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, err := middleware.ExtractChainID(ctx)
	if err != nil {
		return err
	}
	vaas, err := c.srv.FindByChain(ctx.Context(), chainID, p)
	if err != nil {
		fmt.Printf("error finding vaas: %v", err)
	}
	return ctx.JSON(vaas)
}

func (c *Controller) FindByEmitter(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, emitter, _, err := middleware.ExtractVAAParams(ctx)
	if errors.IsOf(err, middleware.ErrMalformedChain, middleware.ErrMalformedAddr) {
		return err
	}
	vaas, err := c.srv.FindByEmitter(ctx.Context(), chainID, *emitter, p)
	if err != nil {
		//TODO logging
		fmt.Printf("error finding vaas: %v", err)
	}
	return ctx.JSON(vaas)
}

func (c *Controller) FindById(ctx *fiber.Ctx) error {
	chainID, emitter, seq, err := middleware.ExtractVAAParams(ctx)
	if err != nil {
		return err
	}
	vaa, err := c.srv.FindById(ctx.Context(), chainID, *emitter, seq)
	if err != nil {
		//TODO logging
		fmt.Printf("error finding vaa: %v", err)
	}
	return ctx.JSON(vaa)
}

func (c *Controller) FindForPythnet(ctx *fiber.Ctx) error {
	return nil
}

func (c *Controller) GetStats(ctx *fiber.Ctx) error {
	stats, err := c.srv.GetVAAStats(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(stats)
}
