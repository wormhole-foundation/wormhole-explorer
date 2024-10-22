package builder

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"go.uber.org/zap"
)

func NewPostgresDatabase(ctx context.Context, cfg *config.Configuration, logger *zap.Logger) (*db.DB, error) {
	// Enable database logging
	var options db.Option
	if cfg.DatabaseLogEnabled {
		options = db.WithTracer(logger)
	}

	return db.NewDB(ctx, cfg.DatabaseUrl, options)
}

func NewMongoDatabase(ctx context.Context, cfg *config.Configuration, logger *zap.Logger) (*dbutil.Session, error) {
	return dbutil.Connect(ctx, logger, cfg.MongoUri, cfg.MongoDatabase, cfg.MongoEnableQueryLog)
}
