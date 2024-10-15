package operations

import (
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

func Test_buildQueryIDsForAddress(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())

	pagination := pagination.Default()

	expectedQuery := `
        SELECT oa.id FROM wormholescan.wh_operation_addresses oa
        WHERE oa.address = $1 AND exists (
            SELECT ot.attestation_vaas_id FROM wormholescan.wh_operation_transactions ot
            WHERE ot.attestation_vaas_id = oa.id 
        ) 
        ORDER BY oa."timestamp" DESC, oa.id DESC
        LIMIT $2 OFFSET $3`

	query, params := repo.buildQueryIDsForAddress("0x1234567890123456789012345678901234567890", nil, nil, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 3, len(params))
	assert.Equal(t, "0x1234567890123456789012345678901234567890", params[0])
	assert.Equal(t, int64(50), params[1])
	assert.Equal(t, int64(0), params[2])
}

func Test_buildQueryIDsForAddressWithFromAndTo(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())

	pagination := pagination.Default()

	expectedQuery := `
        SELECT oa.id FROM wormholescan.wh_operation_addresses oa
        WHERE oa.address = $1 AND exists (
            SELECT ot.attestation_vaas_id FROM wormholescan.wh_operation_transactions ot
            WHERE ot.attestation_vaas_id = oa.id 
        )  AND oa."timestamp" >= $4 AND oa."timestamp" <= $5
        ORDER BY oa."timestamp" DESC, oa.id DESC
        LIMIT $2 OFFSET $3`

	from := time.Date(2021, 2, 19, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 7, 6, 0, 36, 0, 0, time.UTC)
	query, params := repo.buildQueryIDsForAddress("0x1234567890123456789012345678901234567890", &from, &to, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 5, len(params))
	assert.Equal(t, "0x1234567890123456789012345678901234567890", params[0])
	assert.Equal(t, int64(50), params[1])
	assert.Equal(t, int64(0), params[2])
	assert.Equal(t, from, params[3])
	assert.Equal(t, to, params[4])
}

func Test_buildQueryIDsForAddressWithTo(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())

	pagination := pagination.Default()

	expectedQuery := `
        SELECT oa.id FROM wormholescan.wh_operation_addresses oa
        WHERE oa.address = $1 AND exists (
            SELECT ot.attestation_vaas_id FROM wormholescan.wh_operation_transactions ot
            WHERE ot.attestation_vaas_id = oa.id 
        )  AND oa."timestamp" <= $4
        ORDER BY oa."timestamp" DESC, oa.id DESC
        LIMIT $2 OFFSET $3`

	to := time.Date(2024, 7, 6, 0, 36, 0, 0, time.UTC)
	query, params := repo.buildQueryIDsForAddress("0x1234567890123456789012345678901234567890", nil, &to, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 4, len(params))
	assert.Equal(t, "0x1234567890123456789012345678901234567890", params[0])
	assert.Equal(t, int64(50), params[1])
	assert.Equal(t, int64(0), params[2])
	assert.Equal(t, to, params[3])
}

func Test_buildQueryIDsForTxHash(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())

	pagination := pagination.Default()

	expectedQuery := `
        SELECT t.attestation_vaas_id FROM wormholescan.wh_operation_transactions t
        WHERE t.tx_hash = $1 
        ORDER BY t.timestamp DESC, t.attestation_vaas_id DESC
        LIMIT $2 OFFSET $3`

	query, params := repo.buildQueryIDsForTxHash("0x87ebf4ae9a729855e491557270dc3e69da04092e6f6ca0025b2f88a2c1ea9be6", nil, nil, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 3, len(params))
	assert.Equal(t, "0x87ebf4ae9a729855e491557270dc3e69da04092e6f6ca0025b2f88a2c1ea9be6", params[0])
	assert.Equal(t, int64(50), params[1])
	assert.Equal(t, int64(0), params[2])
}

func Test_buildQueryIDsForTxHashWithFromAndTo(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())

	pagination := pagination.Default()
	from := time.Date(2021, 2, 19, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 7, 6, 0, 36, 0, 0, time.UTC)

	expectedQuery := `
        SELECT t.attestation_vaas_id FROM wormholescan.wh_operation_transactions t
        WHERE t.tx_hash = $1  AND t."timestamp" >= $4 AND t."timestamp" <= $5
        ORDER BY t.timestamp DESC, t.attestation_vaas_id DESC
        LIMIT $2 OFFSET $3`

	query, params := repo.buildQueryIDsForTxHash("0x87ebf4ae9a729855e491557270dc3e69da04092e6f6ca0025b2f88a2c1ea9be6", &from, &to, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 5, len(params))
	assert.Equal(t, "0x87ebf4ae9a729855e491557270dc3e69da04092e6f6ca0025b2f88a2c1ea9be6", params[0])
	assert.Equal(t, int64(50), params[1])
	assert.Equal(t, int64(0), params[2])
	assert.Equal(t, from, params[3])
	assert.Equal(t, to, params[4])
}

func Test_buildQueryIDsForQuery(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())
	pagination := pagination.Default()

	expectedQuery := `
            SELECT op.attestation_vaas_id FROM wormholescan.wh_operation_transactions op
            WHERE op."type" = 'source-tx' 
            ORDER BY op."timestamp" DESC, op.attestation_vaas_id DESC
            LIMIT $1 OFFSET $2`

	query, params := repo.buildQueryIDsForQuery(OperationQuery{}, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 2, len(params))
	assert.Equal(t, int64(50), params[0])
	assert.Equal(t, int64(0), params[1])
}

func Test_buildQueryIDsForQueryWithFromAndTo(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())
	pagination := pagination.Default()
	from := time.Date(2021, 2, 19, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 7, 6, 0, 36, 0, 0, time.UTC)
	q := OperationQuery{From: &from, To: &to}

	expectedQuery := `
            SELECT op.attestation_vaas_id FROM wormholescan.wh_operation_transactions op
            WHERE op."type" = 'source-tx'  AND op."timestamp" >= $3 AND op."timestamp" <= $4
            ORDER BY op."timestamp" DESC, op.attestation_vaas_id DESC
            LIMIT $1 OFFSET $2`

	query, params := repo.buildQueryIDsForQuery(q, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 4, len(params))
	assert.Equal(t, int64(50), params[0])
	assert.Equal(t, int64(0), params[1])
	assert.Equal(t, from, params[2])
	assert.Equal(t, to, params[3])
}

func Test_buildQueryIDsForQueryWithFrom(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())
	pagination := pagination.Default()
	from := time.Date(2021, 2, 19, 0, 0, 0, 0, time.UTC)
	q := OperationQuery{From: &from}

	expectedQuery := `
            SELECT op.attestation_vaas_id FROM wormholescan.wh_operation_transactions op
            WHERE op."type" = 'source-tx'  AND op."timestamp" >= $3
            ORDER BY op."timestamp" DESC, op.attestation_vaas_id DESC
            LIMIT $1 OFFSET $2`

	query, params := repo.buildQueryIDsForQuery(q, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 3, len(params))
	assert.Equal(t, int64(50), params[0])
	assert.Equal(t, int64(0), params[1])
	assert.Equal(t, from, params[2])
}

func Test_buildQueryIDsForQueryWithTargetChains(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())
	pagination := pagination.Default()
	q := OperationQuery{TargetChainIDs: []sdk.ChainID{sdk.ChainIDSolana, sdk.ChainIDEthereum}}

	expectedQuery := `
        SELECT p.id FROM wormholescan.wh_attestation_vaa_properties p
        WHERE p.to_chain_id = ANY($1)
        ORDER BY p.timestamp DESC, p.id DESC
        LIMIT $2 OFFSET $3`

	query, params := repo.buildQueryIDsForQuery(q, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 3, len(params))
	assert.Equal(t, pq.Array(q.TargetChainIDs), params[0])
	assert.Equal(t, int64(50), params[1])
	assert.Equal(t, int64(0), params[2])
}

func Test_buildQueryIDsForQueryWithTargetChainsAndFrom(t *testing.T) {

	repo := NewPostgresRepository(nil, zap.NewNop())
	pagination := pagination.Default()
	from := time.Date(2021, 2, 19, 0, 0, 0, 0, time.UTC)
	q := OperationQuery{From: &from, SourceChainIDs: []sdk.ChainID{sdk.ChainIDAptos}}

	expectedQuery := `
        SELECT p.id FROM wormholescan.wh_attestation_vaa_properties p
        WHERE p.from_chain_id = ANY($1) AND p."timestamp" >= $2
        ORDER BY p.timestamp DESC, p.id DESC
        LIMIT $3 OFFSET $4`

	query, params := repo.buildQueryIDsForQuery(q, *pagination)
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, 4, len(params))
	assert.Equal(t, pq.Array(q.SourceChainIDs), params[0])
	assert.Equal(t, from, params[1])
	assert.Equal(t, int64(50), params[2])
	assert.Equal(t, int64(0), params[3])
}
