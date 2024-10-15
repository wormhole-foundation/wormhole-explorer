package healthcheck

import (
	"context"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
)

func Postresql(dbClient *db.DB) Check {
	return func(ctx context.Context) error {
		return dbClient.Ping(ctx)
	}
}
