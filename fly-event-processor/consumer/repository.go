package consumer

import (
	commonRepo "github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Repository exposes operations over the `globalTransactions` collection.
type Repository struct {
	logger        *zap.Logger
	vaas          *mongo.Collection
	duplicateVaas *mongo.Collection
}

// New creates a new repository.
func NewRepository(logger *zap.Logger, db *mongo.Database) *Repository {
	r := Repository{
		logger:        logger,
		vaas:          db.Collection(commonRepo.Vaas),
		duplicateVaas: db.Collection(commonRepo.DuplicateVaas),
	}
	return &r
}
