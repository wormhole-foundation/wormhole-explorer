package repository

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

// Repository is a repository.
type Repository struct {
	db     *db.DB
	logger *zap.Logger
}

// NewRepository creates a new repository.
func NewRepository(db *db.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// GuardianSet is a document for GuardianSet.
func (r *Repository) FindAll(ctx context.Context) ([]*GuardianSet, error) {
	query := `
	SELECT
    	gs.id AS guardian_set_id,
    	gs.expiration_time,
    	gs.created_at AS guardian_set_created_at,
    	gs.updated_at AS guardian_set_updated_at,
    	gsa.index AS address_index,
    	gsa.address,
    	gsa.created_at AS address_created_at,
    	gsa.updated_at AS address_updated_at
	FROM
    	wormhole.wh_guardian_sets gs
	JOIN
    	wormhole.wh_guardian_set_addresses gsa ON gs.id = gsa.guardian_set_id;
	`

	guardianSets := []*GuardianSet{}
	err := r.db.Select(ctx, &guardianSets, query)
	if err != nil {
		r.logger.Error("failed to select guardian sets", zap.Error(err))
		return nil, err
	}
	return guardianSets, nil
}

// Upsert upserts a guardian set document.
func (r *Repository) Upsert(ctx context.Context, gs *GuardianSet) error {
	now := time.Now()
	// start a transaction
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	// query to upsert the guardian set
	query := `
	INSERT INTO wormhole.wh_guardian_sets (id, expiration_time, created_at, updated_at)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE
	SET expiration_time = $2, updated_at = $4;
	`

	// execute the query to upsert the guardian set
	_, err = tx.Exec(ctx, query, gs.GuardianSetIndex, gs.ExpirationTime, now, now)
	if err != nil {
		tx.Rollback(ctx)
		r.logger.Error("failed to upsert guardian set", zap.Error(err))
		return err
	}

	if len(gs.Keys) == 0 {
		return nil
	}

	// build query to upsert the guardian set addresses
	query = `
	INSERT INTO wormhole.wh_guardian_set_addresses (guardian_set_id, index, address, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5) 
	ON CONFLICT (guardian_set_id, index) DO NOTHING;
	`

	// prepare the values for the query
	for _, g := range gs.Keys {
		// execute the query to upsert the guardian set addresses
		_, err = tx.Exec(ctx, query, gs.GuardianSetIndex, g.Index, g.Address, now, now)
		if err != nil {
			tx.Rollback(ctx)
			r.logger.Error("failed to upsert guardian set addresses",
				zap.Error(err))
			return err
		}
	}
	// commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		r.logger.Error("failed to commit transaction",
			zap.Error(err))
		return err
	}

	return nil
}
