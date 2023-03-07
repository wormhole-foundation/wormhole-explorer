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

// GetLastTransactions godoc
// @Description Returns the number of transactions [vaa] by a defined time span and sample rate.
// @Tags Wormscan
// @ID get-last-transactions
// @Param timeSpan query string false "Time Span, default: 1h, examples: 30m, 1h, 1d, 2w, 3mo, 1y, all."
// @Param sampleRate query string false "Sample Rate, default: 1m, examples: 30s, 1m, 1h, 1d, 2w, 3mo, 1y."
// @Param cumulativeSum query boolean false "Cumulative Sum, fill empty values with cumulative sum, default: false, examples: true, false."
// @Success 200 {object} []transactions.TransactionCountResult
// @Failure 400
// @Failure 500
// @Router /api/v1/last-trx [get]
func (c *Controller) GetLastTransactions(ctx *fiber.Ctx) error {
	timeSpan, err := middleware.ExtractTimeSpan(ctx, c.logger)
	if err != nil {
		return err
	}
	sampleRate, err := middleware.ExtractSampleRate(ctx, c.logger)
	if err != nil {
		return err
	}
	cumulativeSum, _ := middleware.ExtractCumulativeSum(ctx, c.logger)

	q := &transactions.TransactionCountQuery{
		TimeSpan:      timeSpan,
		SampleRate:    sampleRate,
		CumulativeSum: cumulativeSum,
	}

	// Get transaction count.
	lastTrx, err := c.srv.GetTransactionCount(ctx.Context(), q)
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
