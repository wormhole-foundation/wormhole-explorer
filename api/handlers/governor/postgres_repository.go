package governor

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/mongo"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type governorLimitResult struct {
	ChainID            uint16 `db:"chain_id"`
	NotionalLimit      uint64 `db:"notional_limit"`
	BigTransactionSize uint64 `db:"big_transaction_size"`
	AvailableNotional  uint64 `db:"available_notional"`
}

type governorStatusResult struct {
	ID           string    `db:"id"`
	GuardianName string    `db:"guardian_name"`
	Message      string    `db:"message"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db,
		logger: logger.With(zap.String("module", "PostgresObservationsRepository"))}
}

func (r *PostgresRepository) GetGovernorLimit(
	ctx context.Context,
	q *GovernorQuery,
) ([]*GovernorLimit, error) {

	query := `
    SELECT 
    	(chains->'chainId')::int AS chain_id, 
    	COALESCE(c.notional_limit,0) AS notional_limit, 
		COALESCE(c.big_transaction_size,0) AS big_transaction_size, 
		(chains->'remainingAvailableNotional')::int AS available_notional
    FROM wormholescan.wh_governor_status AS s
    CROSS JOIN LATERAL jsonb_array_elements(s.message->'chains') AS chains
    JOIN wormholescan.wh_governor_config_chains c ON s.id = c.governor_config_id AND c.chain_id = (chains->'chainId')::int
    ORDER BY c.chain_id
	`

	var result []governorLimitResult
	err := r.db.Select(ctx, &result, query)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return nil, err
	}

	limitsByChainID := make(map[uint16][]governorLimitResult)
	for _, r := range result {
		limitsByChainID[r.ChainID] = append(limitsByChainID[r.ChainID], r)
	}

	limits := make([]*GovernorLimit, 0, len(limitsByChainID))
	for chainID, chainLimits := range limitsByChainID {
		limits = append(limits, createGovernorLimit(chainID, chainLimits))
	}

	sort.Slice(limits, func(i, j int) bool {
		if q.SortOrder == "ASC" {
			return limits[i].ChainID < limits[j].ChainID
		}
		return limits[i].ChainID > limits[j].ChainID
	})

	return paginate(limits, int(q.Skip), int(q.Limit)), nil
}

func (r *PostgresRepository) FindGovernorStatus(
	ctx context.Context,
	q *GovernorQuery,
) ([]*GovStatus, error) {
	query, params := q.toQuery()
	var result []governorStatusResult
	err := r.db.Select(ctx, &result, query, params...)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return nil, err
	}

	statuses := make([]*GovStatus, 0, len(result))
	for _, s := range result {
		govStatus, err := createGovStatus(&s)
		if err != nil {
			r.logger.Error("creating govStatus", zap.Error(err), zap.String("guardian_name", s.GuardianName))
			return nil, err
		}
		statuses = append(statuses, govStatus)
	}

	return statuses, nil
}

func (r *PostgresRepository) FindOneGovernorStatus(
	ctx context.Context,
	q *GovernorQuery,
) (*GovStatus, error) {
	var result governorStatusResult
	query, params := q.toQuery()
	err := r.db.SelectOne(ctx, &result, query, params...)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return nil, err
	}
	return createGovStatus(&result)
}

func (q *GovernorQuery) toQuery() (string, []any) {
	var params []any
	query := "SELECT id, guardian_name, message , created_at , updated_at FROM wormholescan.wh_governor_status \n "
	if q.id != nil {
		params = append(params, q.id.ShortHex())
		query += "WHERE id = $1 \n "
	} else {
		params = append(params, q.Limit, q.Skip)
		query += "ORDER BY id ASC \n "
		query += "LIMIT $1 OFFSET $2 \n "
	}
	return query, params
}

func createGovernorLimit(chainID uint16, chainLimits []governorLimitResult) *GovernorLimit {
	sort.Slice(chainLimits, func(i, j int) bool {
		return chainLimits[i].NotionalLimit > chainLimits[j].NotionalLimit
	})
	var notionalLimit mongo.Uint64
	if len(chainLimits) >= minGuardianNum {
		notionalLimit = mongo.Uint64(chainLimits[minGuardianNum-1].NotionalLimit)
	}

	sort.Slice(chainLimits, func(i, j int) bool {
		return chainLimits[i].BigTransactionSize > chainLimits[j].BigTransactionSize
	})
	var bigTransactionSize mongo.Uint64
	if len(chainLimits) >= minGuardianNum {
		bigTransactionSize = mongo.Uint64(chainLimits[minGuardianNum-1].BigTransactionSize)
	}

	sort.Slice(chainLimits, func(i, j int) bool {
		return chainLimits[i].AvailableNotional > chainLimits[j].AvailableNotional
	})
	var availableNotional mongo.Uint64
	if len(chainLimits) >= minGuardianNum {
		availableNotional = mongo.Uint64(chainLimits[minGuardianNum-1].AvailableNotional)
	}

	return &GovernorLimit{
		ChainID:            sdk.ChainID(chainID),
		NotionalLimit:      notionalLimit,
		MaxTransactionSize: bigTransactionSize,
		AvailableNotional:  availableNotional,
	}
}

func createGovStatus(s *governorStatusResult) (*GovStatus, error) {

	var wrapper struct {
		Chains []*GovStatusChains `json:"chains"`
	}
	err := json.Unmarshal([]byte(s.Message), &wrapper)
	if err != nil {
		return nil, err
	}
	return &GovStatus{
		ID:        s.ID,
		NodeName:  s.GuardianName,
		Chains:    wrapper.Chains,
		CreatedAt: &s.CreatedAt,
		UpdatedAt: &s.UpdatedAt,
	}, nil
}

func paginate(list []*GovernorLimit, skip int, size int) []*GovernorLimit {
	if skip > len(list) {
		skip = len(list)
	}

	end := skip + size
	if end > len(list) {
		end = len(list)
	}

	return list[skip:end]
}
