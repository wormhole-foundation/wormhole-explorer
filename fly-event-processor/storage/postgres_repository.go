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

func (r *PostgresRepository) FindNodeGovernorVaaByNodeAddress(ctx context.Context, nodeAddress string) ([]NodeGovernorVaa, error) {
	query := `SELECT * FROM wormhole.node_governor_vaas WHERE node_address = $1`
	var rows []NodeGovernorVaa
	err := r.db.Select(ctx, &rows, query, nodeAddress)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) FindNodeGovernorVaaByVaaID(ctx context.Context, vaaID string) ([]NodeGovernorVaa, error) {
	query := `SELECT * FROM wormhole.node_governor_vaas WHERE vaa_id = $1`
	var rows []NodeGovernorVaa
	err := r.db.Select(ctx, &rows, query, vaaID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) FindNodeGovernorVaaByVaaIDs(ctx context.Context, vaaID []string) ([]NodeGovernorVaa, error) {
	query := `SELECT * FROM wormhole.node_governor_vaas WHERE vaa_id = ANY($1)`
	var rows []NodeGovernorVaa
	err := r.db.Select(ctx, &rows, query, vaaID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) FindGovernorVaaByVaaIDs(ctx context.Context, vaaID []string) ([]GovernorVaa, error) {
	query := `SELECT * FROM wormhole.governor_vaas WHERE _id = ANY($1)`
	var rows []GovernorVaa
	err := r.db.Select(ctx, &rows, query, vaaID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) UpdateGovernor(
	ctx context.Context,
	nodeGovernorVaaDocToInsert []NodeGovernorVaa,
	nodeGovernorVaaDocToDelete []string,
	governorVaasToInsert []GovernorVaa,
	governorVaaIdsToDelete []string) error {

	// Start transaction.
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	// 2. insert node governor vaas.
	for _, nodeGovernorVaa := range nodeGovernorVaaDocToInsert {
		_, err = tx.Exec(ctx, `
		INSERT INTO wormhole.node_governor_vaas (node_address, node_name, vaa_id)
		VALUES ($1, $2, $3)`,
			nodeGovernorVaa.NodeAddress, nodeGovernorVaa.NodeName, nodeGovernorVaa.VaaID)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	// 3. delete node governor vaas.
	for _, vaaID := range nodeGovernorVaaDocToDelete {
		_, err = tx.Exec(ctx, `DELETE FROM wormhole.node_governor_vaas WHERE vaa_id = $1`, vaaID)
		if err != nil {
			_ = tx.Rollback
			return err
		}
	}

	// 4. insert governor vaas.
	for _, governorVaa := range governorVaasToInsert {
		_, err = tx.Exec(ctx, `
		INSERT INTO wormhole.governor_vaas (_id, chain_id, emitter_address, sequence, tx_hash, release_time, amount)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			governorVaa.ID, governorVaa.ChainID, governorVaa.EmitterAddress, governorVaa.Sequence, governorVaa.TxHash, governorVaa.ReleaseTime, governorVaa.Amount)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	// 5. delete governor vaas.
	for _, vaaID := range governorVaaIdsToDelete {
		_, err = tx.Exec(ctx, `DELETE FROM wormhole.governor_vaas WHERE _id = $1`, vaaID)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	// Commit transaction.
	err = tx.Commit(ctx) // TODO retry commit
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return nil
}
