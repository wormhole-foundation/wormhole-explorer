package stats

import (
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/stats"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
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

// GetTop100Corridors godoc
// @Description Returns a list of the top 100 tokens, sorted in descending order by the number of transactions.
// @Tags wormholescan
// @ID /api/v1/top-100-corridors
// @Param timeSpan query string false "Time span, supported values: 2d and 7d (default is 2d)."
// @Success 200 {object} stats.TopCorridorsResult
// @Failure 400
// @Failure 500
// @Router /api/v1/top-100-corridors [get]
func (c *Controller) GetTopCorridors(ctx *fiber.Ctx) error {
	timeSpan, err := middleware.ExtractTopCorridorsTimeSpan(ctx)
	if err != nil {
		return err
	}

	// Get the chain activity.
	corridors, err := c.srv.GetTopCorridors(ctx.Context(), *timeSpan)
	if err != nil {
		c.logger.Error("Error getting symbol with assets", zap.Error(err))
		return err
	}

	result := createTop100CorridorsResult(corridors)

	return ctx.JSON(TopCorridorsResult{Corridors: result})
}

func createTop100CorridorsResult(corridors []stats.TopCorridorsDTO) []*TopCorridor {

	result := make([]*TopCorridor, 0, len(corridors))
	for _, c := range corridors {
		result = append(result, &TopCorridor{
			EmitterChainID: c.EmitterChainID,
			TargetChainID:  c.DestinationChainID,
			TokenChainID:   c.TokenChainID,
			TokenAddress:   c.TokenAddress,
			Txs:            c.Txs,
		})
	}
	sort.Slice(result[:], func(i, j int) bool {
		return result[i].Txs > result[j].Txs
	})

	if len(result) >= 100 {
		return result[:100]
	}
	return result
}

// GetNativeTokenTransferSummary godoc
// @Description Returns a summary of the Native Token Transfer.
// @Tags wormholescan
// @ID /api/v1/native-token-transfer/summary
// @Param symbol query string true "Symbol of the token. Currently only supports W."
// @Success 200 {object} stats.NativeTokenTransferSummary
// @Failure 400
// @Failure 500
// @Router /api/v1/native-token-transfer/summary [get]
func (c *Controller) GetNativeTokenTransferSummary(ctx *fiber.Ctx) error {
	symbolParam, err := middleware.ExtractSymbol(ctx)
	if err != nil {
		return err
	}

	symbol := strings.ToUpper(symbolParam)
	response, err := c.srv.GetNativeTokenTransferSummary(ctx.Context(), symbol)
	if err != nil {
		return err
	}

	return ctx.JSON(response)
}

// GetNativeTokenTransferActivity godoc
// @Description Returns a list of values (tx count or notional) of the Native Token Transfer for a emitter and destination chains.
// @Tags wormholescan
// @ID /api/v1/native-token-transfer/activity
// @Param symbol query string true "Symbol of the token. Currently only supports W."
// @Param by query string false "Renders the results using notional or tx count (default is notional)."
// @Success 200 {object} []stats.NativeTokenTransferActivity
// @Failure 400
// @Failure 500
// @Router /api/v1/native-token-transfer/activity [get]
func (c *Controller) GetNativeTokenTransferActivity(ctx *fiber.Ctx) error {
	symbolParam, err := middleware.ExtractSymbol(ctx)
	if err != nil {
		return err
	}

	isNotional, err := middleware.ExtractIsNotional(ctx)
	if err != nil {
		isNotional = true
	}
	symbol := strings.ToUpper(symbolParam)
	response, err := c.srv.GetNativeTokenTransferActivity(ctx.Context(), isNotional, strings.ToUpper(symbol))
	if err != nil {
		return err
	}

	return ctx.JSON(response)
}

// GetNativeTokenTransferByTime godoc
// @Description Returns a list of values (tx count or notional) of the Native Token Transfer for a emitter and destination chains.
// @Tags wormholescan
// @ID /api/v1/native-token-transfer/transfer-by-time
// @Param from query string true "From date, supported format 2006-01-02T15:04:05Z07:00"
// @Param to query string true "To date, supported format 2006-01-02T15:04:05Z07:00"
// @Param symbol query string true "Symbol of the token. Currently only supports W."
// @Param by query string false "Renders the results using notional or tx count (default is notional)."
// @Param timeSpan query string true "Time Span, supported values: [1h, 1d, 1mo, 1y]."
// @Success 200 {object} []stats.NativeTokenTransferByTime
// @Failure 400
// @Failure 500
// @Router /api/v1/native-token-transfer/transfer-by-time [get]
func (c *Controller) GetNativeTokenTransferByTime(ctx *fiber.Ctx) error {
	from, err := middleware.ExtractTime(ctx, time.RFC3339, "from")
	if err != nil {
		return err
	}
	to, err := middleware.ExtractTime(ctx, time.RFC3339, "to")
	if err != nil {
		return err
	}
	if from == nil || to == nil {
		return response.NewInvalidParamError(ctx, "missing from/to query params ", nil)
	}
	symbolParam, err := middleware.ExtractSymbol(ctx)
	if err != nil {
		return err
	}
	isNotional, err := middleware.ExtractIsNotional(ctx)
	if err != nil {
		isNotional = true
	}
	timespan, err := middleware.ExtractNttTimeSpan(ctx)
	if err != nil {
		return err
	}
	symbol := strings.ToUpper(symbolParam)
	response, err := c.srv.GetNativeTokenTransferByTime(ctx.Context(), *timespan, symbol, isNotional, *from, *to)
	if err != nil {
		return err
	}

	return ctx.JSON(response)
}

// GetNativeTokenTransferAddressTop godoc
// @Description Returns a list of values (tx count or notional) of the Native Token Transfer for address.
// @Tags wormholescan
// @ID /api/v1/native-token-transfer/top-address
// @Param symbol query string true "Symbol of the token. Currently only supports W."
// @Param by query string false "Renders the results using notional or tx count (default is notional)."
// @Success 200 {object} []stats.NativeTokenTransferTopAddress
// @Failure 400
// @Failure 500
// @Router /api/v1/native-token-transfer/top-address [get]
func (c *Controller) GetNativeTokenTransferAddressTop(ctx *fiber.Ctx) error {
	symbolParam, err := middleware.ExtractSymbol(ctx)
	if err != nil {
		return err
	}
	isNotional, err := middleware.ExtractIsNotional(ctx)
	if err != nil {
		isNotional = true
	}
	symbol := strings.ToUpper(symbolParam)
	response, err := c.srv.GetNativeTokenTransferAddressTop(ctx.Context(), symbol, isNotional)
	if err != nil {
		return err
	}

	return ctx.JSON(response)
}

func (c *Controller) GetNativeTokenTransferTopHolder(ctx *fiber.Ctx) error {
	symbolParam, err := middleware.ExtractSymbol(ctx)
	if err != nil {
		return err
	}

	symbol := strings.ToUpper(symbolParam)
	holders, err := c.srv.GetTopHolder(ctx.Context(), symbol)
	if err != nil {
		return err
	}

	return ctx.JSON(holders)
}
