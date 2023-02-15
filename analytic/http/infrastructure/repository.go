package infrastructure

import (
	"context"

	"go.uber.org/zap"
)

// Repository definition.
type Repository struct {
	// influx client
	logger *zap.Logger
}

// NewRepository create a new Repository instance.
func NewRepository(logger *zap.Logger) *Repository {
	return &Repository{
		logger: logger.With(zap.String("module", "InfraestructureRepository")),
	}
}

// GetInfluxStatus get influx server status.
func (r *Repository) GetInfluxStatus(ctx context.Context) (*InfluxStatus, error) {
	return &InfluxStatus{Message: "ready for queries and writes", Status: "pass"}, nil
}
