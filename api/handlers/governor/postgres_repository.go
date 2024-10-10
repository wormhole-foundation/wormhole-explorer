package governor

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
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

func (r *PostgresRepository) GetGovernorNotionalLimit(ctx context.Context, queryFilter *NotionalLimitQuery) ([]*NotionalLimit, error) {

	limit := queryFilter.Pagination.Limit
	offset := queryFilter.Pagination.Skip

	query := `
		WITH RankedChains AS (SELECT (chain_data.value ->> 'chainid')::SMALLINT     AS chainId,
                             chain_data.value ->> 'notionallimit'       AS notionalLimit,
                             chain_data.value ->> 'bigtransactionsize'  AS maxTransactionSize,
                             ROW_NUMBER() OVER (PARTITION BY chain_data.value ->> 'chainid' ORDER BY chain_data.value ->> 'notionallimit' DESC, chain_data.value ->> 'bigtransactionsize' DESC) AS rowNum
                      FROM wormholescan.wh_governor_config,
                           jsonb_array_elements(chains) AS chain_data)
		SELECT chainId,
		       notionalLimit,
		       maxTransactionSize
		FROM RankedChains
		WHERE rowNum = 13
		ORDER BY chainId ASC
		LIMIT $1 OFFSET $2;
	`

	var result []*NotionalLimit
	var response []notionalLimitSQL
	err := r.db.Select(ctx, &response, query, limit, offset)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return result, err
	}

	for _, a := range response {

		var notionalLimit float64
		var maxTxSize float64

		notionalLimit, err = strconv.ParseFloat(a.NotionalLimit, 10)
		if err != nil {
			r.logger.Error("failed to parse notional limit", zap.Error(err), zap.String("notional_limit", a.NotionalLimit))
			break
		}

		maxTxSize, err = strconv.ParseFloat(a.MaxTransactionSize, 10)
		if err != nil {
			r.logger.Error("failed to parse max transaction size", zap.Error(err), zap.String("max_transaction_size", a.MaxTransactionSize))
			break
		}

		result = append(result, &NotionalLimit{
			ChainID:            a.ChainID,
			NotionalLimit:      uint64(notionalLimit),
			MaxTransactionSize: uint64(maxTxSize),
		})
	}

	return result, err

}

func (r *PostgresRepository) GetNotionalLimitByChainID(
	ctx context.Context,
	q *NotionalLimitQuery,
) ([]*NotionalLimitDetail, error) {
	limit := q.Pagination.Limit
	offset := q.Pagination.Skip

	query := `
		SELECT 	wormholescan.wh_governor_config.id,
		    	wormholescan.wh_governor_config.guardian_name,
       		  	wormholescan.wh_governor_config.created_at,
       			wormholescan.wh_governor_config.updated_at,
       			(chain_data.value ->> 'chainid')::SMALLINT AS chainId,
       			chain_data.value ->> 'notionallimit'       AS notionalLimit,
      			chain_data.value ->> 'bigtransactionsize'  AS maxTransactionSize
		FROM	wormholescan.wh_governor_config,
     			jsonb_array_elements(chains) AS chain_data
		WHERE chain_data.value ->> 'chainid' = $1
		LIMIT $2 OFFSET $3;
	`

	var result []*NotionalLimitDetail
	var response []notionalLimitDetailSQL

	itoa := strconv.Itoa(int(q.chainID))
	err := r.db.Select(ctx, &response, query, itoa, limit, offset)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return result, err
	}

	for _, nl := range response {

		var notionalLimit float64
		var maxTxSize float64

		notionalLimit, err = strconv.ParseFloat(nl.NotionalLimit, 10)
		if err != nil {
			r.logger.Error("failed to parse notional limit", zap.Error(err), zap.String("notional_limit", nl.NotionalLimit))
			break
		}

		maxTxSize, err = strconv.ParseFloat(nl.MaxTransactionSize, 10)
		if err != nil {
			r.logger.Error("failed to parse max transaction size", zap.Error(err), zap.String("max_transaction_size", nl.MaxTransactionSize))
			break
		}

		result = append(result, &NotionalLimitDetail{
			ID:                 nl.ID,
			ChainID:            nl.ChainID,
			NodeName:           nl.NodeName,
			CreatedAt:          nl.CreatedAt,
			UpdatedAt:          nl.UpdatedAt,
			NotionalLimit:      uint64(notionalLimit),
			MaxTransactionSize: uint64(maxTxSize),
		})
	}

	return result, err
}

func (r *PostgresRepository) GetAvailableNotional(
	ctx context.Context,
	q *NotionalLimitQuery,
) ([]*NotionalAvailable, error) {

	limit := q.Pagination.Limit
	offset := q.Pagination.Skip

	query := `
	WITH RankedChains AS (SELECT (chain_data.value ->> 'chainid')::SMALLINT     AS chainId,
                             chain_data.value ->> 'remainingavailablenotional'  AS remainingavailablenotional,
                             ROW_NUMBER()
                             OVER (PARTITION BY chain_data.value ->> 'chainid' ORDER BY chain_data.value ->> 'remainingavailablenotional' DESC) AS rowNum
                      FROM wormholescan.wh_governor_status,
                           jsonb_array_elements(wormholescan.wh_governor_status.message) AS chain_data)
	SELECT chainId,
	       remainingavailablenotional as availableNotional
	FROM RankedChains
	WHERE rowNum = 13
	ORDER BY chainId
	LIMIT $1 OFFSET $2;
	`

	var result []*NotionalAvailable
	var response []notionalAvailableSQL

	err := r.db.Select(ctx, &response, query, limit, offset)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return result, err
	}

	for _, nl := range response {

		var notionalAvailable float64

		notionalAvailable, err = strconv.ParseFloat(nl.AvailableNotional, 10)
		if err != nil {
			r.logger.Error("failed to parse notional limit", zap.Error(err), zap.String("notional_limit", nl.AvailableNotional))
			break
		}

		result = append(result, &NotionalAvailable{
			ChainID:           nl.ChainID,
			AvailableNotional: uint64(notionalAvailable),
		})
	}

	return result, err
}

func (r *PostgresRepository) GetAvailableNotionalByChainID(
	ctx context.Context,
	q *NotionalLimitQuery,
) ([]*NotionalAvailableDetail, error) {

	limit := q.Pagination.Limit
	offset := q.Pagination.Skip

	query := `
	SELECT 	wormholescan.wh_governor_status.id,
	        wormholescan.wh_governor_status.guardian_name,
	       	wormholescan.wh_governor_status.created_at,
	       	wormholescan.wh_governor_status.updated_at,
	       	(message.value ->> 'chainid')::SMALLINT AS chainId,
	       	message.value ->> 'remainingavailablenotional' AS availableNotional
	FROM    wormholescan.wh_governor_status,
	     	jsonb_array_elements(wormholescan.wh_governor_status.message) AS message
	WHERE message.value ->> 'chainid' = $1
	ORDER BY wormholescan.wh_governor_status.id DESC
	LIMIT $2 OFFSET $3;
	`

	var result []*NotionalAvailableDetail
	var response []notionalAvailableDetailSQL

	itoa := strconv.Itoa(int(q.chainID))
	err := r.db.Select(ctx, &response, query, itoa, limit, offset)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return result, err
	}

	for _, nl := range response {

		var notionalAvailable float64

		notionalAvailable, err = strconv.ParseFloat(nl.NotionalAvailable, 10)
		if err != nil {
			r.logger.Error("failed to parse notional limit", zap.Error(err), zap.String("notional_limit", nl.NotionalAvailable))
			break
		}

		result = append(result, &NotionalAvailableDetail{
			ID:                nl.ID,
			ChainID:           nl.ChainID,
			NodeName:          nl.NodeName,
			CreatedAt:         nl.CreatedAt,
			UpdatedAt:         nl.UpdatedAt,
			NotionalAvailable: uint64(notionalAvailable),
		})
	}

	return result, err
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

func (r *PostgresRepository) GetEnqueueVass(ctx context.Context, _ *EnqueuedVaaQuery) ([]*EnqueuedVaas, error) {
	query := `
		WITH flattened AS (	SELECT 	(chain ->> 'chainid')::int 		AS chain_id,
									jsonb_array_elements(chain -> 'emitters') AS emitter
                   			FROM 	wormholescan.wh_governor_status,
                        			jsonb_array_elements(message) AS chain
							),
     		deconstructedChains as (SELECT 	chain_id,
                                    		emitter ->> 'emitteraddress'					AS emitter_address,
                                    		jsonb_array_elements(flattened.emitter -> 'enqueuedvaas')	AS vaa
                             		FROM 	flattened
                             		WHERE 	flattened.emitter -> 'enqueuedvaas' IS NOT NULL
                             		AND 	(flattened.emitter -> 'enqueuedvaas' != 'null'))
		SELECT chain_id,
		       emitter_address,
		       (vaa ->> 'sequence')               AS sequence,
		       (vaa ->> 'releasetime')::bigint    AS release_time,
		       (vaa ->> 'notionalvalue')::numeric AS notional_value,
		       vaa ->> 'txhash'                   AS tx_hash
		FROM deconstructedChains
    `

	var items []struct {
		ChainID        vaa.ChainID     `db:"chain_id"`
		EmitterAddress string          `db:"emitter_address"`
		Sequence       string          `db:"sequence"`
		ReleaseTime    int64           `db:"release_time"`
		NotionalValue  decimal.Decimal `db:"notional_value"`
		TxHash         string          `db:"tx_hash"`
	}

	err := r.db.Select(ctx, &items, query)
	if err != nil {
		r.logger.Error("failed to execute query to get enqueued VAAs",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	// Group the results by chain ID
	enqueuedVaasGroupedByChainID := make(map[vaa.ChainID][]*EnqueuedVaa)
	for _, item := range items {
		detail := &EnqueuedVaa{
			ChainID:        item.ChainID,
			EmitterAddress: item.EmitterAddress,
			Sequence:       item.Sequence,
			NotionalValue:  item.NotionalValue.IntPart(),
			TxHash:         item.TxHash,
		}
		enqueuedVaasGroupedByChainID[item.ChainID] = append(enqueuedVaasGroupedByChainID[item.ChainID], detail)
	}

	// Create the response
	response := make([]*EnqueuedVaas, 0, len(enqueuedVaasGroupedByChainID))
	for chainID, vaas := range enqueuedVaasGroupedByChainID {
		response = append(response, &EnqueuedVaas{
			ChainID:     chainID,
			EnqueuedVaa: vaas,
		})
	}

	// Sort the response by chain ID
	sort.Slice(response, func(i, j int) bool {
		return response[i].ChainID < response[j].ChainID
	})

	return response, nil
}

func (r *PostgresRepository) GetEnqueueVassByChainID(ctx context.Context, q *EnqueuedVaaQuery) ([]*EnqueuedVaaDetail, error) {
	query := `
		WITH flattened AS (	SELECT 	(chain ->> 'chainid')::int 		AS chain_id,
									jsonb_array_elements(chain -> 'emitters') AS emitter
                   			FROM 	wormholescan.wh_governor_status,
                        			jsonb_array_elements(message) AS chain
							WHERE 	(chain ->> 'chainid')::int = $1	
							),
     		deconstructedChains as (SELECT 	chain_id,
                                    		emitter ->> 'emitteraddress'					AS emitter_address,
                                    		jsonb_array_elements(flattened.emitter -> 'enqueuedvaas')	AS vaa
                             		FROM 	flattened
                             		WHERE 	flattened.emitter -> 'enqueuedvaas' IS NOT NULL
                             		AND 	(flattened.emitter -> 'enqueuedvaas' != 'null'))
		SELECT chain_id,
		       emitter_address,
		       (vaa ->> 'sequence')               AS sequence,
		       (vaa ->> 'releasetime')::bigint    AS release_time,
		       (vaa ->> 'notionalvalue')::numeric AS notional_value,
		       vaa ->> 'txhash'                   AS tx_hash
		FROM deconstructedChains
    `

	var items []struct {
		ChainID        vaa.ChainID     `db:"chain_id"`
		EmitterAddress string          `db:"emitter_address"`
		Sequence       string          `db:"sequence"`
		ReleaseTime    int64           `db:"release_time"`
		NotionalValue  decimal.Decimal `db:"notional_value"`
		TxHash         string          `db:"tx_hash"`
	}

	err := r.db.Select(ctx, &items, query, q.chainID)
	if err != nil {
		r.logger.Error("failed to execute query to get enqueued VAAs",
			zap.Error(err),
			zap.String("query", query))
		return nil, err
	}

	// Create the response
	response := make([]*EnqueuedVaaDetail, 0, len(items))
	for _, item := range items {
		detail := &EnqueuedVaaDetail{
			ChainID:        item.ChainID,
			EmitterAddress: item.EmitterAddress,
			Sequence:       item.Sequence,
			NotionalValue:  item.NotionalValue.IntPart(),
			TxHash:         item.TxHash,
			ReleaseTime:    item.ReleaseTime,
		}
		response = append(response, detail)
	}

	// Sort the response by sequence
	sort.Slice(response, func(i, j int) bool {
		seqI, _ := strconv.ParseUint(response[i].Sequence, 10, 64)
		seqJ, _ := strconv.ParseUint(response[j].Sequence, 10, 64)
		return seqI < seqJ
	})

	return response, nil
}
