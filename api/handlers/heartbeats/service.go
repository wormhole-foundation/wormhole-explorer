package heartbeats

import (
	"context"

	"go.uber.org/zap"
)

// Service definition.
type Service struct {
	mongoRepo    *MongoHearbeatRepository
	postgresRepo *PostresqlRepository
	logger       *zap.Logger
}

// NewService create a new Service.
func NewService(dao *MongoHearbeatRepository, postresqlRepo *PostresqlRepository, logger *zap.Logger) *Service {
	return &Service{
		mongoRepo:    dao,
		postgresRepo: postresqlRepo,
		logger:       logger.With(zap.String("module", "HearbeatsService"))}
}

// GetHeartbeatsByIds get heartbeats by IDs.
func (s *Service) GetHeartbeatsByIds(ctx context.Context, usePostgres bool, heartbeatsIDs []string) ([]*HeartbeatDoc, error) {
	if usePostgres {
		return s.postgresRepo.FindByIDs(ctx, heartbeatsIDs)
	}
	return s.mongoRepo.FindByIDs(ctx, heartbeatsIDs)
}
