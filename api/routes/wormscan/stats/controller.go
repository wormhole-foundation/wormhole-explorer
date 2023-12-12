package stats

import (
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/stats"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"go.uber.org/zap"
)

// Controller is the controller for the transactions resource.
type Controller struct {
	srv    *stats.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(statsService *stats.Service, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    statsService,
		logger: logger.With(zap.String("module", "StatsController")),
	}
}

// GetTopSymbolsByVolume godoc
// @Description Returns a list of symbols by origin chain and tokens.
// @Description The volume is calculated using the notional price of the symbol at the day the VAA was emitted.
// @Tags wormholescan
// @ID top-symbols-by-volume
// @Param timeSpan query string false "Time span, supported values: 7d, 15d and 30d (default is 7d)."
// @Success 200 {object} stats.TopSymbolByVolumeResult
// @Failure 400
// @Failure 500
// @Router /api/v1/top-symbols-by-volume [get]
func (c *Controller) GetTopSymbolsByVolume(ctx *fiber.Ctx) error {
	timeSpan, err := middleware.ExtractSymbolWithAssetsTimeSpan(ctx)
	if err != nil {
		return err
	}

	// Get the chain activity.
	assets, err := c.srv.GetSymbolWithAssets(ctx.Context(), *timeSpan)
	if err != nil {
		c.logger.Error("Error getting symbol with assets", zap.Error(err))
		return err
	}

	// Convert the result to the expected format.
	symbols, err := c.createTopSymbolsByVolumeResult(assets)
	if err != nil {
		return err
	}

	return ctx.JSON(TopSymbolByVolumeResult{Symbols: symbols})
}

func (c *Controller) createTopSymbolsByVolumeResult(assets []stats.SymbolWithAssetDTO) ([]*TopSymbolResult, error) {
	txByChainID := make(map[string]*TopSymbolResult)
	for _, item := range assets {
		t, ok := txByChainID[item.Symbol]
		if !ok {
			tokens := make([]TokenResult, 0)
			t = &TopSymbolResult{Symbol: item.Symbol, Volume: decimal.Zero, Txs: decimal.Zero, Tokens: tokens}
		}

		token := TokenResult{
			EmitterChainID: item.EmitterChainID,
			TokenChainID:   item.TokenChainID,
			TokenAddress:   item.TokenAddress,
			Volume:         item.Volume,
			Txs:            item.Txs}

		t.Tokens = append(t.Tokens, token)
		t.Volume = t.Volume.Add(item.Volume)
		t.Txs = t.Txs.Add(item.Txs)
		txByChainID[item.Symbol] = t
	}

	values := make([]*TopSymbolResult, 0, len(txByChainID))

	for _, value := range txByChainID {
		values = append(values, value)
	}

	sort.Slice(values[:], func(i, j int) bool {
		return values[i].Volume.GreaterThan(values[j].Volume)
	})

	if len(values) >= 7 {
		return values[:7], nil
	}
	return values, nil
}
