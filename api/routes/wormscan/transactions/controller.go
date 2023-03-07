package transactions

import (
	"strconv"

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
// @Success 200 {object} []transactions.TransactionCountResult
// @Failure 400
// @Failure 500
// @Router /api/v1/last-txs [get]
func (c *Controller) GetLastTransactions(ctx *fiber.Ctx) error {
	timeSpan, err := middleware.ExtractTimeSpan(ctx, c.logger)
	if err != nil {
		return err
	}
	sampleRate, err := middleware.ExtractSampleRate(ctx, c.logger)
	if err != nil {
		return err
	}

	q := &transactions.TransactionCountQuery{
		TimeSpan:   timeSpan,
		SampleRate: sampleRate,
	}

	// Get transaction count.
	lastTrx, err := c.srv.GetTransactionCount(ctx.Context(), q)
	if err != nil {
		return err
	}

	return ctx.JSON(lastTrx)
}

// GetChainActivity godoc
// @Description Returns a list of tx by source chain and destination chain.
// @Tags Wormscan
// @ID x-chain-activity
// @Param start_time query string false "Star time (format: ISO-8601)."
// @Param end_time query string false "End time (format: ISO-8601)."
// @Param by query string false "Renders the results as notional or tx-count (default is notional)."
// @Param apps query string false "List of apps separated by comma (default is all apps)."
// @Success 200 {object} transactions.ChainActivity
// @Failure 400
// @Failure 500
// @Router /api/v1/x-chain-activity [get]
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
		c.logger.Error("Error getting chain activity", zap.Error(err))
		return err
	}

	// Convert the result to the expected format.
	txByChainID := make(map[int]*Tx)
	total := uint64(0)
	for _, item := range activity {
		chainSourceID, err := strconv.Atoi(item.ChainSourceID)
		if err != nil {
			c.logger.Error("Error during conversion of chainSourceId", zap.Error(err))
			return err
		}
		t, ok := txByChainID[chainSourceID]
		if !ok {
			destinations := make([]Destination, 0)
			t = &Tx{Chain: chainSourceID, Volume: 0, Percentage: 0, Destinations: destinations}
		}
		chainDestinationID, err := strconv.Atoi(item.ChainDestinationID)
		if err != nil {
			c.logger.Error("Error during conversion of chainDestinationId", zap.Error(err))
			return err
		}
		destination := Destination{Chain: chainDestinationID, Volume: item.Volume, Percentage: 0}
		t.Destinations = append(t.Destinations, destination)
		t.Volume += item.Volume
		txByChainID[chainSourceID] = t
		total += item.Volume
	}

	txs := make([]Tx, 0)
	for _, item := range txByChainID {
		if total > 0 {
			percentage := float64(item.Volume*100) / float64(total)
			item.Percentage = percentage
		}
		for i, destination := range item.Destinations {
			if item.Volume > 0 {
				item.Destinations[i].Percentage = float64(destination.Volume*100) / float64(item.Volume)
			}
		}
		txs = append(txs, *item)
	}

	return ctx.JSON(ChainActivity{Txs: txs})
}
