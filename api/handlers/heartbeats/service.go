package heartbeats

import (
	"context"

	"go.uber.org/zap"
)

// Service definition.
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService create a new Service.
func NewService(dao *Repository, logger *zap.Logger) *Service {
	return &Service{repo: dao, logger: logger.With(zap.String("module", "HearbeatsService"))}
}

// GetHeartbeatsByIds get heartbeats by IDs.
func (s *Service) GetHeartbeatsByIds(ctx context.Context, heartbeatsIDs []string) ([]*HeartbeatDoc, error) {
	return s.repo.FindByIDs(ctx, heartbeatsIDs)
}
