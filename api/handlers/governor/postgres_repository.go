package governor

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/mongo"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
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

type governorConfigResult struct {
	ID           string                  `db:"id"`
	GuardianName string                  `db:"guardian_name"`
	Counter      int64                   `db:"counter"`
	Timestamp    time.Time               `db:"timestamp"`
	Chains       []*governorConfigChains `db:"chains"`
	Tokens       []*governorConfigTokens `db:"tokens"`
	CreatedAt    *time.Time              `db:"created_at"`
	UpdatedAt    *time.Time              `db:"updated_at"`
}

type governorConfigChains struct {
	ChainID            vaa.ChainID `db:"chainId"`
	NotionalLimit      uint64      `db:"notionalLimit"`
	BigTransactionSize uint64      `db:"bigTransactionSize"`
}

type governorConfigTokens struct {
	OriginChainID int     `db:"originChainId"`
	OriginAddress string  `db:"originAddress"`
	Price         float32 `db:"price"`
}

type PostgresRepository struct {
	db     *db.DB
	logger *zap.Logger
}

func NewPostgresRepository(db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{db: db,
		logger: logger.With(zap.String("module", "PostgresObservationsRepository"))}
}

func (r *PostgresRepository) FindGovConfigurations(
	ctx context.Context,
	q *GovernorQuery,
) ([]*GovConfig, error) {

	// build query and params.
	query, params := buildGovConfigQuery(q)

	// execute the query.
	var result []governorConfigResult
	err := r.db.Select(ctx, &result, query, params...)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return nil, err
	}

	// process response.
	configs := make([]*GovConfig, 0, len(result))
	for _, r := range result {
		configs = append(configs, createGovConfig(r))
	}
	return configs, nil
}

func buildGovConfigQuery(q *GovernorQuery) (string, []any) {
	baseQuery := `SELECT id, guardian_name, counter, timestamp, chains, tokens, created_at, updated_at FROM wormholescan.wh_governor_config`
	var params []any
	var counter uint8 = 1

	// handle filtering by id (guardian address).
	if q.id != nil {
		baseQuery += fmt.Sprintf(" WHERE id = $%d", counter)
		params = append(params, q.id)
		counter++
	}

	// handle pagination and sorting.
	baseQuery += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d OFFSET $%d", counter, counter+1)
	params = append(params, q.Limit, q.Skip)

	return baseQuery, params
}

func createGovConfig(r governorConfigResult) *GovConfig {
	chains := make([]*GovConfigChains, 0, len(r.Chains))
	for _, c := range r.Chains {
		chains = append(chains, &GovConfigChains{
			ChainID:            c.ChainID,
			NotionalLimit:      mongo.Uint64(c.NotionalLimit),
			BigTransactionSize: mongo.Uint64(c.BigTransactionSize),
		})
	}

	tokens := make([]*GovConfigfTokens, 0, len(r.Tokens))
	for _, t := range r.Tokens {
		tokens = append(tokens, &GovConfigfTokens{
			OriginChainID: t.OriginChainID,
			OriginAddress: t.OriginAddress,
			Price:         t.Price,
		})
	}

	return &GovConfig{
		ID:        r.ID,
		NodeName:  r.GuardianName,
		Counter:   int(r.Counter),
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		Chains:    chains,
		Tokens:    tokens,
	}
}

func mapGovConfigResults(result []governorConfigResult) []*GovConfig {
	configs := make([]*GovConfig, 0, len(result))
	for _, r := range result {
		configs = append(configs, createGovConfig(r))
	}
	return configs
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