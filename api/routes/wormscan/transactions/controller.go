package transactions

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
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
// @Tags wormholescan
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
// @Tags wormholescan
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
		TotalMessages: scorecards.TotalMessages,
		Messages24h:   scorecards.Messages24h,
		TotalTxCount:  scorecards.TotalTxCount,
		TotalVolume:   scorecards.TotalTxVolume,
		Tvl:           scorecards.Tvl,
		TxCount24h:    scorecards.TxCount24h,
		Volume24h:     scorecards.Volume24h,
	}

	return ctx.JSON(response)
}

// GetTopChainPairs godoc
// @Description Returns a list of the emitter_chain and destination_chain pair ordered by transfer count.
// @Tags wormholescan
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
// @Tags wormholescan
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
		tokenMeta, ok := c.srv.GetTokenProvider().GetTokenByAddress(assetDTOs[i].TokenChain, assetDTOs[i].TokenAddress)
		if ok {
			asset.Symbol = tokenMeta.Symbol.String()
		}

		response.Assets = append(response.Assets, asset)
	}

	return ctx.JSON(response)
}

// GetChainActivityTops godoc
// @Description Search for a specific period of time the number of transactions and the volume.
// @Tags wormholescan
// @ID x-chain-activity-tops
// @Method Get
// @Param timespan query string true "Time span, supported values: 1d, 1mo and 1y"
// @Param from query string true "From date, supported format 2006-01-02T15:04:05Z07:00"
// @Param to query string true "To date, supported format 2006-01-02T15:04:05Z07:00"
// @Param appId query string false "Search by appId"
// @Param sourceChain query string false "Search by sourceChain"
// @Param targetChain query string false "Search by targetChain"
// @Success 200 {object} transactions.ChainActivityTopResults
// @Failure 400
// @Failure 500
// @Router /api/v1/x-chain-activity/tops [get]
func (c *Controller) GetChainActivityTops(ctx *fiber.Ctx) error {

	sourceChain, err := middleware.ExtractSourceChain(ctx, c.logger)
	if err != nil {
		return err
	}
	targetChain, err := middleware.ExtractTargetChain(ctx, c.logger)
	if err != nil {
		return err
	}
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

	payload := transactions.ChainActivityTopsQuery{
		SourceChain: sourceChain,
		TargetChain: targetChain,
		From:        *from,
		To:          *to,
		AppId:       middleware.ExtractAppId(ctx, c.logger),
		Timespan:    transactions.Timespan(ctx.Query("timespan")),
	}

	if !payload.Timespan.IsValid() {
		return response.NewInvalidParamError(ctx, "invalid timespan", nil)
	}

	nowUTC := time.Now().UTC()
	if nowUTC.Before(payload.To.UTC()) {
		payload.To = nowUTC
	}

	if payload.To.Sub(payload.From) <= 0 {
		return response.NewInvalidParamError(ctx, "invalid time range", nil)
	}

	// Get the chain activity.
	activity, err := c.srv.GetChainActivityTops(ctx.Context(), payload)
	if err != nil {
		c.logger.Error("Error getting chain activity", zap.Error(err))
		return err
	}

	return ctx.JSON(activity)
}

// GetChainActivity godoc
// @Description Returns a list of chain pairs by origin chain and destination chain.
// @Description The list could be rendered by notional or transaction count.
// @Description The volume is calculated using the notional price of the symbol at the day the VAA was emitted.
// @Tags wormholescan
// @ID x-chain-activity
// @Param timeSpan query string false "Time span, supported values: 7d, 30d, 90d, 1y and all-time (default is 7d)."
// @Param by query string false "Renders the results using notional or tx count (default is notional)."
// @Param apps query string false "List of apps separated by comma (default is all apps)."
// @Success 200 {object} transactions.ChainActivity
// @Failure 400
// @Failure 500
// @Router /api/v1/x-chain-activity [get]
func (c *Controller) GetChainActivity(ctx *fiber.Ctx) error {

	apps, err := middleware.ExtractApps(ctx)
	if err != nil {
		return err
	}
	isNotional, err := middleware.ExtractIsNotional(ctx)
	if err != nil {
		return err
	}
	timeSpan, err := middleware.ExtractChainActivityTimeSpan(ctx)
	if err != nil {
		return err
	}

	q := &transactions.ChainActivityQuery{
		TimeSpan:   timeSpan,
		IsNotional: isNotional,
		AppIDs:     apps,
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
		for i, destination := range item.Destinations {
			if item.Volume.GreaterThan(decimal.Zero) {
				percentage, _ := destination.Volume.Div(item.Volume).Mul(oneHundred).Float64()
				item.Destinations[i].Percentage = percentage
			}
			if isNotional {
				item.Destinations[i].Volume = convertToDecimal(destination.Volume)
			}
		}
		if isNotional {
			item.Volume = convertToDecimal(item.Volume)
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
// @Tags wormholescan
// @ID find-global-transaction-by-id
// @Param chain_id path integer true "id of the blockchain"
// @Param emitter path string true "address of the emitter"
// @Param seq path integer true "sequence of the VAA"
// @Success 200 {object} Tx
// @Failure 400
// @Failure 500
// @Router /api/v1/global-tx/:chain_id/:emitter/:seq [get]
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
// @Tags wormholescan
// @ID get-token-by-chain-and-address
// @Param chain_id path integer true "id of the blockchain"
// @Param token_address path string true "token address"
// @Success 200 {object} transactions.Token
// @Failure 400
// @Failure 404
// @Router /api/v1/token/:chain_id/:token_address [get]
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
// @Tags wormholescan
// @ID list-transactions
// @Param page query integer false "Page number. Starts at 0."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Param address query string false "Filter transactions by Address."
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
	address := middleware.ExtractAddressFromQueryParams(ctx, c.logger)

	// Check pagination max limit
	if pagination.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	// Query transactions from the database
	var dtos []transactions.TransactionDto
	if address != "" {
		dtos, err = c.srv.ListTransactionsByAddress(ctx.Context(), address, pagination)
	} else {
		dtos, err = c.srv.ListTransactions(ctx.Context(), pagination)
	}
	if err != nil {
		return err
	}

	// Populate the response struct and return
	response := c.makeTransactionsResponse(dtos)
	return ctx.JSON(response)
}

func (c *Controller) makeTransactionsResponse(dtos []transactions.TransactionDto) ListTransactionsResponse {

	response := ListTransactionsResponse{
		Transactions: make([]*TransactionDetail, 0, len(dtos)),
	}

	for i := range dtos {
		tx := c.makeTransactionDetail(&dtos[i])
		response.Transactions = append(response.Transactions, tx)
	}

	return response
}

func (c *Controller) makeTransactionDetail(input *transactions.TransactionDto) *TransactionDetail {

	tx := TransactionDetail{
		ID:                     input.ID,
		EmitterChain:           input.EmitterChain,
		EmitterAddress:         input.EmitterAddr,
		Timestamp:              input.Timestamp,
		Symbol:                 input.Symbol,
		TokenAmount:            input.TokenAmount,
		UsdAmount:              input.UsdAmount,
		Payload:                input.Payload,
		StandardizedProperties: input.StandardizedProperties,
	}

	// Translate the emitter address into the emitter chain's native format
	var err error
	tx.EmitterNativeAddress, err = domain.TranslateEmitterAddress(tx.EmitterChain, tx.EmitterAddress)
	if err != nil {
		c.logger.Warn("failed to translate emitter address",
			zap.Stringer("chain", tx.EmitterChain),
			zap.String("address", tx.EmitterAddress),
			zap.Error(err),
		)
	}

	// Set the transaction hash
	isSolanaOrAptos := input.EmitterChain == sdk.ChainIDSolana || input.EmitterChain == sdk.ChainIDAptos
	if isSolanaOrAptos {
		// For Solana and Aptos VAAs, the txHash that we get from the gossip network is
		// not the real transacion hash. We have to overwrite it with the real one.
		if len(input.GlobalTransations) == 1 && input.GlobalTransations[0].OriginTx != nil {
			tx.TxHash = input.GlobalTransations[0].OriginTx.TxHash
		}
	} else {
		tx.TxHash = input.TxHash
	}

	// Set the global transaction, if available
	if len(input.GlobalTransations) == 1 {
		tx.GlobalTx = &input.GlobalTransations[0]
	}

	return &tx
}

// GetTransactionByID godoc
// @Description Find VAA metadata by ID.
// @Tags wormholescan
// @ID get-transaction-by-id
// @Param chain_id path integer true "id of the blockchain"
// @Param emitter path string true "address of the emitter"
// @Param seq path integer true "sequence of the VAA"
// @Success 200 {object} TransactionDetail
// @Failure 400
// @Failure 500
// @Router /api/v1/transactions/:chain_id/:emitter/:seq [get]
func (c *Controller) GetTransactionByID(ctx *fiber.Ctx) error {

	// Extract query params
	chainID, emitter, seq, err := middleware.ExtractVAAParams(ctx, c.logger)
	if err != nil {
		return err
	}

	// Look up the VAA by ID
	dto, err := c.srv.GetTransactionByID(
		ctx.Context(),
		chainID,
		emitter,
		strconv.FormatUint(seq, 10),
	)
	if err != nil {
		return err
	}
	if dto == nil {
		return errors.ErrNotFound
	}

	tx := c.makeTransactionDetail(dto)
	return ctx.JSON(tx)
}
