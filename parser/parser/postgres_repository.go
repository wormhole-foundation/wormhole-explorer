package parser

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

// PostgresRepository is a postgres repository.
type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

// NewPostgresRepository creates a new postgres repository.
func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger}
}

// UpsertParsedVaa saves vaa information and parsed result.
func (s *Repository) UpsertParsedVaa2(ctx context.Context, parsedVAA ParsedVaaUpdate) error {
	return nil
}
