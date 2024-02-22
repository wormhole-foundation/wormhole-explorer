package protocols_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/mock"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/protocols"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	cacheMock "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/mock"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestService_GetProtocolsTotalValues(t *testing.T) {
	var errNil error
	respStatsLatest := &mockQueryTableResult{}
	respStatsLatest.On("Next").Return(true)
	respStatsLatest.On("Err").Return(errNil)
	respStatsLatest.On("Close").Return(errNil)
	respStatsLatest.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           "protocol1",
		"total_messages":     uint64(7),
		"total_value_locked": float64(5),
	}))

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           "protocol1",
		"total_messages":     uint64(4),
		"total_value_locked": float64(5),
	}))

	respActivityLast := &mockQueryTableResult{}
	respActivityLast.On("Next").Return(true)
	respActivityLast.On("Err").Return(errNil)
	respActivityLast.On("Close").Return(errNil)
	respActivityLast.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":                "protocol1",
		"total_messages":          uint64(4),
		"total_value_transferred": float64(7),
		"total_value_secure":      float64(9),
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateLatestPoint, "protocols_bucket", dbconsts.ProtocolsStatsMeasurement, "protocol1", "v1")).Return(respStatsLatest, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateLast24Point, "protocols_bucket", dbconsts.ProtocolsStatsMeasurement, "protocol1", "v1")).Return(respStatsLastDay, nil)

	activityQuery := fmt.Sprintf(protocols.QueryTemplateActivityLatestPoint, "protocols_bucket", dbconsts.ProtocolsActivityMeasurement, "protocol1", "v1")
	queryAPI.On("Query", ctx, activityQuery).Return(respActivityLast, nil)

	repository := protocols.NewRepository(queryAPI, "protocols_bucket", "protocols_bucket", "v1", "v1", zap.NewNop())
	service := protocols.NewService([]string{"protocol1"}, repository, zap.NewNop(), cache.NewDummyCacheClient(), "WORMSCAN:PROTOCOLS", 0)

	values := service.GetProtocolsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "protocol1", values[0].Protocol)
	assert.Equal(t, 5.00, values[0].TotalValueLocked)
	assert.Equal(t, uint64(7), values[0].TotalMessages)
	assert.Equal(t, 9.00, values[0].TotalValueSecured)
	assert.Equal(t, 7.00, values[0].TotalValueTransferred)
	assert.Equal(t, uint64(3), values[0].LastDayMessages)
	assert.Equal(t, "75.00%", values[0].LastDayDiffPercentage)

}

func TestService_GetProtocolsTotalValues_FailedFetchingActivity(t *testing.T) {
	var errNil error
	respStatsLatest := &mockQueryTableResult{}
	respStatsLatest.On("Next").Return(true)
	respStatsLatest.On("Err").Return(errNil)
	respStatsLatest.On("Close").Return(errNil)
	respStatsLatest.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           "protocol1",
		"total_messages":     uint64(7),
		"total_value_locked": float64(5),
	}))

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           "protocol1",
		"total_messages":     uint64(4),
		"total_value_locked": float64(5),
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateLatestPoint, "protocols_bucket", dbconsts.ProtocolsStatsMeasurement, "protocol1", "v1")).Return(respStatsLatest, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateLast24Point, "protocols_bucket", dbconsts.ProtocolsStatsMeasurement, "protocol1", "v1")).Return(respStatsLastDay, nil)

	activityQuery := fmt.Sprintf(protocols.QueryTemplateActivityLatestPoint, "protocols_bucket", dbconsts.ProtocolsActivityMeasurement, "protocol1", "v1")
	queryAPI.On("Query", ctx, activityQuery).Return(&api.QueryTableResult{}, errors.New("mocked_fetching_activity_error"))

	repository := protocols.NewRepository(queryAPI, "protocols_bucket", "protocols_bucket", "v1", "v1", zap.NewNop())
	service := protocols.NewService([]string{"protocol1"}, repository, zap.NewNop(), cache.NewDummyCacheClient(), "WORMSCAN:PROTOCOLS", 0)

	values := service.GetProtocolsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "protocol1", values[0].Protocol)
	assert.NotNil(t, values[0].Error)
	assert.Equal(t, "mocked_fetching_activity_error", values[0].Error)
}

func TestService_GetProtocolsTotalValues_FailedFetchingStats(t *testing.T) {
	var errNil error

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           "protocol1",
		"total_messages":     uint64(4),
		"total_value_locked": float64(5),
	}))

	respActivityLast := &mockQueryTableResult{}
	respActivityLast.On("Next").Return(true)
	respActivityLast.On("Err").Return(errNil)
	respActivityLast.On("Close").Return(errNil)
	respActivityLast.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":                "protocol1",
		"total_messages":          uint64(4),
		"total_value_transferred": float64(7),
		"total_volume_secure":     float64(9),
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateLatestPoint, "protocols_bucket", dbconsts.ProtocolsStatsMeasurement, "protocol1", "v1")).Return(&api.QueryTableResult{}, errors.New("mocked_fetching_stats_error"))
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateLast24Point, "protocols_bucket", dbconsts.ProtocolsStatsMeasurement, "protocol1", "v1")).Return(respStatsLastDay, nil)

	activityQuery := fmt.Sprintf(protocols.QueryTemplateActivityLatestPoint, "protocols_bucket", dbconsts.ProtocolsActivityMeasurement, "protocol1", "v1")
	queryAPI.On("Query", ctx, activityQuery).Return(respActivityLast, errNil)

	repository := protocols.NewRepository(queryAPI, "protocols_bucket", "protocols_bucket", "v1", "v1", zap.NewNop())
	service := protocols.NewService([]string{"protocol1"}, repository, zap.NewNop(), cache.NewDummyCacheClient(), "WORMSCAN:PROTOCOLS", 0)

	values := service.GetProtocolsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "protocol1", values[0].Protocol)
	assert.NotNil(t, values[0].Error)
	assert.Equal(t, "mocked_fetching_stats_error", values[0].Error)
}

func TestService_GetProtocolsTotalValues_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockCache := &cacheMock.CacheMock{}
	var cacheErr error
	cacheErr = nil
	mockCache.On("Get", ctx, "WORMSCAN:PROTOCOLS:protocol1").Return(`{"protocol":"protocol1","total_messages":7,"total_value_locked":5,"total_value_secured":9,"total_value_transferred":7,"last_day_messages":4,"last_day_diff_percentage":"75.00%"}`, cacheErr)
	service := protocols.NewService([]string{"protocol1"}, nil, zap.NewNop(), mockCache, "WORMSCAN:PROTOCOLS", 0)
	values := service.GetProtocolsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "protocol1", values[0].Protocol)
	assert.Equal(t, 5.00, values[0].TotalValueLocked)
	assert.Equal(t, uint64(7), values[0].TotalMessages)
	assert.Equal(t, 9.00, values[0].TotalValueSecured)
	assert.Equal(t, 7.00, values[0].TotalValueTransferred)
	assert.Equal(t, uint64(4), values[0].LastDayMessages)
	assert.Equal(t, "75.00%", values[0].LastDayDiffPercentage)

}

func TestService_GetProtocolsTotalValues_CacheMiss_FetchAndUpdate(t *testing.T) {

	ctx := context.Background()
	mockCache := &cacheMock.CacheMock{}
	mockCache.On("Get", ctx, "WORMSCAN:PROTOCOLS:protocol1").Return("", cache.ErrNotFound) // mock cache miss

	// mock cache update, validate it's called once.
	mockCache.On("Set",
		ctx,
		"WORMSCAN:PROTOCOLS:protocol1",
		`{"protocol":"protocol1","total_messages":7,"total_value_locked":5,"total_value_secured":9,"total_value_transferred":7,"last_day_messages":3,"last_day_diff_percentage":"75.00%"}`,
		time.Duration(60)*time.Minute).
		Return(nil).
		Times(1)

	var errNil error
	respStatsLatest := &mockQueryTableResult{}
	respStatsLatest.On("Next").Return(true)
	respStatsLatest.On("Err").Return(errNil)
	respStatsLatest.On("Close").Return(errNil)
	respStatsLatest.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           "protocol1",
		"total_messages":     uint64(7),
		"total_value_locked": float64(5),
	}))

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           "protocol1",
		"total_messages":     uint64(4),
		"total_value_locked": float64(5),
	}))

	respActivityLast := &mockQueryTableResult{}
	respActivityLast.On("Next").Return(true)
	respActivityLast.On("Err").Return(errNil)
	respActivityLast.On("Close").Return(errNil)
	respActivityLast.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":                "protocol1",
		"total_messages":          uint64(15),
		"total_value_transferred": float64(7),
		"total_value_secure":      float64(9),
	}))

	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateLatestPoint, "protocols_bucket", dbconsts.ProtocolsStatsMeasurement, "protocol1", "v1")).Return(respStatsLatest, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateLast24Point, "protocols_bucket", dbconsts.ProtocolsStatsMeasurement, "protocol1", "v1")).Return(respStatsLastDay, nil)

	activityQuery := fmt.Sprintf(protocols.QueryTemplateActivityLatestPoint, "protocols_bucket", dbconsts.ProtocolsActivityMeasurement, "protocol1", "v1")
	queryAPI.On("Query", ctx, activityQuery).Return(respActivityLast, nil)
	repository := protocols.NewRepository(queryAPI, "protocols_bucket", "protocols_bucket", "v1", "v1", zap.NewNop())
	service := protocols.NewService([]string{"protocol1"}, repository, zap.NewNop(), mockCache, "WORMSCAN:PROTOCOLS", 60)

	values := service.GetProtocolsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "protocol1", values[0].Protocol)
	assert.Equal(t, 5.00, values[0].TotalValueLocked)
	assert.Equal(t, uint64(7), values[0].TotalMessages)
	assert.Equal(t, 9.00, values[0].TotalValueSecured)
	assert.Equal(t, 7.00, values[0].TotalValueTransferred)
	assert.Equal(t, uint64(3), values[0].LastDayMessages)
	assert.Equal(t, "75.00%", values[0].LastDayDiffPercentage)
}

type mockQueryAPI struct {
	mock.Mock
}

func (m *mockQueryAPI) Query(ctx context.Context, q string) (protocols.QueryResult, error) {
	args := m.Called(ctx, q)
	return args.Get(0).(protocols.QueryResult), args.Error(1)
}

type mockQueryTableResult struct {
	mock.Mock
}

func (m *mockQueryTableResult) Next() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockQueryTableResult) Record() *query.FluxRecord {
	args := m.Called()
	return args.Get(0).(*query.FluxRecord)
}

func (m *mockQueryTableResult) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockQueryTableResult) Close() error {
	args := m.Called()
	return args.Error(0)
}
