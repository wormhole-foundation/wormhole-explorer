package transactions

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"go.uber.org/zap"
)

// Controller is the controller for the transactions resource.
type Controller struct {
	srv    *transactions.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(transactionsService *transactions.Service, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    transactionsService,
		logger: logger.With(zap.String("module", "TransactionsController")),
	}
}

// GetLastTrx godoc
func (c *Controller) GetLastTrx(ctx *fiber.Ctx) error {
	timeSpan, err := middleware.ExtractTimeSpan(ctx, c.logger)
	if err != nil {
		return err
	}
	sampleRate, err := middleware.ExtractSampleRate(ctx, c.logger)
	if err != nil {
		return err
	}

	// Get the last transactions.
	lastTrx, err := c.srv.GetLastTrx(timeSpan, sampleRate)
	if err != nil {
		return err
	}

	return ctx.JSON(lastTrx)
}

// GetChainActivity godoc
func (c *Controller) GetChainActivity(ctx *fiber.Ctx) error {
	startTime, err := middleware.ExtractTime(ctx, "start_time")
	if err != nil {
		return err
	}
	endTime, err := middleware.ExtractTime(ctx, "end_time")
	if err != nil {
		return err
	}

	apps, err := middleware.ExtractApps(ctx)
	if err != nil {
		return err
	}

	isNotional, err := middleware.ExtractIsNotional(ctx)
	if err != nil {
		return err
	}

	q := &transactions.ChainActivityQuery{
		Start:      startTime,
		End:        endTime,
		AppIDs:     apps,
		IsNotional: isNotional,
	}
	// Get the chain activity.
	activity, err := c.srv.GetChainActivity(ctx.Context(), q)
	if err != nil {
		return err
	}

	return ctx.JSON(activity)
}
