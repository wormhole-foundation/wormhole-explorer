package builder

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"go.uber.org/zap"
)

func NewDatabase(ctx context.Context, cfg *config.Configuration, logger *zap.Logger) (*db.DB, error) {
	return db.NewDB(ctx, cfg.DatabaseUrl, db.WithTracer(logger))
}
