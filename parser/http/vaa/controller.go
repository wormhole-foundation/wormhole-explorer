package vaa

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/parser/processor"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	logger     *zap.Logger
	repository *Repository
	processor  processor.ProcessorFunc
}

// NewController creates a Controller instance.
func NewController(repository *Repository, processor processor.ProcessorFunc, logger *zap.Logger) *Controller {
	return &Controller{repository: repository, processor: processor, logger: logger}
}

func (c *Controller) Parse(ctx *fiber.Ctx) error {
	payload := struct {
		ID string `json:"id"`
	}{}

	if err := ctx.BodyParser(&payload); err != nil {
		return err
	}

	c.logger.Info("Parsing VAA from endpoint", zap.String("id", payload.ID))

	vaa, err := c.repository.FindById(ctx.Context(), payload.ID)
	if err != nil {
		return err
	}

	vaaParsed, err := c.processor(ctx.Context(), vaa.Vaa)
	if err != nil {
		return err
	}

	return ctx.JSON(struct {
		Result any `json:"result"`
	}{Result: vaaParsed})
}
