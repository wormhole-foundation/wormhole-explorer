package stats

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	stats2 "github.com/wormhole-foundation/wormhole-explorer/common/stats"
	"go.uber.org/zap"
)

type Service struct {
	repo             *Repository
	statsRepositorty *stats2.AddressRepository
	cache            cache.Cache
	expiration       time.Duration
	metrics          metrics.Metrics
	logger           *zap.Logger
}

const (
	topSymbolsByVolumeKey  = "wormscan:top-assets-symbol-by-volume"
	topCorridorsByCountKey = "wormscan:top-corridors-by-count"
	nttSummary             = "wormscan:ntt-summary"
	nttChainActivity       = "wormscan:ntt-ntt-chain-activity"
	nttTransferByTime      = "wormscan:ntt-transfer-by-time"
)

// NewService create a new Service.
func NewService(repo *Repository, statsRepository *stats2.AddressRepository, cache cache.Cache,
	expiration time.Duration, metrics metrics.Metrics, logger *zap.Logger) *Service {
	return &Service{
		repo:             repo,
		statsRepositorty: statsRepository,
		cache:            cache,
		expiration:       expiration,
		metrics:          metrics,
		logger:           logger.With(zap.String("module", "StatsService"))}
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

func (s *Service) GetNativeTokenTransferAddressTop(ctx context.Context, symbol string, isNotional bool) ([]stats2.NativeTokenTransferTopAddress, error) {
	if symbol != "W" {
		return nil, errors.New("symbol not supported")
	}

	return s.statsRepositorty.GetNativeTokenTransferTopAddress(ctx, symbol, isNotional)
}

func (s *Service) GetTopHolder(ctx context.Context, symbol string) ([]TopHolder, error) {
	//return s.repo.GetTopHolder(ctx, symbol)
	return nil, nil
}
