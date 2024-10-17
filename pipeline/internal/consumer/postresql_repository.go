package consumer

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
)

type PostreSqlRepository interface {
	GetTxHash(ctx context.Context, vaaDigest string) (string, error)
	CreateOperationTransaction(ctx context.Context, opTx OperationTransaction) error
}

type PostreSqlRepositoryImpl struct {
	db *db.DB
}

func NewPostreSqlRepository(db *db.DB) *PostreSqlRepositoryImpl {
	return &PostreSqlRepositoryImpl{db: db}
}

func (r *PostreSqlRepositoryImpl) GetTxHash(ctx context.Context, vaaDigest string) (string, error) {
	var txHash string
	err := r.db.SelectOne(ctx, &txHash, "SELECT tx_hash FROM wormholescan.wh_observations WHERE wh_observations.hash = $1 LIMIT 1", vaaDigest)
	return txHash, err
}

func (r *PostreSqlRepositoryImpl) CreateOperationTransaction(ctx context.Context, opTx OperationTransaction) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO wormholescan.wh_operation_transactions (chain_id, tx_hash, type, created_at,
		 updated_at, attestation_id, message_id, from_address, timestamp, source_event, track_id_event)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (message_id, tx_hash) DO NOTHING`,
		opTx.ChainID,
		opTx.TxHash,
		opTx.Type,
		opTx.CreatedAt,
		opTx.UpdatedAt,
		opTx.AttestationVaaID,
		opTx.VaaID,
		opTx.FromAddress,
		opTx.Timestamp,
		opTx.Source,
		opTx.TrackID)
	return err
}
