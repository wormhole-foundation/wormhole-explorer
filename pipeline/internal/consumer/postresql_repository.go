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
	err := r.db.SelectOne(ctx, &txHash, "SELECT tx_hash FROM wormhole.wh_observations WHERE wh_observations.hash = $1", vaaDigest)
	return txHash, err
}

func (r *PostreSqlRepositoryImpl) CreateOperationTransaction(ctx context.Context, opTx OperationTransaction) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO wormhole.wh_operation_transactions (chain_id, tx_hash, type, created_at, updated_at, attestation_vaas_id, vaa_id, from_address, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6,$7,$8,$9)`,
		opTx.ChainID,
		opTx.TxHash,
		opTx.Type,
		opTx.CreatedAt,
		opTx.UpdatedAt,
		opTx.AttestationVaaID,
		opTx.VaaID,
		opTx.FromAddress,
		opTx.Timestamp)
	return err
}
