package transactions

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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
// @Description Returns the number of transactions by a defined time span and sample rate.
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
// @Description TVL is total value locked by token bridge contracts in USD.
// @Description Volume is the all-time total volume transferred through the token bridge in USD.
// @Description 24h volume is the volume transferred through the token bridge in the last 24 hours, in USD.
// @Description Total Tx count is the number of transaction bridging assets since the creation of the network (does not include Pyth or other messages).
// @Description 24h tx count is the number of transaction bridging assets in the last 24 hours (does not include Pyth or other messages).
// @Description Total messages is the number of VAAs emitted since the creation of the network (includes Pyth messages).
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
		Messages24h:  scorecards.Messages24h,
		TotalTxCount: scorecards.TotalTxCount,
		TotalVolume:  scorecards.TotalTxVolume,
		Tvl:          scorecards.Tvl,
		TxCount24h:   scorecards.TxCount24h,
		Volume24h:    scorecards.Volume24h,
	}

	return ctx.JSON(response)
}

// GetTopChainPairs godoc
// @Description Returns a list of the emitter_chain and destination_chain pair ordered by transfer count.
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
// @Description Returns a list of emitter_chain and asset pairs with ordered by volume.
// @Description The volume is calculated using the notional price of the symbol at the day the VAA was emitted.
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
			EmitterChain: assetDTOs[i].EmitterChain,
			TokenChain:   assetDTOs[i].TokenChain,
			TokenAddress: assetDTOs[i].TokenAddress,
			Volume:       assetDTOs[i].Volume,
		}

		// Look up the token symbol
		tokenMeta, ok := domain.GetTokenByAddress(assetDTOs[i].TokenChain, assetDTOs[i].TokenAddress)
		if ok {
			asset.Symbol = tokenMeta.Symbol.String()
		}

		response.Assets = append(response.Assets, asset)
	}

	return ctx.JSON(response)
}

// GetChainActivity godoc
// @Description Returns a list of chain pairs by origin chain and destination chain.
// @Description The list could be rendered by volume or transaction count.
// @Description The volume is calculated using the notional price of the symbol at the day the VAA was emitted.
// @Tags Wormscan
// @ID x-chain-activity
// @Param start_time query string false "Star time (format: ISO-8601)."
// @Param end_time query string false "End time (format: ISO-8601)."
// @Param by query string false "Renders the results using volume or tx count (default is volume)."
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
	txs, err := c.createChainActivityResponse(activity, isNotional)
	if err != nil {
		return err
	}

	return ctx.JSON(ChainActivity{Txs: txs})
}

func (c *Controller) createChainActivityResponse(activity []transactions.ChainActivityResult, isNotional bool) ([]Tx, error) {
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
		if isNotional {
			item.Volume = convertToDecimal(item.Volume)
		}
		for i, destination := range item.Destinations {
			if item.Volume.GreaterThan(decimal.Zero) {
				percentage, _ := destination.Volume.Div(item.Volume).Mul(oneHundred).Float64()
				item.Destinations[i].Percentage = percentage
			}
			if isNotional {
				item.Destinations[i].Volume = convertToDecimal(destination.Volume)
			}
		}
		txs = append(txs, *item)
	}
	return txs, nil
}

// FindGlobalTransactionByID godoc
// @Description Find a global transaction by VAA ID
// @Description Global transactions is a logical association of two transactions that are related to each other by a unique VAA ID.
// @Description The first transaction is created on the origin chain when the VAA is emitted.
// @Description The second transaction is created on the destination chain when the VAA is redeemed.
// @Description If the response only contains an origin tx the VAA was not redeemed.
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

func convertToDecimal(amount decimal.Decimal) decimal.Decimal {
	eigthDecimals := decimal.NewFromInt(1_0000_0000)
	return amount.Div(eigthDecimals)
}

// GetTokenByChainAndAddress godoc
// @Description Returns a token symbol, coingecko id and address by chain and token address.
// @Tags Wormscan
// @ID get-token-by-chain-and-address
// @Param chain_id path integer true "id of the blockchain"
// @Param token_address path string true "token address"
// @Success 200 {object} Token
// @Failure 400
// @Failure 404
// @Router /api/v1/token/{chain_id}/{token_address} [get]
func (c *Controller) GetTokenByChainAndAddress(ctx *fiber.Ctx) error {
	chain, err := middleware.ExtractChainID(ctx, c.logger)
	if err != nil {
		return err
	}

	tokenAddress, err := middleware.ExtractTokenAddress(ctx, c.logger)
	if err != nil {
		return err
	}

	token, err := c.srv.GetTokenByChainAndAddress(ctx.Context(), chain, tokenAddress)
	if err != nil {
		return err
	}

	return ctx.JSON(token)
}

// ListTransactions godoc
// @Description Returns transactions. Output is paginated.
// @Tags Wormscan
// @ID list-transactions
// @Param page query integer false "Page number. Starts at 0."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} ListTransactionsResponse
// @Failure 400
// @Failure 500
// @Router /api/v1/transactions/ [get]
func (c *Controller) ListTransactions(ctx *fiber.Ctx) error {

	// Extract query parameters
	pagination, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}
	address, err := middleware.ExtractAddressFromQueryParams(ctx, c.logger)
	if err != nil {
		return err
	}

	// Query transactions from the database
	var queryResult *transactions.ListTransactonsOutput
	if address != nil {
		queryResult, err = c.srv.ListTransactionsByAddress(ctx.Context(), address, pagination)
	} else {
		queryResult, err = c.srv.ListTransactions(ctx.Context(), pagination)
	}
	if err != nil {
		return err
	}

	// Convert query results into the response model
	response := ListTransactionsResponse{
		Transactions: make([]TransactionOverview, 0, len(queryResult.Transactions)),
	}
	for i := range queryResult.Transactions {

		var tx TransactionOverview

		// Base VAA fields
		tx.ID = queryResult.Transactions[i].ID
		tx.Timestamp = queryResult.Transactions[i].Timestamp
		tx.OriginChain = queryResult.Transactions[i].EmitterChain

		// Some VAAs are not generated by the token bridge
		if len(queryResult.Transactions[i].ParsedVaa) == 1 {
			tx.DestinationAddress = queryResult.Transactions[i].ParsedVaa[0].Result.ToAddress
			tx.DestinationChain = queryResult.Transactions[i].ParsedVaa[0].Result.ToChain
		}
		if len(queryResult.Transactions[i].TransferPrices) == 1 {
			tx.Symbol = queryResult.Transactions[i].TransferPrices[0].Symbol
			tx.TokenAmount = queryResult.Transactions[i].TransferPrices[0].TokenAmount
			tx.UsdAmount = queryResult.Transactions[i].TransferPrices[0].UsdAmount
		}

		// For Solana VAAs, the txHash that we get from the gossip network is not the real transacion hash,
		// so we have to overwrite it with the real txHash.
		if queryResult.Transactions[i].EmitterChain == sdk.ChainIDSolana &&
			len(queryResult.Transactions[i].GlobalTransations) == 1 &&
			queryResult.Transactions[i].GlobalTransations[0].OriginTx != nil {

			tx.TxHash = queryResult.Transactions[i].GlobalTransations[0].OriginTx.TxHash
		} else {
			tx.TxHash = queryResult.Transactions[i].TxHash
		}

		// Set the status based on the outcome of the redeem transaction.
		if len(queryResult.Transactions[i].GlobalTransations) == 1 &&
			queryResult.Transactions[i].GlobalTransations[0].DestinationTx != nil &&
			queryResult.Transactions[i].GlobalTransations[0].DestinationTx.Status == domain.DstTxStatusConfirmed {

			tx.Status = TxStatusCompleted
		} else {
			tx.Status = TxStatusOngoing
		}

		response.Transactions = append(response.Transactions, tx)
	}

	return ctx.JSON(response)
}
