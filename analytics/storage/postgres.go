package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

// PostgresPricesRepository is a storage repository.
type PostgresPricesRepository struct {
	db     *db.DB
	logger *zap.Logger
}

// NewPostgresRepository creates a new storage repository.
func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresPricesRepository {
	return &PostgresPricesRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresPricesRepository) Upsert(ctx context.Context, op OperationPrice) error {
	query := `
		INSERT INTO wormhole.wh_operation_prices 
		(id, vaa_id, token_chain_id, token_address, coingecko_id, symbol, token_usd_price, total_token, total_usd, "timestamp", created_at, updated_at)  
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
		ON CONFLICT(id) DO UPDATE 
		SET updated_at = $12
		RETURNING updated_at;
		`

	now := time.Now()
	var result *time.Time
	err := r.db.ExecAndScan(ctx,
		&result,
		query,
		op.Digest,
		op.VaaID,
		op.TokenChainID,
		op.TokenAddress,
		op.CoinGeckoID,
		op.Symbol,
		op.TokenUSDPrice,
		op.TotalToken,
		op.TotalUSD,
		op.Timestamp,
		now,
		now)

	return err
}

// PostgresPricesRepository is a storage repository.
type PostgresVaaRepository struct {
	db     *db.DB
	logger *zap.Logger
}

// VaaDoc vaa document struct definition.
type AttestationVaa struct {
	ID    string `db:"id"`
	VaaID string `db:"vaa_id"`
	Raw   []byte `db:"raw"`
}

// NewRepository create a new Repository.
func NewPostgresVaaRepository(db *db.DB, logger *zap.Logger) *PostgresVaaRepository {
	return &PostgresVaaRepository{
		db:     db,
		logger: logger.With(zap.String("module", "PostgresVaaRepository")),
	}
}

// FindById find a vaa by id.
func (r *PostgresVaaRepository) FindByVaaID(ctx context.Context, vaaID string) (*Vaa, error) {

	query := `
		SELECT id, vaa_id, raw
		FROM wormhole.wh_attestation_vaas
		WHERE vaa_id = $1 AND active = true;`

	var AttestationVaas []*AttestationVaa
	err := r.db.Select(ctx, &AttestationVaas, query, vaaID)
	if err != nil {
		r.logger.Error("Error finding vaas by vaaID",
			zap.String("vaaId", vaaID),
			zap.Error(err))
		return nil, err
	}

	if len(AttestationVaas) == 0 {
		return nil, nil
	}

	if len(AttestationVaas) > 1 {
		r.logger.Error("Error finding vaas by vaaID",
			zap.String("vaaId", vaaID),
			zap.Error(err))
		return nil, err
	}

	return &Vaa{
		ID:    AttestationVaas[0].ID,
		VaaID: AttestationVaas[0].VaaID,
		Vaa:   AttestationVaas[0].Raw,
	}, nil

}

func (r *PostgresVaaRepository) FindPage(ctx context.Context, q VaaPageQuery, pagination Pagination) ([]*Vaa, error) {

	var conditions []string
	var params []any
	if q.StartTime != nil {
		condition := fmt.Sprintf("timestamp >= $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.StartTime)
	}
	if q.EndTime != nil {
		condition := fmt.Sprintf("timestamp <= $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.EndTime)
	}
	if q.EmitterChainID != nil {
		condition := fmt.Sprintf("emitter_chain_id = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.EmitterChainID)
	}
	if q.EmitterAddress != nil {
		condition := fmt.Sprintf("emitter_address = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.EmitterAddress)
	}
	if q.Sequence != nil {
		condition := fmt.Sprintf("sequence = $%d", len(params)+1)
		conditions = append(conditions, condition)
		params = append(params, q.Sequence)
	}

	where := "1 = 1"
	if len(conditions) > 0 {
		where = strings.Join(conditions, " AND ")
	}

	sort := "DESC"
	if pagination.SortAsc {
		sort = "ASC"
	}

	params = append(params, pagination.Page*pagination.PageSize, pagination.PageSize)

	query := fmt.Sprintf(`
	SELECT id, vaa_id, raw
	FROM wormhole.wh_attestation_vaas
	WHERE %s AND active = true
	ORDER BY timestamp %s
	OFFSET $%d
	LIMIT $%d`, where, sort, len(params)+1, len(params)+2)

	var AttestationVaas []*AttestationVaa
	err := r.db.Select(ctx, &AttestationVaas, query, params...)
	if err != nil {
		r.logger.Error("Error finding by page",
			zap.Error(err))
		return nil, err
	}

	vaas := make([]*Vaa, 0, len(AttestationVaas))
	for _, v := range AttestationVaas {
		vaas = append(vaas, &Vaa{
			ID:    v.ID,
			VaaID: v.VaaID,
			Vaa:   v.Raw,
		})
	}
	return vaas, nil
}
