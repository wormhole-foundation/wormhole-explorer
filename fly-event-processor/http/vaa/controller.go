package vaa

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/config"
	processor "github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/processor/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	dbLayer            string
	logger             *zap.Logger
	repository         *storage.Repository
	postgresRepository *storage.PostgresRepository
	processor          processor.ProcessorFunc
}

// NewController creates a Controller instance.
// func NewController(repository *Repository, processor processor.ProcessorFunc, logger *zap.Logger) *Controller {
func NewController(cfg *config.ServiceConfiguration,
	processor processor.ProcessorFunc,
	repository *storage.Repository,
	postgresRepository *storage.PostgresRepository,
	logger *zap.Logger) *Controller {
	return &Controller{
		dbLayer:            cfg.DbLayer,
		processor:          processor,
		repository:         repository,
		postgresRepository: postgresRepository,
		logger:             logger}
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

	params, err := c.getProcessorParams(ctx.Context(), request.VaaID)
	if err != nil {
		c.logger.Error("error getting processor params", zap.Error(err))
		return err
	}

	err = c.processor(ctx.Context(), params)
	if err != nil {
		c.logger.Error("error processing vaa", zap.Error(err))
		return err
	}

	return ctx.JSON(fiber.Map{"message": "success"})
}

func (c *Controller) getProcessorParams(ctx context.Context, vaaID string) (*processor.Params, error) {

	switch c.dbLayer {
	case config.DbLayerPostgres:
		attestationVaa, err := c.postgresRepository.FindAttestationVaaByVaaId(ctx, vaaID)
		if err != nil {
			c.logger.Error("error getting attestation vaas", zap.Error(err))
			return nil, err
		}
		return &processor.Params{
			TrackID: fmt.Sprintf("controller-%s", vaaID),
			VaaID:   vaaID,
			ChainID: attestationVaa[0].EmitterChain,
		}, nil
	case config.DbLayerMongo:
		vaa, err := c.repository.FindVAAById(ctx, vaaID)
		if err != nil {
			c.logger.Error("error getting vaa from collection", zap.Error(err))
			return nil, err
		}

		return &processor.Params{
			TrackID: fmt.Sprintf("controller-%s", vaaID),
			VaaID:   vaaID,
			ChainID: vaa.EmitterChain,
		}, nil
	case config.DbLayerDual:
		vaa, err := c.repository.FindVAAById(ctx, vaaID)
		if err != nil {
			c.logger.Error("error getting vaa from collection", zap.Error(err))
			return nil, err
		}

		return &processor.Params{
			TrackID: fmt.Sprintf("controller-%s", vaaID),
			VaaID:   vaaID,
			ChainID: vaa.EmitterChain,
		}, nil
	}
	return nil, errors.New("invalid db layer")
}
