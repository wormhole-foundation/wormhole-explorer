package repository

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

// PostgresVaaRepository is a repository for VAA.
type PostgresVaaRepository struct {
	db     *db.DB
	logger *zap.Logger
}

// NewPostgresVaaRepository creates a new Vaa repository.
func NewPostgresVaaRepository(db *db.DB, logger *zap.Logger) *PostgresVaaRepository {
	return &PostgresVaaRepository{db: db, logger: logger}
}

// FindPage finds VAA by query and pagination.
func (r *PostgresVaaRepository) FindPage(ctx context.Context, queryFilters VaaQuery,
	pagination Pagination) ([]*AttestationVaa, error) {

	query := `SELECT id, vaa_id, version, emitter_chain_id, emitter_address, sequence, guardian_set_index, raw, timestamp, active, is_duplicated, created_at, updated_at 
	FROM wormhole.wh_attestation_vaas`

	// build query by args
	var args []interface{}
	var argIdx int = 1
	var conditions string

	// add time filters
	if queryFilters.StartTime != nil {
		if conditions != "" {
			conditions += " AND "
		}
		conditions += fmt.Sprintf("timestamp >= $%d", argIdx)
		args = append(args, *queryFilters.StartTime)
		argIdx++
	}

	if queryFilters.EndTime != nil {
		if conditions != "" {
			conditions += " AND "
		}
		conditions += fmt.Sprintf("timestamp < $%d", argIdx)
		args = append(args, *queryFilters.EndTime)
		argIdx++
	}

	// add emitter chain id filter
	if queryFilters.EmitterChainID != nil {
		if conditions != "" {
			conditions += " AND "
		}
		conditions += fmt.Sprintf("emitter_chain_id = $%d", argIdx)
		args = append(args, *queryFilters.EmitterChainID)
		argIdx++
	}

	// add emitter address filter
	if queryFilters.EmitterAddress != nil {
		if conditions != "" {
			conditions += " AND "
		}
		conditions += fmt.Sprintf("emitter_address = $%d", argIdx)
		args = append(args, *queryFilters.EmitterAddress)
		argIdx++
	}

	// add sequence filter
	if queryFilters.Sequence != nil {
		if conditions != "" {
			conditions += " AND "
		}
		conditions += fmt.Sprintf("sequence = $%d", argIdx)
		args = append(args, *queryFilters.Sequence)
		argIdx++
	}

	// add conditions to query
	if conditions != "" {
		query = fmt.Sprintf("%s WHERE %s", query, conditions)
	}

	// add order and limit to query
	sortOrder := "DESC"
	if pagination.SortAsc {
		sortOrder = "ASC"
	}

	query = fmt.Sprintf("%s ORDER BY timestamp %s LIMIT $%d OFFSET $%d",
		query, sortOrder, argIdx, argIdx+1)

	// add pagination args
	args = append(args, pagination.PageSize, pagination.Page*pagination.PageSize)

	// execute query
	var attestationVaas []*AttestationVaa
	err := r.db.Select(ctx, &attestationVaas, query, args...)
	if err != nil {
		return nil, err
	}

	return attestationVaas, nil
}
