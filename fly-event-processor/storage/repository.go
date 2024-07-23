package storage

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

// PostgresRepository is a repository for postgres.
type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

// NewPostgresRepository creates a new repository.
func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

// GetGovernorConfig gets a governor config by node address.
func (r *PostgresRepository) GetGovernorConfig(
	ctx context.Context,
	nodeAddress string) ([]GovernorConfigChain, error) {

	query := `
	SELECT governor_config_id, chain_id, notional_limit, big_transaction_size, created_at, updated_at
	FROM wormhole.governor_config_chains 
	WHERE governor_config_id = $1`

	var rows []GovernorConfigChain
	err := r.db.Select(ctx, &rows, query, nodeAddress)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// UpdateGovernorConfigChains updates governor config chains.
func (r *PostgresRepository) UpdateGovernorConfigChains(
	ctx context.Context,
	nodeAddress string,
	chains []GovernorConfigChain) error {

	// Start transaction.
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	// Delete existing governor config chains.
	_, err = tx.Exec(ctx, `DELETE FROM wormhole.governor_config_chains WHERE governor_config_id = $1`, nodeAddress)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	// Insert new governor config chains.
	now := time.Now()
	for _, chain := range chains {
		_, err = tx.Exec(ctx, `
		INSERT INTO wormhole.governor_config_chains (governor_config_id, chain_id, notional_limit, big_transaction_size, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
			nodeAddress, chain.ChainID, chain.NotionalLimit, chain.BigTransactionSize, now, now)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	// Commit transaction.
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
