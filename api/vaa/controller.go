// Package observations handle the request of VAA data from governor endpoint defined in the api.
package vaa

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/pagination"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(serv *Service, logger *zap.Logger) *Controller {
	return &Controller{srv: serv, logger: logger.With(zap.String("module", "VaaController"))}
}

// FindAll handler for the endpoint /vaas/.
func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	vaas, err := c.srv.FindAll(ctx.Context(), p)
	if err != nil {
		fmt.Printf("error finding vaas: %v", err)
	}
	return ctx.JSON(vaas)
}

// FindByChain handler for the endpoint /vaas/:chain.
func (c *Controller) FindByChain(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, err := middleware.ExtractChainID(ctx)
	if err != nil {
		return err
	}
	vaas, err := c.srv.FindByChain(ctx.Context(), chainID, p)
	if err != nil {
		fmt.Printf("error finding vaas: %v", err)
		return err
	}
	return ctx.JSON(vaas)
}

// FindByEmitter handler for the endpoint /vaas/:chain/:emitter.
func (c *Controller) FindByEmitter(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, emitter, err := middleware.ExtractVAAChainIDEmitter(ctx)
	if err != nil {
		return err
	}
	vaas, err := c.srv.FindByEmitter(ctx.Context(), chainID, *emitter, p)
	if err != nil {
		//TODO logging
		fmt.Printf("error finding vaas: %v", err)
	}
	return ctx.JSON(vaas)
}

// FindById handler for the endpoint /vaas/:chain/:emitter/:sequence/:signer/:hash.
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

// FindSignedVAAByID get a signedVAA []byte from a chainID, emitter address and sequence.
func (c *Controller) FindSignedVAAByID(ctx *fiber.Ctx) error {
	chainID, emitter, seq, err := middleware.ExtractVAAParams(ctx)
	if err != nil {
		// TODO: Handle InvalidArgument code (3) in the response.
		return err
	}
	vaa, err := c.srv.FindById(ctx.Context(), chainID, *emitter, seq)
	if err != nil {
		// TODO: handle not found(5) and internal(13) response code.
		return err
	}
	response := struct {
		VaaBytes []byte `json:"vaaBytes"`
	}{
		VaaBytes: vaa.Data.Vaa,
	}
	return ctx.JSON(response)
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
