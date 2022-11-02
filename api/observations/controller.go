package observations

import (
	"github.com/certusone/wormhole/node/pkg/vaa"
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

func NewController(srv *Service, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    srv,
		logger: logger.With(zap.String("module", "ObservationsController")),
	}
}

func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	obs, err := c.srv.FindAll(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}

func (c *Controller) FindAllByChain(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, err := middleware.ExtractChainID(ctx)
	obs, err := c.srv.FindByChain(ctx.Context(), chainID, p)
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}

func (c *Controller) FindAllByEmitter(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, addr, _, err := middleware.ExtractVAAParams(ctx)
	if errors.IsOf(err, middleware.ErrMalformedChain, middleware.ErrMalformedAddr) {
		return err
	}
	obs, err := c.srv.FindByEmitter(ctx.Context(), chainID, addr, p)
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}

func (c *Controller) FindAllByVAA(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, addr, seq, err := middleware.ExtractVAAParams(ctx)
	if err != nil {
		return err
	}
	obs, err := c.srv.FindByVAA(ctx.Context(), chainID, addr, seq, p)
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}

func (c *Controller) FindOne(ctx *fiber.Ctx) error {
	chainID, addr, seq, err := middleware.ExtractVAAParams(ctx)
	if err != nil {
		return err
	}
	signer := ctx.Params("signer")
	signerAddr, err := vaa.StringToAddress(signer)
	if err != nil {
		return err
	}
	hash := ctx.Params("hash")
	obs, err := c.srv.FindOne(ctx.Context(), chainID, addr, seq, &signerAddr, []byte(hash))
	if err != nil {
		return err
	}
	return ctx.JSON(obs)
}
