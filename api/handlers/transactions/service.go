package transactions

import (
	"context"
	errors "errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Service struct {
	repo              repository
	cache             cache.Cache
	expiration        time.Duration
	supportedChainIDs map[vaa.ChainID]string
	tokenProvider     *domain.TokenProvider
	metrics           metrics.Metrics
	logger            *zap.Logger
}

// decouple service from repository
type repository interface {
	GetTopAssets(ctx context.Context, timeSpan *TopStatisticsTimeSpan) ([]AssetDTO, error)
	GetTopChainPairs(ctx context.Context, timeSpan *TopStatisticsTimeSpan) ([]ChainPairDTO, error)
	FindChainActivity(ctx context.Context, q *ChainActivityQuery) ([]ChainActivityResult, error)
	GetScorecards(ctx context.Context) (*Scorecards, error)
	FindGlobalTransactionByID(ctx context.Context, q *GlobalTransactionQuery) (*GlobalTransactionDoc, error)
	FindTransactions(ctx context.Context, input *FindTransactionsInput) ([]TransactionDto, error)
	ListTransactionsByAddress(ctx context.Context, address string, pagination *pagination.Pagination) ([]TransactionDto, error)
	FindChainActivityTops(ctx *fasthttp.RequestCtx, q ChainActivityTopsQuery) ([]ChainActivityTopResult, error)
	FindApplicationActivity(ctx *fasthttp.RequestCtx, q ApplicationActivityQuery) ([]ApplicationActivityTotalsResult, []ApplicationActivityResult, error)
	FindTokensVolume(ctx context.Context) ([]TokenVolume, error)
	FindTokenSymbolActivity(ctx context.Context, payload TokenSymbolActivityQuery) ([]TokenSymbolActivityResult, error)
	GetTransactionCount(ctx context.Context, q *TransactionCountQuery) ([]TransactionCountResult, error)
}

const (
	lastTxsKey                     = "wormscan:last-txs"
	scorecardsKey                  = "wormscan:scorecards"
	topAssetsByVolumeKey           = "wormscan:top-assets-by-volume"
	topChainPairsByNumTransfersKey = "wormscan:top-chain-pairs-by-num-transfers"
	chainActivityKey               = "wormscan:chain-activity"
	chainActivityTopsKey           = "wormscan:chain-activity-tops"
)

// NewService create a new Service.
func NewService(repo repository, cache cache.Cache, expiration time.Duration, tokenProvider *domain.TokenProvider, metrics metrics.Metrics, logger *zap.Logger) *Service {
	supportedChainIDs := domain.GetSupportedChainIDs()
	return &Service{repo: repo, supportedChainIDs: supportedChainIDs,
		cache: cache, expiration: expiration, tokenProvider: tokenProvider, metrics: metrics,
		logger: logger.With(zap.String("module", "TransactionService"))}
}

// GetTransactionCount get the last transactions.
func (s *Service) GetTransactionCount(ctx context.Context, q *TransactionCountQuery) ([]TransactionCountResult, error) {
	key := fmt.Sprintf("%s:%s:%s:%v", lastTxsKey, q.TimeSpan, q.SampleRate, q.CumulativeSum)
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, key, s.metrics,
		func() ([]TransactionCountResult, error) {
			return s.repo.GetTransactionCount(ctx, q)
		})
}

func (s *Service) GetScorecards(ctx context.Context) (*Scorecards, error) {
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, scorecardsKey, s.metrics,
		func() (*Scorecards, error) {
			return s.repo.GetScorecards(ctx)
		})
}

func (s *Service) GetTopAssets(ctx context.Context, timeSpan *TopStatisticsTimeSpan) ([]AssetDTO, error) {
	key := topAssetsByVolumeKey
	if timeSpan != nil {
		key = fmt.Sprintf("%s:%s", key, *timeSpan)
	}
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, key, s.metrics,
		func() ([]AssetDTO, error) {
			return s.repo.GetTopAssets(ctx, timeSpan)
		})
}

func (s *Service) GetTopChainPairs(ctx context.Context, timeSpan *TopStatisticsTimeSpan) ([]ChainPairDTO, error) {
	key := topChainPairsByNumTransfersKey
	if timeSpan != nil {
		key = fmt.Sprintf("%s:%s", key, *timeSpan)
	}
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, key, s.metrics,
		func() ([]ChainPairDTO, error) {
			return s.repo.GetTopChainPairs(ctx, timeSpan)
		})
}

// GetChainActivity get chain activity.
func (s *Service) GetChainActivity(ctx context.Context, q *ChainActivityQuery) ([]ChainActivityResult, error) {
	key := fmt.Sprintf("%s:%s:%v:%s", chainActivityKey, q.TimeSpan, q.IsNotional, strings.Join(q.GetAppIDs(), ","))
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, key, s.metrics,
		func() ([]ChainActivityResult, error) {
			return s.repo.FindChainActivity(ctx, q)
		})
}

func (s *Service) GetTokensByVolume(ctx context.Context, limit int) ([]TokenVolume, error) {
	key := "wormscan:tokens-by-volume"
	value, err := cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, key, s.metrics,
		func() ([]TokenVolume, error) {
			return s.repo.FindTokensVolume(ctx)
		})
	if err == nil && limit < len(value) {
		value = value[:limit]
	}
	return value, err
}

// FindGlobalTransactionByID find a global transaction by id.
func (s *Service) FindGlobalTransactionByID(ctx context.Context, chainID vaa.ChainID, emitter *types.Address, seq string) (*GlobalTransactionDoc, error) {

	key := fmt.Sprintf("%d/%s/%s", chainID, emitter.Hex(), seq)
	q := GlobalTransactionQuery{id: key}

	return s.repo.FindGlobalTransactionByID(ctx, &q)
}

// GetTokenByChainAndAddress get token by chain and address.
func (s *Service) GetTokenByChainAndAddress(ctx context.Context, chainID vaa.ChainID, tokenAddress *types.Address) (*Token, error) {
	// check if chainID is valid
	if _, ok := s.supportedChainIDs[chainID]; !ok {
		return nil, errs.ErrNotFound
	}

	//get token by contractID (chainID + tokenAddress)
	tokenMetadata, ok := s.tokenProvider.GetTokenByAddress(chainID, tokenAddress.Hex())
	if !ok {
		return nil, errs.ErrNotFound
	}

	return &Token{
		Symbol:      tokenMetadata.Symbol,
		CoingeckoID: tokenMetadata.CoingeckoID,
		Decimals:    tokenMetadata.Decimals,
	}, nil
}

func (s *Service) ListTransactions(
	ctx context.Context,
	pagination *pagination.Pagination,
) ([]TransactionDto, error) {

	input := FindTransactionsInput{
		sort:       true,
		pagination: pagination,
	}
	return s.repo.FindTransactions(ctx, &input)
}

func (s *Service) ListTransactionsByAddress(
	ctx context.Context,
	address string,
	pagination *pagination.Pagination,
) ([]TransactionDto, error) {

	return s.repo.ListTransactionsByAddress(ctx, address, pagination)
}

func (s *Service) GetTransactionByID(
	ctx context.Context,
	chain vaa.ChainID,
	emitter *types.Address,
	seq string,
) (*TransactionDto, error) {

	// Execute the database query
	input := FindTransactionsInput{
		id: fmt.Sprintf("%d/%s/%s", chain, emitter.Hex(), seq),
	}
	output, err := s.repo.FindTransactions(ctx, &input)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, errs.ErrNotFound
	}

	// Return matching document
	return &output[0], nil
}

func (s *Service) GetTokenProvider() *domain.TokenProvider {
	return s.tokenProvider
}

func (s *Service) GetChainActivityTops(ctx *fasthttp.RequestCtx, q ChainActivityTopsQuery) (ChainActivityTopResults, error) {

	timeDuration := q.To.Sub(q.From)

	if q.Timespan == Hour && timeDuration > 15*24*time.Hour {
		return nil, errors.New("time range is too large for hourly data. Max time range allowed: 15 days")
	}

	if q.Timespan == Day {
		if timeDuration < 24*time.Hour {
			return nil, errors.New("time range is too small for daily data. Min time range allowed: 2 day")
		}

		if timeDuration > 365*24*time.Hour {
			return nil, errors.New("time range is too large for daily data. Max time range allowed: 1 year")
		}
	}

	if q.Timespan == Month {
		if timeDuration < 30*24*time.Hour {
			return nil, errors.New("time range is too small for monthly data. Min time range allowed: 60 days")
		}

		if timeDuration > 10*365*24*time.Hour {
			return nil, errors.New("time range is too large for monthly data. Max time range allowed: 1 year")
		}
	}

	if q.Timespan == Year {
		if timeDuration < 365*24*time.Hour {
			return nil, errors.New("time range is too small for yearly data. Min time range allowed: 1 year")
		}

		if timeDuration > 10*365*24*time.Hour {
			return nil, errors.New("time range is too large for yearly data. Max time range allowed: 10 year")
		}
	}

	return s.repo.FindChainActivityTops(ctx, q)
}

func (s *Service) GetApplicationActivity(ctx *fasthttp.RequestCtx, q ApplicationActivityQuery) ([]AppActivityTotalData, error) {
	totals, appActivities, err := s.repo.FindApplicationActivity(ctx, q)
	if err != nil {
		return nil, err
	}

	result := make([]AppActivityTotalData, 0, len(totals)+len(appActivities))

	for _, total := range totals {
		total.AppID, _ = strings.CutPrefix(total.AppID, "TOTAL_")
		foundTotalObj := false
		for i := 0; i < len(result); i++ {
			if result[i].AppID == total.AppID {
				foundTotalObj = true
				result[i].TimeRangeData = append(result[i].TimeRangeData, TimeRangeData[AggregationsAppActivity]{
					TotalMessages:         total.Txs,
					TotalValueTransferred: total.Volume,
					From:                  total.From,
					To:                    total.To,
					Aggregations:          make([]AggregationsAppActivity, 0, len(appActivities)),
				})
				break
			}
		}

		if !foundTotalObj {
			data := AppActivityTotalData{
				AppID: total.AppID,
				TimeRangeData: []TimeRangeData[AggregationsAppActivity]{
					{
						TotalMessages:         total.Txs,
						TotalValueTransferred: total.Volume,
						From:                  total.From,
						To:                    total.To,
						Aggregations:          make([]AggregationsAppActivity, 0, len(appActivities)),
					},
				},
			}
			result = append(result, data)
		}
	}

	for _, ac := range appActivities {
		result = addAppActivity(ac.AppID1, ac.AppID2, ac.From, ac.To, ac.Volume, ac.Txs, result)
		if ac.AppID2 != "none" {
			result = addAppActivity(ac.AppID2, ac.AppID1, ac.From, ac.To, ac.Volume, ac.Txs, result)
		}
	}

	if q.AppId != "" {
		for _, rs := range result {
			if rs.AppID == q.AppId {
				return []AppActivityTotalData{rs}, nil
			}
		}
	}

	// remove UNKNOWN from response
	for i := 0; i < len(result); i++ {
		if result[i].AppID == "UNKNOWN" {
			result = append(result[:i], result[i+1:]...)
			break
		}
	}

	return result, nil
}

func (s *Service) GetTokenSymbolActivity(ctx context.Context, payload TokenSymbolActivityQuery) (TokenSymbolActivityResponse, error) {
	rows, err := s.repo.FindTokenSymbolActivity(ctx, payload)
	if err != nil {
		return TokenSymbolActivityResponse{}, err
	}

	// Map to accumulate tokens data
	tokens := make(map[string]TokenSymbolActivity)

	for _, row := range rows {

		token, exists := tokens[row.Symbol]
		if !exists {
			token = TokenSymbolActivity{
				TokenSymbol:   row.Symbol,
				TimeRangeData: []*TimeRangeData[*TokenSymbolPerChainPairData]{},
			}
		}

		// Update the total messages and value transferred
		token.TotalMessages += row.Txs
		token.TotalValueTransferred += row.Volume

		// Find the correct time range or create a new one
		var timeRange *TimeRangeData[*TokenSymbolPerChainPairData]
		for i := range token.TimeRangeData {
			if token.TimeRangeData[i].From == row.From {
				timeRange = token.TimeRangeData[i]
				break
			}
		}

		if timeRange == nil {
			timeRange = &TimeRangeData[*TokenSymbolPerChainPairData]{
				From:         row.From,
				To:           row.To,
				Aggregations: []*TokenSymbolPerChainPairData{},
			}
			token.TimeRangeData = append(token.TimeRangeData, timeRange)
		}

		// Update time range data
		timeRange.TotalMessages += row.Txs
		timeRange.TotalValueTransferred += row.Volume

		// Create aggregation
		agg := &TokenSymbolPerChainPairData{
			SourceChain:           row.EmitterChain,
			TargetChain:           row.DestinationChain,
			TotalMessages:         row.Txs,
			TotalValueTransferred: row.Volume,
		}
		timeRange.Aggregations = append(timeRange.Aggregations, agg)

		tokens[row.Symbol] = token
	}

	resp := TokenSymbolActivityResponse{}
	for _, token := range tokens {
		resp.Tokens = append(resp.Tokens, token)
	}

	return resp, nil
}

func addAppActivity(appID1, appID2 string, from, to time.Time, volume float64, txs uint64, result []AppActivityTotalData) []AppActivityTotalData {

	appID := appID1
	if appID2 != "none" {
		appID = appID2
	}

	for i := 0; i < len(result); i++ {
		res := &result[i]
		if res.AppID == appID1 {
			for j := 0; j < len(res.TimeRangeData); j++ {
				rtrd := &res.TimeRangeData[j]
				if rtrd.From == from && rtrd.To == to {
					rtrd.Aggregations = append(rtrd.Aggregations, AggregationsAppActivity{
						AppID:                 appID,
						TotalMessages:         txs,
						TotalValueTransferred: volume,
					})
					return result
				}
			}
			res.TimeRangeData = append(res.TimeRangeData, TimeRangeData[AggregationsAppActivity]{
				TotalMessages:         txs,
				TotalValueTransferred: volume,
				From:                  from,
				To:                    to,
				Aggregations: []AggregationsAppActivity{
					{
						AppID:                 appID,
						TotalMessages:         txs,
						TotalValueTransferred: volume,
					},
				},
			})
			return result
		}
	}

	data := AppActivityTotalData{
		AppID: appID1,
		TimeRangeData: []TimeRangeData[AggregationsAppActivity]{
			{
				TotalMessages:         txs,
				TotalValueTransferred: volume,
				From:                  from,
				To:                    to,
				Aggregations: []AggregationsAppActivity{
					{
						AppID:                 appID,
						TotalMessages:         txs,
						TotalValueTransferred: volume,
					},
				},
			},
		},
	}
	return append(result, data)
}

type AppActivityTotalData struct {
	AppID         string                                   `json:"app_id"`
	TimeRangeData []TimeRangeData[AggregationsAppActivity] `json:"time_range_data"`
}

type TimeRangeData[T any] struct {
	From                  time.Time `json:"from"`
	To                    time.Time `json:"to"`
	TotalMessages         uint64    `json:"total_messages"`
	TotalValueTransferred float64   `json:"total_value_transferred"`
	Aggregations          []T       `json:"aggregations,omitempty"`
}

type AggregationsAppActivity struct {
	AppID                 string  `json:"app_id"`
	TotalMessages         uint64  `json:"total_messages"`
	TotalValueTransferred float64 `json:"total_value_transferred"`
}

type TokenSymbolActivity struct {
	TokenSymbol           string                                         `json:"token_symbol"`
	TotalMessages         uint64                                         `json:"total_messages"`
	TotalValueTransferred float64                                        `json:"total_value_transferred"`
	TimeRangeData         []*TimeRangeData[*TokenSymbolPerChainPairData] `json:"time_range_data"`
}

type TokenSymbolPerChainPairData struct {
	TotalMessages         uint64      `json:"total_messages"`
	TotalValueTransferred float64     `json:"total_value_transferred"`
	SourceChain           vaa.ChainID `json:"source_chain"`
	TargetChain           vaa.ChainID `json:"target_chain"`
}

type TokenSymbolActivityResponse struct {
	Tokens []TokenSymbolActivity `json:"tokens"`
}
