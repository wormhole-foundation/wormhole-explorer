package vaa

import (
	"context"
	"errors"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db,
		logger: logger.With(zap.String("module", "PostgresRepository")),
	}
}

type AttestationVaa struct {
	ID             string      `db:"id"`
	VaaID          string      `db:"vaa_id"`
	Version        uint8       `db:"version"`
	EmitterChain   vaa.ChainID `db:"emitter_chain_id"`
	EmitterAddress string      `db:"emitter_address"`
	Sequence       uint64      `db:"sequence"`
	GuardianSetIdx uint32      `db:"guardian_set_index"`
	Raw            []byte      `db:"raw"`
	Timestamp      time.Time   `db:"timestamp"`
	Active         bool        `db:"active"`
	IsDuplicated   bool        `db:"is_duplicated"`
	CreatedAt      time.Time   `db:"created_at"`
	UpdatedAt      *time.Time  `db:"updated_at"`
}

// FindActiveAttestationVaaByVaaID finds active attestation vaa by vaa id.
func (r *PostgresRepository) FindActiveAttestationVaaByVaaID(ctx context.Context, vaaID string) (*AttestationVaa, error) {
	query := `SELECT id, vaa_id, version, emitter_chain_id, emitter_address, sequence, 
	guardian_set_index, raw, timestamp, active, is_duplicated, created_at, updated_at 
	FROM wormholescan.wh_attestation_vaas 
	WHERE vaa_id = $1 AND active = true`
	var rows []*AttestationVaa
	err := r.db.Select(ctx, &rows, query, vaaID)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		r.logger.Error("attestation_vaa not found", zap.String("vaaID", vaaID))
		return nil, errors.New("attestation vaa not found")
	}

	if len(rows) > 1 {
		r.logger.Error("only one attestation_vaa can be active", zap.String("vaaID", vaaID))
		return nil, errors.New("only attestation_vaa vaa can be active")
	}

	return rows[0], nil
}
