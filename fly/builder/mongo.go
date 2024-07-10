package builder

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"go.uber.org/zap"
)

func NewMongoDatabase(ctx context.Context, cfg *config.Configuration, logger *zap.Logger) (*dbutil.Session, error) {
	return dbutil.Connect(ctx, logger, cfg.MongoUri, cfg.MongoDatabase, cfg.MongoEnableQueryLog)
}
