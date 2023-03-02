package transactions

import (
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
}

// NewService create a new Service.
func NewService(logger *zap.Logger) *Service {
	return &Service{logger.With(zap.String("module", "TransactionService"))}
}

// GetLastTrx get the last transactions.
func (s *Service) GetLastTrx(timeSpan string, sampleRate string) ([]string, error) {
	// TODO invoke repository to get the last transactions.

	return []string{}, nil
}
