package transactions

import (
	"context"

	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService create a new Service.
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.With(zap.String("module", "TransactionService"))}
}

// GetLastTrx get the last transactions.
func (s *Service) GetLastTrx(timeSpan string, sampleRate string) ([]string, error) {
	// TODO invoke repository to get the last transactions.

	return []string{}, nil
}

// GetChainActivity get chain activity.
func (s *Service) GetChainActivity(ctx context.Context, q *ChainActivityQuery) ([]ChainActivityResult, error) {
	return s.repo.FindChainActivity(ctx, q)
}
