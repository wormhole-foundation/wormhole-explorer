package contributors

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"go.uber.org/zap"
)

type Controller struct {
	//srv    *stats.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(logger *zap.Logger) *Controller {
	return &Controller{
		logger: logger.With(zap.String("module", "ContributorsController")),
	}
}

func (c *Controller) GetContributorsTotalValues(ctx *fiber.Ctx) error {
	timeSpan, err := middleware.ExtractSymbolWithAssetsTimeSpan(ctx)
	if err != nil {
		return err
	}

	return ctx.JSON(TopSymbolByVolumeResult{Symbols: symbols})
}
