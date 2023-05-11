package portalanalytic

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Repository is a portal analytic repository.
type Repository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		portalAnalytic *mongo.Collection
	}
}

// NewPortalAnalytic create a new portal analytic repository.
func NewPortalAnalytic(db *mongo.Database, log *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: log,
	}
}

// GetPortalAnalytic get portal analytic.
func (r *Repository) GetPortalAnalyticByIds(ctx context.Context, ids []string) ([]*PortalAnalyticdDoc, error) {
	return []*PortalAnalyticdDoc{}, nil
}
