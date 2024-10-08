package observations

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type observationResult struct {
	ID             string     `db:"id"`
	EmitterChainID uint16     `db:"emitter_chain_id"`
	EmitterAddress string     `db:"emitter_address"`
	Sequence       string     `db:"sequence"`
	Hash           string     `db:"hash"`
	TxHash         string     `db:"tx_hash"`
	GuardianAddr   string     `db:"guardian_address"`
	Signature      []byte     `db:"signature"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      *time.Time `db:"updated_at"`
}

type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db,
		logger: logger.With(zap.String("module", "PostgresObservationsRepository"))}
}

func (r *PostgresRepository) Find(ctx context.Context, q *ObservationQuery) ([]*ObservationDoc, error) {
	query, params := q.toQuery()
	var obs []*observationResult
	err := r.db.Select(ctx, &obs, query, params...)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return nil, err
	}

	result := make([]*ObservationDoc, 0, len(obs))
	for _, o := range obs {
		hash, err := hex.DecodeString(o.Hash)
		if err != nil {
			r.logger.Error("failed to decode hash", zap.Error(err), zap.String("hash", o.Hash))
		}
		parsedTxHash, err := types.ParseTxHash(o.TxHash)
		if err != nil {
			r.logger.Error("failed to parse tx hash", zap.Error(err), zap.String("tx_hash", o.TxHash))
		}
		var txHash []byte
		if parsedTxHash != nil {
			txHash = parsedTxHash.Binary()
		}
		result = append(result, &ObservationDoc{
			ID:           o.ID,
			EmitterChain: sdk.ChainID(o.EmitterChainID),
			EmitterAddr:  o.EmitterAddress,
			Sequence:     o.Sequence,
			Hash:         hash,
			TxHash:       txHash,
			GuardianAddr: utils.NormalizeHex(o.GuardianAddr),
			Signature:    o.Signature,
			IndexedAt:    &o.CreatedAt,
			UpdatedAt:    o.UpdatedAt,
		})
	}

	return result, nil
}

func (r *PostgresRepository) FindOne(ctx context.Context, q *ObservationQuery) (*ObservationDoc, error) {
	result, err := r.Find(ctx, q)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get observations",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	if len(result) == 0 {
		return nil, errs.ErrNotFound
	}
	if len(result) > 1 {
		return nil, errs.ErrInternalError
	}
	return result[0], nil
}

func (q *ObservationQuery) toQuery() (string, []any) {
	var conditions []string
	var params []any
	if q.chainId > 0 {
		condition := fmt.Sprintf("obs.emitter_chain_id = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.chainId)
	}
	if q.emitter != "" {
		condition := fmt.Sprintf("obs.emitter_address = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.emitter)
	}
	if q.sequence != "" {
		condition := fmt.Sprintf("obs.sequence = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.sequence)
	}
	if len(q.hash) > 0 {
		condition := fmt.Sprintf("obs.hash = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, utils.NormalizeBytesToHex(q.hash))
	}
	if q.guardianAddr != "" {
		condition := fmt.Sprintf("obs.guardian_address = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, utils.NormalizeHex(q.guardianAddr))
	}
	if q.hash != nil {
		condition := fmt.Sprintf("obs.hash = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, utils.NormalizeBytesToHex(q.hash))
	}
	hasTxHash := false
	if q.txHash != nil {
		txHash := q.txHash.String()
		conditions = append(conditions, fmt.Sprintf("ot.tx_hash = $%d", len(params)+1))
		params = append(params, txHash)
		hasTxHash = true
	}

	where := "1 = 1"
	if len(conditions) > 0 {
		where = strings.Join(conditions, " AND ")
	}

	query := "SELECT obs.* FROM wormholescan.wh_observations obs \n"
	if hasTxHash {
		query += "JOIN wormholescan.wh_operation_transactions ot ON obs.hash = ot.attestation_vaas_id \n"
	}

	query += fmt.Sprintf("WHERE %s \n", where)
	query += fmt.Sprintf("ORDER BY obs.created_at %s \n", q.SortOrder)
	query += fmt.Sprintf("LIMIT %d OFFSET %d", q.Limit, q.Skip)

	return query, params

}
