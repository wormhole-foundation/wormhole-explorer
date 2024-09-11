package vaa

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

type RepositoryPostreSQL struct {
	postreSQLClient *db.DB
	logger          *zap.Logger
}

func NewVaaRepositoryPostreSQL(postreSQLClient *db.DB, logger *zap.Logger) *RepositoryPostreSQL {
	return &RepositoryPostreSQL{
		postreSQLClient: postreSQLClient,
		logger:          logger,
	}
}

func (r *RepositoryPostreSQL) GetVaa(ctx context.Context, id string) (*VaaDoc, error) {
	res := &VaaDoc{}
	err := r.postreSQLClient.SelectOne(
		ctx,
		res,
		"SELECT id,vaa_id,raw as vaas,active FROM wormholescan.wh_attestation_vaas WHERE vaa_id = $1 and active = true",
		id)

	if err != nil {
		// fallback: in case the vaa is not found in the attestation_vaas table, try to find it in the operation_transactions table to grab the digest and tx_hash
		r.logger.Debug("Failed to get vaa from wh_attestation_vaas table", zap.Error(err), zap.String("vaa_id", id))
		err = r.postreSQLClient.SelectOne(
			ctx,
			res,
			"SELECT attestation_vaas_id as id, vaa_id, tx_hash FROM wormholescan.wh_operation_transactions WHERE vaa_id = $1 LIMIT 1", // LIMIT 1 is due to wormchain transactions which have 2 txs.
			id)
	}
	return res, err
}
