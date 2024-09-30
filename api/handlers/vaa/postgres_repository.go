package vaa

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// PostgresRepository definition.
type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

// NewPostgresRepository creates a new PostgresRepository.
func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db, logger: logger}
}

// VaaQuery represents a query for the vaa postgres document.
type vaaResult struct {
	ID               string     `db:"id"`
	VaaID            string     `db:"vaa_id"`
	Version          uint8      `db:"version"`
	EmitterChainId   uint16     `db:"emitter_chain_id"`
	EmitterAddress   string     `db:"emitter_address"`
	Sequence         string     `db:"sequence"`
	GuardianSetIdx   uint32     `db:"guardian_set_index"`
	Raw              []byte     `db:"raw"`
	Timestamp        time.Time  `db:"timestamp"`
	Active           bool       `db:"active"`
	IsDuplicated     bool       `db:"is_duplicated"`
	ConsistencyLevel *uint8     `db:"consistency_level"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        *time.Time `db:"updated_at"`
	// field TxHash belongs to table wh_operation_transactions.
	// It was added for compatibility.
	TxHash *string `db:"tx_hash"`
	// field ParsedPayload belongs to table wh_attestation_vaa_properties.
	// It was added for compatibility.
	ParsedPayload map[string]any `db:"payload"`
}

// FindVaas finds VAAs.
func (p *PostgresRepository) Find(
	ctx context.Context,
	q *VaaQuery,
) ([]*VaaDoc, error) {
	query, param := q.toQuery()
	var vaas []*vaaResult
	err := p.db.Select(ctx, &vaas, query, param...)
	if err != nil {
		p.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return nil, err
	}
	result := make([]*VaaDoc, 0, len(vaas))
	for _, v := range vaas {

		// calculate emitter native addres
		var emitterNativeAddress string
		emitterNativeAddress, err := domain.TranslateEmitterAddress(sdk.ChainID(v.EmitterChainId), v.EmitterAddress)
		if err != nil {
			p.logger.Warn("failed to translate emitter address for VAA",
				zap.Stringer("emitterChain", sdk.ChainID(v.EmitterChainId)),
				zap.String("emitterAddr", v.EmitterAddress),
				zap.Error(err),
			)
		}

		result = append(result, &VaaDoc{
			ID:                v.VaaID,
			Version:           v.Version,
			EmitterChain:      sdk.ChainID(v.EmitterChainId),
			EmitterAddr:       v.EmitterAddress,
			EmitterNativeAddr: emitterNativeAddress,
			Sequence:          v.Sequence,
			GuardianSetIndex:  v.GuardianSetIdx,
			Vaa:               v.Raw,
			TxHash:            v.TxHash,
			Digest:            v.ID,
			IsDuplicated:      v.IsDuplicated,
			Payload:           v.ParsedPayload, // CHECK THIS IN MONGO
			Timestamp:         &v.CreatedAt,
			IndexedAt:         &v.CreatedAt,
			UpdatedAt:         v.UpdatedAt,
		})
	}
	return result, nil
}

func (q *VaaQuery) toQuery() (string, []any) {
	if len(q.ids) > 0 {
		// query by ids
		// includeParsedPayload is optional.
		return q.toQueryByIds()
	}

	if q.txHash != "" {
		// query by txHash
		// includeParsedPayload is optional.
		return q.toQueryByTxHash()
	}

	if q.toChain > 0 {
		// query by chainId, emitter, toChain.
		// includeParsedPayload is optional.
		return q.toQueryByToChain()
	}

	if q.duplicated {
		// query duplicated vaas by chainId, emitter, sequence.
		return q.toQueryDuplicatedVaaId()
	}

	// query all vaas with filters.
	// chain, emitter, includeParsedPayload are optional.
	return q.toQueryWithFilters()
}

func (q VaaQuery) toQueryByIds() (string, []any) {
	query := "SELECT v.id, v.vaa_id, v.version, v.emitter_chain_id, v.emitter_address, v.sequence, v.guardian_set_index, v.raw, v.timestamp"
	query += ", v.active, v.is_duplicated, v.created_at, v.updated_at, v.consistency_level, t.tx_hash"

	if q.includeParsedPayload {
		query += ", p.payload \n"
	} else {
		query += "\n"
	}

	query += "FROM wormholescan.wh_attestation_vaas as v \n"
	query += "LEFT JOIN wormholescan.wh_operation_transactions as t ON v.id = t.attestation_vaas_id \n"

	if q.includeParsedPayload {
		query += "LEFT JOIN wormholescan.wh_attestation_vaa_properties as p ON v.id = p.id \n"
	}

	conditions := make([]string, 0, len(q.ids))
	params := make([]any, 0, len(q.ids))
	for i, id := range q.ids {
		conditions = append(conditions, fmt.Sprintf("v.vaa_id = $%d", i+1))
		params = append(params, id)
	}

	query += "WHERE " + strings.Join(conditions, " OR ") + "\n"
	query += fmt.Sprintf("ORDER BY v.timestamp %s \n", q.SortOrder)
	query += fmt.Sprintf("LIMIT %d OFFSET %d", q.Limit, q.Skip)
	return query, params
}

func (q *VaaQuery) toQueryByTxHash() (string, []any) {
	query := "SELECT v.id, v.vaa_id, v.version, v.emitter_chain_id, v.emitter_address, v.sequence, v.guardian_set_index, v.raw, v.timestamp"
	query += ", v.active, v.is_duplicated, v.created_at, v.updated_at, v.consistency_level, t.tx_hash"

	if q.includeParsedPayload {
		query += ", p.payload \n"
	} else {
		query += "\n"
	}

	query += "FROM wormholescan.wh_operation_transactions as t \n"
	query += "INNER JOIN wormholescan.wh_attestation_vaas as v ON v.id = t.attestation_vaas_id AND v.active = TRUE \n"

	if q.includeParsedPayload {
		query += "LEFT JOIN wormholescan.wh_attestation_vaa_properties as p ON t.attestation_vaas_id = p.id \n"
	}

	query += "WHERE t.tx_hash = $1 \n"
	query += fmt.Sprintf("ORDER BY t.timestamp %s \n", q.SortOrder)
	query += fmt.Sprintf("LIMIT %d OFFSET %d", q.Limit, q.Skip)
	return query, []any{q.txHash}
}

func (q *VaaQuery) toQueryWithFilters() (string, []any) {
	var params []any
	query := "SELECT v.id, v.vaa_id, v.version, v.emitter_chain_id, v.emitter_address, v.sequence, v.guardian_set_index, v.raw, v.timestamp"
	query += ", v.active, v.is_duplicated, v.created_at, v.updated_at, v.consistency_level, t.tx_hash"

	if q.includeParsedPayload {
		query += ", p.payload \n"
	} else {
		query += "\n"
	}

	query += "FROM wormholescan.wh_attestation_vaas as v \n"
	query += "LEFT JOIN wormholescan.wh_operation_transactions as t ON v.id = t.attestation_vaas_id \n"

	if q.includeParsedPayload {
		query += "LEFT JOIN wormholescan.wh_attestation_vaa_properties as p ON v.id = p.id \n"
	}

	query += "WHERE 1 = 1 \n"
	if q.chainId > 0 {
		query += "AND v.emitter_chain_id = $1 \n"
		params = append(params, q.chainId)
	}
	if q.emitter != "" {
		query += "AND v.emitter_address = $2 \n"
		params = append(params, q.emitter)
	}

	query += fmt.Sprintf("ORDER BY v.timestamp %s \n", q.SortOrder)
	query += fmt.Sprintf("LIMIT %d OFFSET %d", q.Limit, q.Skip)
	return query, params
}

func (q *VaaQuery) toQueryByToChain() (string, []any) {
	var params []any
	query := "SELECT v.id, v.vaa_id, v.version, v.emitter_chain_id, v.emitter_address, v.sequence, v.guardian_set_index, v.raw, v.timestamp"
	query += ", v.active, v.is_duplicated, v.created_at, v.updated_at, v.consistency_level, t.tx_hash"

	if q.includeParsedPayload {
		query += ", p.payload \n"
	} else {
		query += "\n"
	}

	query += "FROM wormholescan.wh_attestation_vaa_properties as p \n"
	query += "INNER JOIN wormholescan.wh_attestation_vaas as v ON v.id = p.id \n"
	query += "LEFT JOIN wormholescan.wh_operation_transactions as t ON p.id = t.attestation_vaas_id \n"

	if q.toChain > 0 {
		query += "WHERE p.to_chain_id = $1 \n"
		params = append(params, q.toChain)
	}

	query += fmt.Sprintf("ORDER BY v.timestamp %s \n", q.SortOrder)
	query += fmt.Sprintf("LIMIT %d OFFSET %d", q.Limit, q.Skip)
	return query, params
}

func (q *VaaQuery) toQueryDuplicatedVaaId() (string, []any) {
	query := `
	SELECT v.id, v.vaa_id, v.version, v.emitter_chain_id, v.emitter_address, v.sequence, v.guardian_set_index, v.raw, v.timestamp, 
	v.active, v.is_duplicated, v.created_at, v.updated_at, v.consistency_level, o.tx_hash 
	FROM wormholescan.wh_attestation_vaas as v
	INNER JOIN wormholescan.wh_observations as o ON v.id = o.hash
	WHERE v.vaa_id = $1 AND v.is_duplicated = TRUE
	`
	// query := "SELECT v.id, v.vaa_id, v.version, v.emitter_chain_id, v.emitter_address, v.sequence, v.guardian_set_index, v.raw, v.timestamp"
	// query += ", v.active, v.is_duplicated, v.created_at, v.updated_at, v.consistency_level, o.tx_hash"
	// query += "FROM wh_attestation_vaas as v \n"
	// query += "INNER JOIN wh_observations as o ON v.id = o.hash \n"
	// query += "WHERE v.vaa_id = $1 AND v.is_duplicated = TRUE \n"
	vaaId := fmt.Sprintf("%d/%s/%s", q.chainId, q.emitter, q.sequence)
	return query, []any{vaaId}
}
