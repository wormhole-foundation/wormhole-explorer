package vaa

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	logger     *zap.Logger
	repository *storage.Repository
	processor  processor.ProcessorFunc
}

// NewController creates a Controller instance.
// func NewController(repository *Repository, processor processor.ProcessorFunc, logger *zap.Logger) *Controller {
func NewController(processor processor.ProcessorFunc, repository *storage.Repository, logger *zap.Logger) *Controller {
	return &Controller{processor: processor, repository: repository, logger: logger}
}

// Process processes the VAA message.
func (c *Controller) Process(ctx *fiber.Ctx) error {
	request := struct {
		VaaID string `json:"vaaId"`
	}{}

	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("error parsing request", zap.Error(err))
		return err
	}

	c.logger.Info("processing duplicated vaa in controller", zap.String("vaaId", request.VaaID))

	vaa, err := c.repository.FindVAAById(ctx.Context(), request.VaaID)
	if err != nil {
		c.logger.Error("error getting vaa from collection", zap.Error(err))
		return err
	}

	params := processor.Params{
		TrackID: fmt.Sprintf("controller-%s", request.VaaID),
		VaaID:   request.VaaID,
		ChainID: vaa.EmitterChain,
	}

	err = c.processor(ctx.Context(), &params)
	if err != nil {
		c.logger.Error("error processing vaa", zap.Error(err))
		return err
	}

	return ctx.JSON(fiber.Map{"message": "success"})
}
