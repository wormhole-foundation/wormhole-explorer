package storage

import (
	"context"
	"errors"
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
	FROM wormhole.wh_governor_config_chains 
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
	_, err = tx.Exec(ctx, `DELETE FROM wormhole.wh_governor_config_chains WHERE governor_config_id = $1`, nodeAddress)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	// Insert new governor config chains.
	now := time.Now()
	for _, chain := range chains {
		_, err = tx.Exec(ctx, `
		INSERT INTO wormhole.wh_governor_config_chains (governor_config_id, chain_id, notional_limit, big_transaction_size, created_at, updated_at)
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
	query := `SELECT * FROM wormhole.wh_guardian_governor_vaas WHERE guardian_address = $1`
	var rows []NodeGovernorVaa
	err := r.db.Select(ctx, &rows, query, nodeAddress)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) FindNodeGovernorVaaByVaaID(ctx context.Context, vaaID string) ([]NodeGovernorVaa, error) {
	query := `SELECT * FROM wormhole.wh_guardian_governor_vaas WHERE vaa_id = $1`
	var rows []NodeGovernorVaa
	err := r.db.Select(ctx, &rows, query, vaaID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) FindNodeGovernorVaaByVaaIDs(ctx context.Context, vaaID []string) ([]NodeGovernorVaa, error) {
	query := `SELECT * FROM wormhole.wh_guardian_governor_vaas WHERE vaa_id = ANY($1)`
	var rows []NodeGovernorVaa
	err := r.db.Select(ctx, &rows, query, vaaID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) FindGovernorVaaByVaaIDs(ctx context.Context, vaaID []string) ([]GovernorVaa, error) {
	query := `SELECT * FROM wormhole.wh_governor_vaas WHERE id = ANY($1)`
	var rows []GovernorVaa
	err := r.db.Select(ctx, &rows, query, vaaID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) UpdateGovernorStatus(
	ctx context.Context,
	nodeGovernorVaaDocToInsert []NodeGovernorVaa,
	nodeGovernorVaaDocToDelete []string,
	governorVaasToInsert []GovernorVaa,
	governorVaaIdsToDelete []string) error {

	// 1. Start transaction.
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	// 2. insert node governor vaas.
	now := time.Now()
	for _, nodeGovernorVaa := range nodeGovernorVaaDocToInsert {
		_, err = tx.Exec(ctx, `
		INSERT INTO wormhole.wh_guardian_governor_vaas (guardian_address, guardian_name, vaa_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`,
			nodeGovernorVaa.NodeAddress, nodeGovernorVaa.NodeName, nodeGovernorVaa.VaaID, now, now)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	// 3. delete node governor vaas.
	for _, vaaID := range nodeGovernorVaaDocToDelete {
		_, err = tx.Exec(ctx, `DELETE FROM wormhole.wh_guardian_governor_vaas WHERE vaa_id = $1`, vaaID)
		if err != nil {
			_ = tx.Rollback
			return err
		}
	}

	// 4. insert governor vaas.
	for _, governorVaa := range governorVaasToInsert {
		_, err = tx.Exec(ctx, `
		INSERT INTO wormhole.wh_governor_vaas (id, chain_id, emitter_address, sequence, tx_hash, release_time, notional_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			governorVaa.ID, governorVaa.ChainID, governorVaa.EmitterAddress, governorVaa.Sequence, governorVaa.TxHash, governorVaa.ReleaseTime, governorVaa.Amount, now, now)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}

	// 5. delete governor vaas.
	for _, vaaID := range governorVaaIdsToDelete {
		_, err = tx.Exec(ctx, `DELETE FROM wormhole.wh_governor_vaas WHERE id = $1`, vaaID)
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

// FindActiveAttestationVaaByVaaID finds active attestation vaa by vaa id.
func (r *PostgresRepository) FindActiveAttestationVaaByVaaID(ctx context.Context, vaaID string) (*AttestationVaa, error) {
	query := `SELECT * FROM wormhole.wh_attestation_vaas WHERE vaa_id = $1 AND is_active = true`
	var rows []*AttestationVaa
	err := r.db.Select(ctx, &rows, query, vaaID)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		r.logger.Error("attestation vaa not found", zap.String("vaaID", vaaID))
		return nil, errors.New("attestation vaa not found")
	}

	if len(rows) > 1 {
		r.logger.Error("only one vaa can be active", zap.String("vaaID", vaaID))
		return nil, errors.New("only one vaa can be active")
	}

	return rows[0], nil
}

// FindAttestationVaaByVaaId finds attestation vaa by vaa id.
func (r *PostgresRepository) FindAttestationVaaByVaaId(ctx context.Context, vaaID string) ([]AttestationVaa, error) {
	query := `SELECT * FROM wormhole.wh_attestation_vaas WHERE vaa_id = $1`
	var rows []AttestationVaa
	err := r.db.Select(ctx, &rows, query, vaaID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// FixActiveVaa fixes active vaa.
func (r *PostgresRepository) FixActiveVaa(ctx context.Context, id string, vaaID string) error {
	// Start transaction.
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	// update all the attestation_vaa exlude the one with the id the field is_active to false and is_duplicated to true.
	_, err = tx.Exec(ctx, `
	UPDATE wormhole.wh_attestation_vaas
	SET is_active = false, is_duplicated = true
	WHERE vaa_id = $1 AND id != $2`, vaaID, id)
	if err != nil {
		_ = tx.Rollback
		return err
	}

	// update the atteation_vaa with id the field is_active to true and is_duplicated to true.
	_, err = tx.Exec(ctx, `
	UPDATE wormhole.wh_attestation_vaas
	SET is_active = true, is_duplicated = true
	WHERE id = $1`, id)
	if err != nil {
		_ = tx.Rollback
		return err
	}

	// Commit transaction.
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
