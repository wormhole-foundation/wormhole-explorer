package observations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPostgresRepository(db *sql.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db, logger: logger}
}

// Find get a list of ObservationDoc pointers.
// The input parameter [q *ObservationQuery] define the filters to apply in the query.
func (r *PostgresRepository) Find(ctx context.Context, q *ObservationQuery) ([]*ObservationDoc, error) {

	return obs, err
}

func (q *ObservationQuery) toWhere() (string, []any) {
	var conditions []string
	var params []any
	if q.chainId > 0 {
		condition := fmt.Sprintf("emitter_chain_id = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.chainId)
	}
	if q.emitter != "" {
		condition := fmt.Sprintf("emitter_address = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.emitter)
	}
	if q.sequence != "" {
		condition := fmt.Sprintf("sequence = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.sequence)
	}
	if len(q.hash) > 0 {
		condition := fmt.Sprintf("hash = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, utils.NormalizeBytesToHex(q.hash))
	}
	if q.guardianAddr != "" {
		condition := fmt.Sprintf("guardian_address = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, utils.NormalizeHex(q.guardianAddr))
	}
	if q.txHash != nil {
		condition := fmt.Sprintf("hash = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, utils.NormalizeBytesToHex(q.hash))
	}
	if q.txHash != nil {
		nativeTxHash := q.txHash.String()
		r = append(r, bson.E{"nativeTxHash", nativeTxHash})
	}

	return &r
}
