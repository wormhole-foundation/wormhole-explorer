package stats

import (
	"context"
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
)

// NewService create a new Service.
func NewService(repo *Repository, cache cache.Cache, expiration time.Duration, metrics metrics.Metrics, logger *zap.Logger) *Service {
	return &Service{repo: repo,
		cache:      cache,
		expiration: expiration,
		metrics:    metrics,
		logger:     logger.With(zap.String("module", "StatsService")),
	}
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
