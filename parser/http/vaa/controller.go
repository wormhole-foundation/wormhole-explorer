package vaa

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/parser/config"
	"github.com/wormhole-foundation/wormhole-explorer/parser/processor"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	dbMode             string
	logger             *zap.Logger
	repository         *Repository
	postgresRepository *PostgresRepository
	processor          processor.ProcessorFunc
}

// NewController creates a Controller instance.
func NewController(dbMode string, repository *Repository, postgresRepository *PostgresRepository,
	processor processor.ProcessorFunc, logger *zap.Logger) *Controller {
	return &Controller{
		dbMode:             dbMode,
		repository:         repository,
		postgresRepository: postgresRepository,
		processor:          processor,
		logger:             logger}
}

func (c *Controller) Parse(ctx *fiber.Ctx) error {
	payload := struct {
		ID string `json:"id"`
	}{}

	if err := ctx.BodyParser(&payload); err != nil {
		return err
	}

	c.logger.Info("Parsing VAA from endpoint", zap.String("id", payload.ID))

	rawVaa, err := c.findByVaaId(ctx.Context(), payload.ID)
	if err != nil {
		return err
	}

	trackID := fmt.Sprintf("controller-%s", payload.ID)

	vaaParsed, err := c.processor(ctx.Context(), &processor.Params{Vaa: rawVaa, Source: "controller", TrackID: trackID})
	if err != nil {
		return err
	}

	return ctx.JSON(struct {
		Result any `json:"result"`
	}{Result: vaaParsed})
}

func (c *Controller) findByVaaId(ctx context.Context, vaaId string) ([]byte, error) {
	switch c.dbMode {
	case config.DbLayerMongo:
		vaa, err := c.repository.FindById(ctx, vaaId)
		if err != nil {
			return nil, err
		}
		return vaa.Vaa, nil
	case config.DbLayerPostgres:
		attestationVaa, err := c.postgresRepository.FindActiveAttestationVaaByVaaID(ctx, vaaId)
		if err != nil {
			return nil, err
		}
		return attestationVaa.Raw, nil
	case config.DbLayerDual:
		vaa, err := c.repository.FindById(ctx, vaaId)
		if err != nil {
			attestationVaa, err := c.postgresRepository.FindActiveAttestationVaaByVaaID(ctx, vaaId)
			if err != nil {
				return nil, err
			}
			return attestationVaa.Raw, nil
		}
		return vaa.Vaa, nil
	default:
		return nil, fmt.Errorf("unknown db mode: %s", c.dbMode)
	}
}
