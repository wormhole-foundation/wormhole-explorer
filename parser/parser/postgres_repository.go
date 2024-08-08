package parser

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

// PostgresRepository is a postgres repository.
type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

// NewPostgresRepository creates a new postgres repository.
func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger}
}

// UpsertAttestationVaaProperties upserts attestation vaa properties.
func (r *PostgresRepository) UpsertAttestationVaaProperties(ctx context.Context,
	attestationVaaProperites AttestationVaaProperties) error {

	now := time.Now()

	query := `INSERT INTO wormhole.wh_attestation_vaa_properties (id, vaa_id, app_id, payload, 
	raw_standard_fields, from_chain_id, from_address, to_chain_id, to_address, token_chain_id,
	token_address, amount, fee_chain_id, fee_address, fee, "timestamp", created_at, 
	updated_at) VAlUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
	$17, $18) ON CONFLICT (id) DO UPDATE SET vaa_id = $2, app_id = $3, payload = $4, 
	raw_standard_fields = $5, from_chain_id = $6, from_address = $7, to_chain_id = $8, 
	to_address = $9, token_chain_id = $10, token_address = $11, amount = $12, fee_chain_id = $13,
	fee_address = $14, fee = $15, "timestamp" = $16, updated_at = $17`

	_, err := r.db.Exec(ctx,
		query,
		attestationVaaProperites.ID,
		attestationVaaProperites.VaaID,
		attestationVaaProperites.AppID,
		attestationVaaProperites.Payload,
		attestationVaaProperites.RawStandardFields,
		attestationVaaProperites.FromChainID,
		attestationVaaProperites.FromAddress,
		attestationVaaProperites.ToChainID,
		attestationVaaProperites.ToAddress,
		attestationVaaProperites.TokenChainID,
		attestationVaaProperites.TokenAddress,
		attestationVaaProperites.Amount,
		attestationVaaProperites.FeeChainID,
		attestationVaaProperites.FeeAddress,
		attestationVaaProperites.Fee,
		attestationVaaProperites.Timestamp,
		now,
		now)

	if err != nil {
		r.logger.Error("Error upserting attestation vaa properties", zap.Error(err))
		return err
	}
	return nil
}
