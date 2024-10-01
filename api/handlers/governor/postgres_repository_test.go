package governor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/mongo"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func TestPostgresRepository_createGovernorLimit(t *testing.T) {

	chainLimits := []governorLimitResult{
		{NotionalLimit: 100, BigTransactionSize: 5800, AvailableNotional: 9800},
		{NotionalLimit: 200, BigTransactionSize: 5900, AvailableNotional: 10000},
		{NotionalLimit: 300, BigTransactionSize: 6000, AvailableNotional: 11000},
		{NotionalLimit: 400, BigTransactionSize: 6100, AvailableNotional: 9000},
		{NotionalLimit: 500, BigTransactionSize: 7100, AvailableNotional: 8000},
		{NotionalLimit: 600, BigTransactionSize: 8100, AvailableNotional: 7000},
		{NotionalLimit: 700, BigTransactionSize: 4100, AvailableNotional: 7000},
		{NotionalLimit: 100, BigTransactionSize: 9000, AvailableNotional: 6000},
		{NotionalLimit: 200, BigTransactionSize: 900, AvailableNotional: 5000},
		{NotionalLimit: 300, BigTransactionSize: 800, AvailableNotional: 4000},
		{NotionalLimit: 400, BigTransactionSize: 7500, AvailableNotional: 6000},
		{NotionalLimit: 500, BigTransactionSize: 4000, AvailableNotional: 9000},
		{NotionalLimit: 600, BigTransactionSize: 6000, AvailableNotional: 4000},
		{NotionalLimit: 500, BigTransactionSize: 1000, AvailableNotional: 3000},
		{NotionalLimit: 600, BigTransactionSize: 2000, AvailableNotional: 1700},
	}

	governorLimit := createGovernorLimit(1, chainLimits)

	assert.Equal(t, governorLimit.ChainID, sdk.ChainIDSolana)
	assert.Equal(t, governorLimit.NotionalLimit, mongo.Uint64(200))
	assert.Equal(t, governorLimit.MaxTransactionSize, mongo.Uint64(1000))
	assert.Equal(t, governorLimit.AvailableNotional, mongo.Uint64(4000))
}

func TestPostgresRepository_createGovernorLimitEmpty(t *testing.T) {

	chainLimits := []governorLimitResult{
		{NotionalLimit: 100, BigTransactionSize: 5800, AvailableNotional: 9800},
		{NotionalLimit: 200, BigTransactionSize: 5900, AvailableNotional: 10000},
		{NotionalLimit: 300, BigTransactionSize: 6000, AvailableNotional: 11000},
	}
	governorLimit := createGovernorLimit(2, chainLimits)

	assert.Equal(t, governorLimit.ChainID, sdk.ChainIDEthereum)
	assert.Equal(t, governorLimit.NotionalLimit, mongo.Uint64(0))
	assert.Equal(t, governorLimit.MaxTransactionSize, mongo.Uint64(0))
	assert.Equal(t, governorLimit.AvailableNotional, mongo.Uint64(0))
}

func TestPostgresRepository_paginateEmpty(t *testing.T) {

	limits := make([]*GovernorLimit, 0)
	paginated := paginate(limits, 0, 10)

	assert.Equal(t, len(paginated), 0)
}

func TestPostgresRepository_paginate(t *testing.T) {

	limits := []*GovernorLimit{
		{ChainID: sdk.ChainID(1), NotionalLimit: mongo.Uint64(100), MaxTransactionSize: mongo.Uint64(1000), AvailableNotional: mongo.Uint64(4000)},
		{ChainID: sdk.ChainID(2), NotionalLimit: mongo.Uint64(200), MaxTransactionSize: mongo.Uint64(2000), AvailableNotional: mongo.Uint64(5000)},
		{ChainID: sdk.ChainID(3), NotionalLimit: mongo.Uint64(300), MaxTransactionSize: mongo.Uint64(3000), AvailableNotional: mongo.Uint64(6000)},
		{ChainID: sdk.ChainID(4), NotionalLimit: mongo.Uint64(400), MaxTransactionSize: mongo.Uint64(4000), AvailableNotional: mongo.Uint64(7000)},
		{ChainID: sdk.ChainID(5), NotionalLimit: mongo.Uint64(500), MaxTransactionSize: mongo.Uint64(5000), AvailableNotional: mongo.Uint64(8000)},
		{ChainID: sdk.ChainID(6), NotionalLimit: mongo.Uint64(600), MaxTransactionSize: mongo.Uint64(6000), AvailableNotional: mongo.Uint64(9000)},
		{ChainID: sdk.ChainID(7), NotionalLimit: mongo.Uint64(700), MaxTransactionSize: mongo.Uint64(7000), AvailableNotional: mongo.Uint64(10000)},
		{ChainID: sdk.ChainID(8), NotionalLimit: mongo.Uint64(800), MaxTransactionSize: mongo.Uint64(8000), AvailableNotional: mongo.Uint64(11000)},
		{ChainID: sdk.ChainID(9), NotionalLimit: mongo.Uint64(900), MaxTransactionSize: mongo.Uint64(9000), AvailableNotional: mongo.Uint64(12000)},
		{ChainID: sdk.ChainID(10), NotionalLimit: mongo.Uint64(1000), MaxTransactionSize: mongo.Uint64(10000), AvailableNotional: mongo.Uint64(13000)},
	}
	paginated := paginate(limits, 2, 5)

	assert.Equal(t, len(paginated), 5)
	assert.Equal(t, paginated[0].ChainID, sdk.ChainID(3))
}

func TestPostgresRepository_paginateBigSize(t *testing.T) {

	limits := []*GovernorLimit{
		{ChainID: sdk.ChainID(1), NotionalLimit: mongo.Uint64(100), MaxTransactionSize: mongo.Uint64(1000), AvailableNotional: mongo.Uint64(4000)},
		{ChainID: sdk.ChainID(2), NotionalLimit: mongo.Uint64(200), MaxTransactionSize: mongo.Uint64(2000), AvailableNotional: mongo.Uint64(5000)},
		{ChainID: sdk.ChainID(3), NotionalLimit: mongo.Uint64(300), MaxTransactionSize: mongo.Uint64(3000), AvailableNotional: mongo.Uint64(6000)},
		{ChainID: sdk.ChainID(4), NotionalLimit: mongo.Uint64(400), MaxTransactionSize: mongo.Uint64(4000), AvailableNotional: mongo.Uint64(7000)},
		{ChainID: sdk.ChainID(5), NotionalLimit: mongo.Uint64(500), MaxTransactionSize: mongo.Uint64(5000), AvailableNotional: mongo.Uint64(8000)},
		{ChainID: sdk.ChainID(6), NotionalLimit: mongo.Uint64(600), MaxTransactionSize: mongo.Uint64(6000), AvailableNotional: mongo.Uint64(9000)},
		{ChainID: sdk.ChainID(7), NotionalLimit: mongo.Uint64(700), MaxTransactionSize: mongo.Uint64(7000), AvailableNotional: mongo.Uint64(10000)},
		{ChainID: sdk.ChainID(8), NotionalLimit: mongo.Uint64(800), MaxTransactionSize: mongo.Uint64(8000), AvailableNotional: mongo.Uint64(11000)},
		{ChainID: sdk.ChainID(9), NotionalLimit: mongo.Uint64(900), MaxTransactionSize: mongo.Uint64(9000), AvailableNotional: mongo.Uint64(12000)},
		{ChainID: sdk.ChainID(10), NotionalLimit: mongo.Uint64(1000), MaxTransactionSize: mongo.Uint64(10000), AvailableNotional: mongo.Uint64(13000)},
	}
	paginated := paginate(limits, 3, 50)

	assert.Equal(t, len(paginated), 7)
	assert.Equal(t, paginated[0].ChainID, sdk.ChainID(4))
}

func TestGovernorQuery_toQuery(t *testing.T) {
	q := NewGovernorQuery()
	expected := "SELECT id, guardian_name, message , created_at , updated_at FROM wormholescan.wh_governor_status \n ORDER BY id ASC \n LIMIT $1 OFFSET $2 \n "
	query, params := q.toQuery()
	assert.Equal(t, query, expected)
	assert.Equal(t, len(params), 2)
	assert.Equal(t, params[0], q.Limit)
	assert.Equal(t, params[1], q.Skip)
}

func TestGovernorQuery_toQueryWithID(t *testing.T) {
	addr := "000ac0076727b35fbea2dac28fee5ccb0fea768e"
	address, err := types.StringToAddress(addr, true)
	assert.Nil(t, err)
	assert.NotNil(t, address)
	q := NewGovernorQuery().SetID(address)
	expected := "SELECT id, guardian_name, message , created_at , updated_at FROM wormholescan.wh_governor_status \n WHERE id = $1 \n "
	query, params := q.toQuery()
	assert.Equal(t, query, expected)
	assert.Equal(t, len(params), 1)
	assert.Equal(t, params[0], addr)
}
