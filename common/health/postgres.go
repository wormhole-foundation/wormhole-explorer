package health

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
)

// Postgres checks the connection to the Postgres database.
func Postgres(db *db.DB) Check {
	return func(ctx context.Context) error {
		return db.Ping(ctx)
	}
}
