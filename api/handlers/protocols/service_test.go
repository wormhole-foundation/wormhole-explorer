package protocols_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/mock"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/protocols"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	cacheMock "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/mock"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestService_GetProtocolsTotalValues_Allbridge(t *testing.T) {
	const allbridge = "allbridge"
	var errNil error
	respStatsLatest := &mockQueryTableResult{}
	respStatsLatest.On("Next").Return(true)
	respStatsLatest.On("Err").Return(errNil)
	respStatsLatest.On("Close").Return(errNil)
	respStatsLatest.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           allbridge,
		"total_messages":     uint64(7),
		"total_value_locked": float64(5),
	}))

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":       allbridge,
		"total_messages": uint64(4),
	}))

	respActivityLast := &mockQueryTableResult{}
	respActivityLast.On("Next").Return(true)
	respActivityLast.On("Err").Return(errNil)
	respActivityLast.On("Close").Return(errNil)
	ts := time.Now().UTC()
	respActivityLast.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":                allbridge,
		"total_messages":          uint64(4),
		"total_value_transferred": float64(7),
		"_time":                   ts,
	}))

	respActivity2 := &mockQueryTableResult{}
	respActivity2.On("Next").Return(true)
	respActivity2.On("Err").Return(errNil)
	respActivity2.On("Close").Return(errNil)
	respActivity2.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":                allbridge,
		"total_messages":          uint64(4),
		"total_value_transferred": float64(7),
	}))

	last24respActivity := &mockQueryTableResult{}
	last24respActivity.On("Next").Return(true)
	last24respActivity.On("Err").Return(errNil)
	last24respActivity.On("Close").Return(errNil)
	last24respActivity.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":                allbridge,
		"total_messages":          uint64(4),
		"total_value_transferred": float64(67),
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolStatsNow, "bucket30d", dbconsts.ProtocolsStatsMeasurementHourly, allbridge)).Return(respStatsLatest, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolStats24HrAgo, "bucket30d", dbconsts.ProtocolsStatsMeasurementHourly, allbridge)).Return(respStatsLastDay, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolActivity, "bucketInfinite", "1970-01-01T00:00:00Z", dbconsts.ProtocolsActivityMeasurementDaily, allbridge)).Return(respActivityLast, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolActivity, "bucket30d", ts.Format(time.RFC3339), dbconsts.ProtocolsActivityMeasurementHourly, allbridge)).Return(respActivity2, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryLast24HrActivity, "bucketInfinite", dbconsts.ProtocolsActivityMeasurementDaily, allbridge)).Return(last24respActivity, nil)

	// core protocols influx calls
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolStats24HrAgo, "bucketInfinite")).Return(emptyQueryTableResult(), nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaSinceStartOfDay, "bucket30d")).Return(emptyQueryTableResult(), nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaLastDay, "bucket30d")).Return(emptyQueryTableResult(), nil)

	repository := protocols.NewRepository(queryAPI, "bucketInfinite", "bucket30d", "bucket24hr", zap.NewNop())
	service := protocols.NewService([]string{protocols.ALLBRIDGE}, repository, zap.NewNop(), cache.NewDummyCacheClient(), "WORMSCAN:PROTOCOLS", 0, metrics.NewNoOpMetrics(), &mockTvl{})

	values := service.GetProtocolsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, allbridge, values[0].Protocol)
	assert.Equal(t, uint64(7), values[0].TotalMessages)
	assert.Equal(t, 14.00, values[0].TotalValueTransferred)
	assert.Equal(t, uint64(3), values[0].LastDayMessages)
	assert.Equal(t, "75.00%", values[0].LastDayDiffPercentage)
	assert.Equal(t, float64(67), values[0].Last24HourVolume)

}

func TestService_GetProtocolsTotalValues_Allbridge_FailedFetchingActivity(t *testing.T) {
	const allbridge = "allbridge"
	var errNil error
	respStatsLatest := &mockQueryTableResult{}
	respStatsLatest.On("Next").Return(true)
	respStatsLatest.On("Err").Return(errNil)
	respStatsLatest.On("Close").Return(errNil)
	respStatsLatest.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           allbridge,
		"total_messages":     uint64(7),
		"total_value_locked": float64(5),
	}))

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           allbridge,
		"total_messages":     uint64(4),
		"total_value_locked": float64(5),
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	// Allbridge influx calls
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolStatsNow, "bucket30d", dbconsts.ProtocolsStatsMeasurementHourly, allbridge)).Return(respStatsLatest, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolStats24HrAgo, "bucket30d", dbconsts.ProtocolsStatsMeasurementHourly, allbridge)).Return(respStatsLastDay, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolActivity, "bucketInfinite", "1970-01-01T00:00:00Z", dbconsts.ProtocolsActivityMeasurementDaily, allbridge)).Return(&mockQueryTableResult{}, errors.New("mocked_error"))

	// core protocols influx calls
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolStats24HrAgo, "bucketInfinite")).Return(emptyQueryTableResult(), nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaSinceStartOfDay, "bucket30d")).Return(emptyQueryTableResult(), nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaLastDay, "bucket30d")).Return(emptyQueryTableResult(), nil)

	repository := protocols.NewRepository(queryAPI, "bucketInfinite", "bucket30d", "bucket24hr", zap.NewNop())
	service := protocols.NewService([]string{protocols.ALLBRIDGE}, repository, zap.NewNop(), cache.NewDummyCacheClient(), "WORMSCAN:PROTOCOLS", 0, metrics.NewNoOpMetrics(), &mockTvl{})

	values := service.GetProtocolsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, allbridge, values[0].Protocol)
	assert.NotNil(t, values[0].Error)
	assert.Equal(t, "mocked_error", values[0].Error)
}

func TestService_GetProtocolsTotalValues_Allbridge_FailedFetchingStats(t *testing.T) {
	const allbridge = "allbridge"
	var errNil error

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":           allbridge,
		"total_messages":     uint64(4),
		"total_value_locked": float64(5),
	}))

	respActivityLast := &mockQueryTableResult{}
	respActivityLast.On("Next").Return(true)
	respActivityLast.On("Err").Return(errNil)
	respActivityLast.On("Close").Return(errNil)
	ts := time.Now().UTC()
	respActivityLast.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":                allbridge,
		"total_messages":          uint64(4),
		"total_value_transferred": float64(7),
		"total_volume_secure":     float64(9),
		"_time":                   ts,
	}))

	respActivity2 := &mockQueryTableResult{}
	respActivity2.On("Next").Return(true)
	respActivity2.On("Err").Return(errNil)
	respActivity2.On("Close").Return(errNil)
	respActivity2.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":                allbridge,
		"total_messages":          uint64(4),
		"total_value_transferred": float64(7),
		"total_value_secure":      float64(9),
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolStatsNow, "bucket30d", dbconsts.ProtocolsStatsMeasurementHourly, allbridge)).Return(&mockQueryTableResult{}, errors.New("mocked_error"))
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolStats24HrAgo, "bucket30d", dbconsts.ProtocolsStatsMeasurementHourly, allbridge)).Return(respStatsLastDay, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolActivity, "bucketInfinite", "1970-01-01T00:00:00Z", dbconsts.ProtocolsActivityMeasurementDaily, allbridge)).Return(respActivityLast, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolActivity, "bucket30d", ts.Format(time.RFC3339), dbconsts.ProtocolsActivityMeasurementHourly, allbridge)).Return(respActivity2, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryLast24HrActivity, "bucketInfinite", dbconsts.ProtocolsActivityMeasurementDaily, allbridge)).Return(respActivity2, nil)

	// core protocols influx calls
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolStats24HrAgo, "bucketInfinite")).Return(emptyQueryTableResult(), nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaSinceStartOfDay, "bucket30d")).Return(emptyQueryTableResult(), nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaLastDay, "bucket30d")).Return(emptyQueryTableResult(), nil)

	repository := protocols.NewRepository(queryAPI, "bucketInfinite", "bucket30d", "bucket24hr", zap.NewNop())
	service := protocols.NewService([]string{protocols.ALLBRIDGE}, repository, zap.NewNop(), cache.NewDummyCacheClient(), "WORMSCAN:PROTOCOLS", 0, metrics.NewNoOpMetrics(), &mockTvl{})

	values := service.GetProtocolsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, allbridge, values[0].Protocol)
	assert.NotNil(t, values[0].Error)
	assert.Equal(t, "mocked_error", values[0].Error)
}

func TestService_GetProtocolsTotalValues_Mayan(t *testing.T) {
	const mayan = "mayan"
	var errNil error
	respStatsLatest := &mockQueryTableResult{}
	respStatsLatest.On("Next").Return(true)
	respStatsLatest.On("Err").Return(errNil)
	respStatsLatest.On("Close").Return(errNil)
	respStatsLatest.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":       mayan,
		"total_messages": uint64(7),
		"volume":         float64(10),
	}))

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"protocol":       mayan,
		"total_messages": uint64(4),
		"volume":         float64(5),
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolStatsNow, "bucket30d", dbconsts.ProtocolsStatsMeasurementHourly, mayan)).Return(respStatsLatest, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.QueryTemplateProtocolStats24HrAgo, "bucket30d", dbconsts.ProtocolsStatsMeasurementHourly, mayan)).Return(respStatsLastDay, nil)

	// core protocols influx calls
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolStats24HrAgo, "bucketInfinite")).Return(emptyQueryTableResult(), nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaSinceStartOfDay, "bucket30d")).Return(emptyQueryTableResult(), nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaLastDay, "bucket30d")).Return(emptyQueryTableResult(), nil)

	repository := protocols.NewRepository(queryAPI, "bucketInfinite", "bucket30d", "bucket24hr", zap.NewNop())
	service := protocols.NewService([]string{protocols.MAYAN}, repository, zap.NewNop(), cache.NewDummyCacheClient(), "WORMSCAN:PROTOCOLS", 0, metrics.NewNoOpMetrics(), &mockTvl{})

	values := service.GetProtocolsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, mayan, values[0].Protocol)
	assert.Equal(t, uint64(7), values[0].TotalMessages)
	assert.Equal(t, 10.00, values[0].TotalValueTransferred)
	assert.Equal(t, uint64(3), values[0].LastDayMessages)
	assert.Equal(t, "75.00%", values[0].LastDayDiffPercentage)
	assert.Equal(t, float64(5), values[0].Last24HourVolume)

}

func TestService_GetProtocolsTotalValues_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockCache := &cacheMock.CacheMock{}
	var cacheErr error
	cacheErr = nil
	cachedValue := fmt.Sprintf(`{"result": [{"protocol":"protocol1","total_messages":7,"total_value_locked":5,"total_value_secured":9,"total_value_transferred":7,"last_day_messages":4,"last_day_diff_percentage":"75.00%%"}],"timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	mockCache.On("Get", ctx, "WORMSCAN:PROTOCOLS:ALL_PROTOCOLS").Return(cachedValue, cacheErr)
	service := protocols.NewService([]string{}, nil, zap.NewNop(), mockCache, "WORMSCAN:PROTOCOLS", 60, metrics.NewNoOpMetrics(), &mockTvl{})
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

func TestService_GetPortalTokenBridge_Stats(t *testing.T) {

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}

	// core protocols influx calls
	totalStartOfCurrentDay := &multirowQueryTableResult{
		Result: []*query.FluxRecord{
			query.NewFluxRecord(1, map[string]interface{}{
				"app_id":                  protocols.PortalTokenBridge,
				"total_messages":          uint64(50),
				"total_value_transferred": 3e8,
			}),
		},
	}

	deltaSinceStartOfDay := &multirowQueryTableResult{
		Result: []*query.FluxRecord{
			query.NewFluxRecord(1, map[string]interface{}{
				"app_id":                  protocols.PortalTokenBridge,
				"total_messages":          uint64(25),
				"total_value_transferred": 2e8,
			}),
		},
	}

	deltaLastDay := &multirowQueryTableResult{
		Result: []*query.FluxRecord{
			query.NewFluxRecord(1, map[string]interface{}{
				"app_id":                  protocols.PortalTokenBridge,
				"total_messages":          uint64(10),
				"total_value_transferred": 1e8,
			}),
		},
	}

	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolStats24HrAgo, "bucketInfinite")).Return(totalStartOfCurrentDay, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaSinceStartOfDay, "bucket30d")).Return(deltaSinceStartOfDay, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(protocols.AllProtocolsDeltaLastDay, "bucket30d")).Return(deltaLastDay, nil)

	repository := protocols.NewRepository(queryAPI, "bucketInfinite", "bucket30d", "bucket24hr", zap.NewNop())
	service := protocols.NewService([]string{}, repository, zap.NewNop(), cache.NewDummyCacheClient(), "WORMSCAN:PROTOCOLS", 0, metrics.NewNoOpMetrics(), &mockTvl{})
	values := service.GetProtocolsTotalValues(ctx)
	assert.NotNil(t, values)
	assert.Equal(t, 1, len(values))
	for i := range values {
		switch values[i].Protocol {
		case "portal_token_bridge":
			assert.Equal(t, uint64(75), values[i].TotalMessages)
			assert.Equal(t, 5.0, values[i].TotalValueTransferred)
			assert.Equal(t, uint64(10), values[i].LastDayMessages)
			assert.Equal(t, "15.38%", values[i].LastDayDiffPercentage)
			assert.Equal(t, 1235.523, values[i].TotalValueLocked)
		default:
			t.Errorf("unexpected protocol %s", values[i].Protocol)
		}
	}
}

type mockQueryAPI struct {
	mock.Mock
}

func (m *mockQueryAPI) Query(ctx context.Context, q string) (protocols.QueryResult, error) {
	args := m.Called(ctx, q)
	return args.Get(0).(protocols.QueryResult), args.Error(1)
}

func emptyQueryTableResult() *mockQueryTableResult {
	m := &mockQueryTableResult{}
	m.On("Next").Return(false)
	m.On("Err").Return(nil)
	m.On("Close").Return(nil)
	return m
}

type multirowQueryTableResult struct {
	Result []*query.FluxRecord
	index  int // this is to track how many times Next() has been called
}

func (m *multirowQueryTableResult) Next() bool {
	return m.index < len(m.Result)
}

func (m *multirowQueryTableResult) Record() *query.FluxRecord {
	record := m.Result[m.index]
	m.index++
	return record
}

func (m *multirowQueryTableResult) Err() error {
	return nil
}

func (m *multirowQueryTableResult) Close() error {
	return nil
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

type mockTvl struct {
}

func (t *mockTvl) Get(_ context.Context) (string, error) {
	return "1235.523", nil
}
