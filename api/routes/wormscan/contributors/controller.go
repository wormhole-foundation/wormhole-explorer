package contributors

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/contributors"
	"go.uber.org/zap"
)

type Controller struct {
	srv    service
	logger *zap.Logger
}

type service interface {
	GetContributorsTotalValues(ctx context.Context) []contributors.ContributorTotalValuesDTO
}

func NewController(logger *zap.Logger, service service) *Controller {
	return &Controller{
		logger: logger.With(zap.String("module", "ContributorsController")),
		srv:    service,
	}
}

// GetTopContributors godoc
// @Description Returns the representative stats for the top contributors
// @Tags wormholescan
// @ID get-top-contributors-stats
// @Success 200 {object} []contributors.ContributorTotalValuesDTO
// @Failure 500 {object} []contributors.ContributorTotalValuesDTO
// @Router /api/v1/contributors/stats [get]
func (c *Controller) GetContributorsTotalValues(ctx *fiber.Ctx) error {
	values := c.srv.GetContributorsTotalValues(ctx.Context())
	allFailed := true
	for i := range values {
		allFailed = allFailed && values[i].Error != nil
	}

	err := ctx.JSON(values)
	if allFailed {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	return err
}
