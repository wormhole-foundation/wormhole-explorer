package builder

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"go.uber.org/zap"
)

func NewDatabase(ctx context.Context, cfg *config.Configuration, logger *zap.Logger) (*db.DB, error) {
	// Enable database logging
	var options db.Option
	if cfg.DatabaseLogEnabled {
		options = db.WithTracer(logger)
	}

	return db.NewDB(ctx, cfg.DatabaseUrl, options)
}
