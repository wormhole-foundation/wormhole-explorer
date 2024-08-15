package recordcap

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

type Repository struct {
	db     *db.DB
	logger *zap.Logger
}

func NewRepository(db *db.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) DeletePyth(ctx context.Context, maxTime time.Time) error {
	query := "DELETE FROM wormholescan.wh_attestation_vaas_pythnet WHERE timestamp < $1"
	_, err := r.db.Exec(ctx, query, maxTime)
	if err != nil {
		r.logger.Error("error delete record from table wh_attestation_vaas_pythnet", zap.Error(err))
		return err
	}
	return nil
}
