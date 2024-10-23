package stats

import (
	"context"
	"errors"
	"fmt"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/stats"
	"go.uber.org/zap"
)

type Service struct {
	repo               *Repository
	addressRepositorty *stats.AddressRepository
	holderRepository   *stats.HolderRepositoryReadable
	cache              cache.Cache
	expiration         time.Duration
	metrics            metrics.Metrics
	logger             *zap.Logger
}

const (
	topSymbolsByVolumeKey  = "wormscan:top-assets-symbol-by-volume"
	topCorridorsByCountKey = "wormscan:top-corridors-by-count"
	nttSummary             = "wormscan:ntt-summary"
	nttChainActivity       = "wormscan:ntt-ntt-chain-activity"
	nttTransferByTime      = "wormscan:ntt-transfer-by-time"
)

// NewService create a new Service.
func NewService(repo *Repository, statsRepository *stats.AddressRepository,
	holderRepository *stats.HolderRepositoryReadable, cache cache.Cache,
	expiration time.Duration, metrics metrics.Metrics, logger *zap.Logger) *Service {
	return &Service{
		repo:               repo,
		addressRepositorty: statsRepository,
		holderRepository:   holderRepository,
		cache:              cache,
		expiration:         expiration,
		metrics:            metrics,
		logger:             logger.With(zap.String("module", "StatsService"))}
}

func (s *Service) GetSymbolWithAssets(ctx context.Context, ts SymbolWithAssetsTimeSpan) ([]SymbolWithAssetDTO, error) {
	key := topSymbolsByVolumeKey
	key = fmt.Sprintf("%s:%s", key, ts)
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, key, s.metrics,
		func() ([]SymbolWithAssetDTO, error) {
			return s.repo.GetSymbolWithAssets(ctx, ts)
		})
}

func (s *Service) GetTopCorridors(ctx context.Context, ts TopCorridorsTimeSpan) ([]TopCorridorsDTO, error) {
	key := topCorridorsByCountKey
	key = fmt.Sprintf("%s:%s", key, ts)
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, key, s.metrics,
		func() ([]TopCorridorsDTO, error) {
			return s.repo.GetTopCorridores(ctx, ts)
		})
}

func (s *Service) GetNativeTokenTransferSummary(ctx context.Context, symbol string) (*NativeTokenTransferSummary, error) {
	if strings.ToUpper(symbol) != "W" {
		return nil, errors.New("symbol not supported")
	}

	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, nttSummary, s.metrics,
		func() (*NativeTokenTransferSummary, error) {
			return s.repo.GetNativeTokenTransferSummary(ctx, symbol)
		})
}

func (s *Service) GetNativeTokenTransferActivity(ctx context.Context, isNotional bool, symbol string) ([]NativeTokenTransferActivity, error) {
	if symbol != "W" {
		return nil, errors.New("symbol not supported")
	}
	key := fmt.Sprintf("%s:%s:%t", nttChainActivity, symbol, isNotional)
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, key, s.metrics,
		func() ([]NativeTokenTransferActivity, error) {
			return s.repo.GetNativeTokenTransferActivity(ctx, isNotional, symbol)
		})
}

func (s *Service) GetNativeTokenTransferByTime(ctx context.Context, timespan NttTimespan, symbol string, isNotional bool, from, to time.Time) ([]NativeTokenTransferByTime, error) {
	if symbol != "W" {
		return nil, errors.New("symbol not supported")
	}

	timeDuration := to.Sub(from)

	if timespan == HourNttTimespan && timeDuration > 15*24*time.Hour {
		return nil, errors.New("time range is too large for hourly data. Max time range allowed: 15 days")
	}

	if timespan == DayNttTimespan {
		if timeDuration < 24*time.Hour {
			return nil, errors.New("time range is too small for daily data. Min time range allowed: 2 day")
		}

		if timeDuration > 365*24*time.Hour {
			return nil, errors.New("time range is too large for daily data. Max time range allowed: 1 year")
		}
	}

	if timespan == MonthNttTimespan {
		if timeDuration < 30*24*time.Hour {
			return nil, errors.New("time range is too small for monthly data. Min time range allowed: 60 days")
		}

		if timeDuration > 10*365*24*time.Hour {
			return nil, errors.New("time range is too large for monthly data. Max time range allowed: 1 year")
		}
	}

	if timespan == YearNttTimespan {
		if timeDuration < 365*24*time.Hour {
			return nil, errors.New("time range is too small for yearly data. Min time range allowed: 1 year")
		}

		if timeDuration > 10*365*24*time.Hour {
			return nil, errors.New("time range is too large for yearly data. Max time range allowed: 10 year")
		}
	}
	fromStr := from.Format(time.RFC3339)
	toStr := to.Format(time.RFC3339)
	key := fmt.Sprintf("%s:%s:%s:%t:%s:%s", nttTransferByTime, timespan, symbol, isNotional, fromStr, toStr)

	return cacheable.GetOrLoad(ctx, s.logger, s.cache, s.expiration, key, s.metrics,
		func() ([]NativeTokenTransferByTime, error) {
			return s.repo.GetNativeTokenTransferByTime(ctx, timespan, symbol, isNotional, from, to)
		})

}

func (s *Service) GetNativeTokenTransferAddressTop(ctx context.Context, symbol string, isNotional bool) ([]stats.NativeTokenTransferTopAddress, error) {
	if symbol != "W" {
		return nil, errors.New("symbol not supported")
	}

	return s.addressRepositorty.GetNativeTokenTransferTopAddress(ctx, symbol, isNotional)
}

func (s *Service) GetNativeTokenTransferTopHolder(ctx context.Context, symbol string) ([]stats.NativeTokenTransferTopHolder, error) {
	if symbol != "W" {
		return nil, errors.New("symbol not supported")
	}
	return s.holderRepository.GetNativeTokenTransferTopHolder(ctx, symbol)
}

func (s *Service) GetNativeTokenTransferTokensList(ctx context.Context) ([]Token, error) {

	nttTokens, err := s.repo.RetrieveTokenListFromNTTVaas(ctx)
	if err != nil {
		s.logger.Error("failed to retrieve token list from ntt vaas", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("retrieved token list from ntt vaas", zap.Int("count", len(nttTokens)))
	result := make([]Token, 0, len(nttTokens))

	for _, token := range nttTokens {

		tokenAddr, err := domain.NormalizeContractAddress(token.TokenAddress)
		if err != nil {
			s.logger.Error("failed to normalize token address", zap.Error(err), zap.String("token_address", token.TokenAddress), zap.String("token_chain", token.TokenChain))
			continue
		}
		cacheKey := fmt.Sprintf("wormscan:ntt-token:%s:%s", token.TokenChain, tokenAddr)

		tokenData, err := cacheable.GetOrLoad(ctx, s.logger, s.cache, time.Minute*60*24, cacheKey, s.metrics, func() (Token, error) {
			coingeckoToken, err := s.repo.FetchTokenFromCoingecko(ctx, token.TokenChain, tokenAddr)
			if err != nil {
				s.logger.Error("failed to fetch token from coingecko", zap.Error(err), zap.String("token_address", token.TokenAddress), zap.String("token_chain", token.TokenChain))
				return Token{}, err
			}
			return Token{
				Chain:       token.chainID,
				Address:     tokenAddr,
				Symbol:      coingeckoToken.Symbol,
				CoingeckoID: coingeckoToken.Id,
			}, nil
		})

		if err == nil {
			result = append(result, tokenData)
		}
	}

	return result, nil
}
