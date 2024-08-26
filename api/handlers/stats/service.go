package stats

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"
)

type Service struct {
	repo       *Repository
	cache      cache.Cache
	expiration time.Duration
	metrics    metrics.Metrics
	logger     *zap.Logger
}

const (
	topSymbolsByVolumeKey  = "wormscan:top-assets-symbol-by-volume"
	topCorridorsByCountKey = "wormscan:top-corridors-by-count"
	nttSummary             = "wormscan:ntt-summary"
	nttChainActivity       = "wormscan:ntt-ntt-chain-activity"
)

// NewService create a new Service.
func NewService(repo *Repository, cache cache.Cache, expiration time.Duration, metrics metrics.Metrics, logger *zap.Logger) *Service {
	return &Service{repo: repo, cache: cache, expiration: expiration, metrics: metrics, logger: logger.With(zap.String("module", "StatsService"))}
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
	if symbol != "W" {
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

func (s *Service) GetNativeTokenTransferByTime(ctx context.Context, symbol string, isNotional bool, from, to time.Time) (*NativeTokenTransferByTime, error) {
	return nil, nil
}

func (s *Service) GetNativeTokenTransferTop(ctx context.Context, symbol string, isNotional bool, from, to time.Time) (*NativeTokenTransferTop, error) {
	return nil, nil
}
