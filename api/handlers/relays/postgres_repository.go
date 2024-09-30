package relays

import (
	"context"
	"fmt"
	"time"

	"encoding/json"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

type relaysResult struct {
	VaaID       string          `db:"vaa_id"`
	Relayer     string          `db:"relayer"`
	Event       string          `db:"event"`
	Status      string          `db:"status"`
	ReceivedAt  *time.Time      `db:"received_at"`
	CompletedAt *time.Time      `db:"completed_at"`
	FailedAt    *time.Time      `db:"failed_at"`
	FromTxHash  string          `db:"from_tx_hash"`
	ToTxHash    string          `db:"to_tx_hash"`
	Message     json.RawMessage `db:"message"`
	Signature   []byte          `db:"signature"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   *time.Time      `db:"updated_at"`
}

type dataWrapper struct {
	Data RelayData `json:"data"`
}

func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db,
		logger: logger.With(zap.String("module", "PostgresRelaysRepository"))}
}

func (r *PostgresRepository) FindOne(ctx context.Context, q *RelaysQuery) (*RelayDoc, error) {
	var result relaysResult
	query, params := q.toQuery()
	err := r.db.SelectOne(ctx, &result, query, params...)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get relays",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	var wrapper dataWrapper
	err = json.Unmarshal(result.Message, &wrapper)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed to unmarshal message", zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	response := &RelayDoc{
		ID:     result.VaaID,
		Event:  result.Event,
		Origin: result.Relayer,
		Data:   wrapper.Data,
	}

	return response, nil
}

func (q *RelaysQuery) toQuery() (string, []any) {

	query :=
		`SELECT vaa_id, relayer, "event", status, received_at, completed_at, failed_at, from_tx_hash, to_tx_hash, message, created_at, updated_at
     FROM wormholescan.wh_relays
     WHERE vaa_id = $1`

	vaaID := fmt.Sprintf("%d/%s/%s", q.chainId, q.emitter, q.sequence)

	return query, []any{vaaID}

}
