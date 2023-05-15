package transactions

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
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
// @Param timeSpan query string false "Time Span, default: 1d, supported values: [1d, 1w, 1mo]. 1mo ​​is 30 days."
// @Param sampleRate query string false "Sample Rate, default: 1h, supported values: [1h, 1d]. Valid configurations with timeSpan: 1d/1h, 1w/1d, 1mo/1d"
// @Success 200 {object} []transactions.TransactionCountResult
// @Failure 400
// @Failure 500
// @Router /api/v1/last-txs [get]
func (c *Controller) GetLastTransactions(ctx *fiber.Ctx) error {
	timeSpan, sampleRate, err := middleware.ExtractTimeSpanAndSampleRate(ctx, c.logger)
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

// GetScorecards godoc
// @Description Returns a list of KPIs for Wormhole.
// @Tags Wormscan
// @ID get-scorecards
// @Success 200 {object} ScorecardsResponse
// @Failure 500
// @Router /api/v1/scorecards [get]
func (c *Controller) GetScorecards(ctx *fiber.Ctx) error {

	// Query indicators from the database
	scorecards, err := c.srv.GetScorecards(ctx.Context())
	if err != nil {
		c.logger.Error("failed to get scorecards", zap.Error(err))
		return err
	}

	// Convert indicators to the response model
	response := ScorecardsResponse{
		TotalTxCount: scorecards.TotalTxCount,
		TotalVolume:  scorecards.TotalTxVolume,
		TxCount24h:   scorecards.TxCount24h,
		Volume24h:    scorecards.Volume24h,
	}

	return ctx.JSON(response)
}

// GetTopChainPairs godoc
// @Description Returns a list of the (emitter_chain, destination_chain) pairs with the highest number of transfers.
// @Tags Wormscan
// @ID get-top-chain-pairs-by-num-transfers
// @Param timeSpan query string true "Time span, supported values: 7d, 15d, 30d."
// @Success 200 {object} TopChainPairsResponse
// @Failure 500
// @Router /api/v1/top-chain-pairs-by-num-transfers [get]
func (c *Controller) GetTopChainPairs(ctx *fiber.Ctx) error {

	// Extract query parameters
	timeSpan, err := middleware.ExtractTopStatisticsTimeSpan(ctx)
	if err != nil {
		return err
	}

	// Query chain pairs from the database
	chainPairDTOs, err := c.srv.GetTopChainPairs(ctx.Context(), timeSpan)
	if err != nil {
		c.logger.Error("failed to get top chain pairs by number of transfers", zap.Error(err))
		return err
	}

	// Convert DTOs to the response model
	response := TopChainPairsResponse{
		ChainPairs: make([]ChainPair, 0, len(chainPairDTOs)),
	}
	for i := range chainPairDTOs {
		chainPair := ChainPair{
			EmitterChain:      chainPairDTOs[i].EmitterChain,
			DestinationChain:  chainPairDTOs[i].DestinationChain,
			NumberOfTransfers: chainPairDTOs[i].NumberOfTransfers,
		}
		response.ChainPairs = append(response.ChainPairs, chainPair)
	}

	return ctx.JSON(response)
}

// GetTopAssets godoc
// @Description Returns a list of the (emitter_chain, asset) pairs with the most volume.
// @Tags Wormscan
// @ID get-top-assets-by-volume
// @Param timeSpan query string true "Time span, supported values: 7d, 15d, 30d."
// @Success 200 {object} TopAssetsResponse
// @Failure 500
// @Router /api/v1/top-assets-by-volume [get]
func (c *Controller) GetTopAssets(ctx *fiber.Ctx) error {

	// Extract query parameters
	timeSpan, err := middleware.ExtractTopStatisticsTimeSpan(ctx)
	if err != nil {
		return err
	}

	// Query assets from the database
	assetDTOs, err := c.srv.GetTopAssets(ctx.Context(), timeSpan)
	if err != nil {
		c.logger.Error("failed to get top assets by volume", zap.Error(err))
		return err
	}

	// Convert DTOs to the response model
	response := TopAssetsResponse{
		Assets: make([]AssetWithVolume, 0, len(assetDTOs)),
	}
	for i := range assetDTOs {

		asset := AssetWithVolume{
			TokenChain:   assetDTOs[i].EmitterChain,
			TokenAddress: assetDTOs[i].TokenAddress,
			Volume:       assetDTOs[i].Volume,
		}

		// Look up the token symbol
		//
		// The explorer UI doesn't use this field, it uses the pair (tokenChain, tokenAddress) instead.
		// The symbol field is not strictly necessary, but it's nice to have it in the response.
		tokenMeta, ok := domain.GetTokenByAddress(assetDTOs[i].TokenChain, assetDTOs[i].TokenAddress)
		if ok {
			asset.Symbol = tokenMeta.Symbol
		}

		response.Assets = append(response.Assets, asset)
	}

	return ctx.JSON(response)
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
	startTime, endTime, err := middleware.ExtractTimeRange(ctx)
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
	txs, err := c.createChainActivityResponse(activity)
	if err != nil {
		return err
	}

	return ctx.JSON(ChainActivity{Txs: txs})
}

func (c *Controller) createChainActivityResponse(activity []transactions.ChainActivityResult) ([]Tx, error) {
	txByChainID := make(map[int]*Tx)
	total := decimal.Zero
	for _, item := range activity {
		chainSourceID, err := strconv.Atoi(item.ChainSourceID)
		if err != nil {
			c.logger.Error("Error during conversion of chainSourceId", zap.Error(err))
			return nil, err
		}
		t, ok := txByChainID[chainSourceID]
		if !ok {
			destinations := make([]Destination, 0)
			t = &Tx{Chain: chainSourceID, Volume: decimal.Zero, Percentage: 0, Destinations: destinations}
		}
		chainDestinationID, err := strconv.Atoi(item.ChainDestinationID)
		if err != nil {
			c.logger.Error("Error during conversion of chainDestinationId", zap.Error(err))
			return nil, err
		}
		volume, err := decimal.NewFromString(strconv.FormatUint(item.Volume, 10))
		if err != nil {
			c.logger.Error("Error during conversion of volume to decimal", zap.Error(err))
			return nil, err
		}
		destination := Destination{Chain: chainDestinationID, Volume: volume, Percentage: 0}
		t.Destinations = append(t.Destinations, destination)
		t.Volume = t.Volume.Add(volume)
		txByChainID[chainSourceID] = t
		total = total.Add(volume)
	}

	txs := make([]Tx, 0)
	oneHundred := decimal.NewFromInt(100)
	for _, item := range txByChainID {
		if total.GreaterThan(decimal.Zero) {
			percentage, _ := item.Volume.Div(total).Mul(oneHundred).Float64()
			item.Percentage = percentage
		}
		for i, destination := range item.Destinations {
			if item.Volume.GreaterThan(decimal.Zero) {
				percentage, _ := destination.Volume.Div(item.Volume).Mul(oneHundred).Float64()
				item.Destinations[i].Percentage = percentage
			}
		}
		txs = append(txs, *item)
	}
	return txs, nil
}

// FindGlobalTransactionByID godoc
// @Description Find a global transaction by ID.
// @Tags Wormscan
// @ID find-global-transaction-by-id
// @Param chain_id path integer true "id of the blockchain"
// @Param emitter path string true "address of the emitter"
// @Param seq path integer true "sequence of the VAA"
// @Success 200 {object} Tx
// @Failure 400
// @Failure 500
// @Router /api/v1/global-tx/{chain_id}/{emitter}/{seq} [get]
func (c *Controller) FindGlobalTransactionByID(ctx *fiber.Ctx) error {
	chainID, emitter, seq, err := middleware.ExtractVAAParams(ctx, c.logger)
	if err != nil {
		return err
	}

	globalTransaction, err := c.srv.FindGlobalTransactionByID(ctx.Context(), chainID, emitter, strconv.FormatUint(seq, 10))
	if err != nil {
		return err
	}

	return ctx.JSON(globalTransaction)
}
