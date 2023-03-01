package transactions

import "go.uber.org/zap"

type Service struct {
	logger *zap.Logger
}

// NewService create a new Service.
func NewService(logger *zap.Logger) *Service {
	return &Service{logger.With(zap.String("module", "TransactionService"))}
}
