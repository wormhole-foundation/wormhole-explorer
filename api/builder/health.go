package builder

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
)

// NewHealthChecks creates a list of health checks.
func NewHealthChecks(
	ctx context.Context,
	db *db.DB,
	mongoDB *dbutil.Session,
) ([]health.Check, error) {

	var checks []health.Check

	if mongoDB != nil {
		checks = append(checks, health.Mongo(mongoDB.Database))
	}

	if db != nil {
		checks = append(checks, health.Postgres(db))
	}

	return checks, nil

}
