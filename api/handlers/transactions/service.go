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
	repo              *Repository
	cache             cache.Cache
	expiration        time.Duration
	supportedChainIDs map[vaa.ChainID]string
	tokenProvider     *domain.TokenProvider
	metrics           metrics.Metrics
	logger            *zap.Logger
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
func NewService(repo *Repository, cache cache.Cache, expiration time.Duration, tokenProvider *domain.TokenProvider, metrics metrics.Metrics, logger *zap.Logger) *Service {
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
