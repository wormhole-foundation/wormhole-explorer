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

	query := `INSERT INTO wormholescan.wh_attestation_vaa_properties (id, message_id, app_id, payload,
	payload_type, raw_standard_fields, from_chain_id, from_address, to_chain_id, to_address, token_chain_id,
	token_address, amount, fee_chain_id, fee_address, fee, "timestamp", created_at, 
	updated_at) VAlUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
	$17, $18, $19) ON CONFLICT (id) DO UPDATE SET message_id = $2, app_id = $3, payload = $4, payload_type = $5,
	raw_standard_fields = $6, from_chain_id = $7, from_address = $8, to_chain_id = $9, 
	to_address = $10, token_chain_id = $11, token_address = $12, amount = $13, fee_chain_id = $14,
	fee_address = $15, fee = $16, "timestamp" = $17, updated_at = $18`

	_, err := r.db.Exec(ctx,
		query,
		attestationVaaProperites.ID,
		attestationVaaProperites.VaaID,
		attestationVaaProperites.AppID,
		attestationVaaProperites.Payload,
		attestationVaaProperites.PayloadType,
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
