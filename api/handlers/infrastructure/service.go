package infrastructure

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService create a new governor.Service.
func NewService(dao *Repository, logger *zap.Logger) *Service {
	return &Service{repo: dao, logger: logger.With(zap.String("module", "Infrastructureervice"))}
}

// CheckMongoServerStatus
func (s *Service) CheckMongoServerStatus(ctx context.Context) (bool, error) {
	mongoStatus, err := s.repo.GetMongoStatus(ctx)
	if err != nil {
		return false, err
	}

	// check mongo server status
	mongoStatusCheck := (mongoStatus.Ok == 1 && mongoStatus.Pid > 0 && mongoStatus.Uptime > 0)
	if !mongoStatusCheck {
		return false, fmt.Errorf("mongo server not ready (Ok = %v, Pid = %v, Uptime = %v)", mongoStatus.Ok, mongoStatus.Pid, mongoStatus.Uptime)
	}

	// check mongo connections
	if mongoStatus.Connections.Available <= 0 {
		return false, fmt.Errorf("mongo server without available connections (availableConection = %v)", mongoStatus.Connections.Available)
	}
	return true, nil
}
