package contributors_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/mock"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/contributors"
	"go.uber.org/zap"
	"testing"
)

func TestService_GetContributorsTotalValues(t *testing.T) {
	var errNil error
	respStatsLatest := &mockQueryTableResult{}
	respStatsLatest.On("Next").Return(true)
	respStatsLatest.On("Err").Return(errNil)
	respStatsLatest.On("Close").Return(errNil)
	respStatsLatest.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"contributor":        "contributor1",
		"total_messages":     "7",
		"total_value_locked": "5",
	}))

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"contributor":        "contributor1",
		"total_messages":     "4",
		"total_value_locked": "5",
	}))

	respActivityLast := &mockQueryTableResult{}
	respActivityLast.On("Next").Return(true)
	respActivityLast.On("Err").Return(errNil)
	respActivityLast.On("Close").Return(errNil)
	respActivityLast.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"contributor":             "contributor1",
		"total_messages":          "4",
		"total_value_transferred": "7",
		"total_volume_secure":     "9",
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(contributors.QueryTemplateLatestPoint, "contributors_bucket", "contributors_stats", "contributor1", "v1")).Return(respStatsLatest, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(contributors.QueryTemplateLast24Point, "contributors_bucket", "contributors_stats", "contributor1", "v1")).Return(respStatsLastDay, nil)

	activityQuery := fmt.Sprintf(contributors.QueryTemplateActivityLatestPoint, "contributors_bucket", "contributors_activity", "contributor1", "v1")
	queryAPI.On("Query", ctx, activityQuery).Return(respActivityLast, nil)

	repository := contributors.NewRepository(queryAPI, "contributors_bucket", "contributors_bucket", "v1", "v1", zap.NewNop())
	service := contributors.NewService([]string{"contributor1"}, repository, zap.NewNop())

	values := service.GetContributorsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "contributor1", values[0].Contributor)
	assert.Equal(t, "5", values[0].TotalValueLocked)
	assert.Equal(t, "7", values[0].TotalMessages)
	assert.Equal(t, "9", values[0].TotalValueSecured)
	assert.Equal(t, "7", values[0].TotalValueTransferred)
	assert.Equal(t, "3", values[0].LastDayMessages)
	assert.Equal(t, "75.00%", values[0].LastDayDiffPercentage)

}

func TestService_GetContributorsTotalValues_FailedFetchingActivity(t *testing.T) {
	var errNil error
	respStatsLatest := &mockQueryTableResult{}
	respStatsLatest.On("Next").Return(true)
	respStatsLatest.On("Err").Return(errNil)
	respStatsLatest.On("Close").Return(errNil)
	respStatsLatest.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"contributor":        "contributor1",
		"total_messages":     "7",
		"total_value_locked": "5",
	}))

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"contributor":        "contributor1",
		"total_messages":     "4",
		"total_value_locked": "5",
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(contributors.QueryTemplateLatestPoint, "contributors_bucket", "contributors_stats", "contributor1", "v1")).Return(respStatsLatest, nil)
	queryAPI.On("Query", ctx, fmt.Sprintf(contributors.QueryTemplateLast24Point, "contributors_bucket", "contributors_stats", "contributor1", "v1")).Return(respStatsLastDay, nil)

	activityQuery := fmt.Sprintf(contributors.QueryTemplateActivityLatestPoint, "contributors_bucket", "contributors_activity", "contributor1", "v1")
	queryAPI.On("Query", ctx, activityQuery).Return(&api.QueryTableResult{}, errors.New("mocked_fetching_activity_error"))

	repository := contributors.NewRepository(queryAPI, "contributors_bucket", "contributors_bucket", "v1", "v1", zap.NewNop())
	service := contributors.NewService([]string{"contributor1"}, repository, zap.NewNop())

	values := service.GetContributorsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "contributor1", values[0].Contributor)
	assert.NotNil(t, values[0].Error)
	assert.Equal(t, "mocked_fetching_activity_error", values[0].Error)
}

func TestService_GetContributorsTotalValues_FailedFetchingStats(t *testing.T) {
	var errNil error

	respStatsLastDay := &mockQueryTableResult{}
	respStatsLastDay.On("Next").Return(true)
	respStatsLastDay.On("Err").Return(errNil)
	respStatsLastDay.On("Close").Return(errNil)
	respStatsLastDay.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"contributor":        "contributor1",
		"total_messages":     "4",
		"total_value_locked": "5",
	}))

	respActivityLast := &mockQueryTableResult{}
	respActivityLast.On("Next").Return(true)
	respActivityLast.On("Err").Return(errNil)
	respActivityLast.On("Close").Return(errNil)
	respActivityLast.On("Record").Return(query.NewFluxRecord(1, map[string]interface{}{
		"contributor":             "contributor1",
		"total_messages":          "4",
		"total_value_transferred": "7",
		"total_volume_secure":     "9",
	}))

	ctx := context.Background()
	queryAPI := &mockQueryAPI{}
	queryAPI.On("Query", ctx, fmt.Sprintf(contributors.QueryTemplateLatestPoint, "contributors_bucket", "contributors_stats", "contributor1", "v1")).Return(&api.QueryTableResult{}, errors.New("mocked_fetching_stats_error"))
	queryAPI.On("Query", ctx, fmt.Sprintf(contributors.QueryTemplateLast24Point, "contributors_bucket", "contributors_stats", "contributor1", "v1")).Return(respStatsLastDay, nil)

	activityQuery := fmt.Sprintf(contributors.QueryTemplateActivityLatestPoint, "contributors_bucket", "contributors_activity", "contributor1", "v1")
	queryAPI.On("Query", ctx, activityQuery).Return(respActivityLast, errNil)

	repository := contributors.NewRepository(queryAPI, "contributors_bucket", "contributors_bucket", "v1", "v1", zap.NewNop())
	service := contributors.NewService([]string{"contributor1"}, repository, zap.NewNop())

	values := service.GetContributorsTotalValues(ctx)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "contributor1", values[0].Contributor)
	assert.NotNil(t, values[0].Error)
	assert.Equal(t, "mocked_fetching_stats_error", values[0].Error)
}

type mockQueryAPI struct {
	mock.Mock
}

func (m *mockQueryAPI) Query(ctx context.Context, q string) (contributors.QueryResult, error) {
	args := m.Called(ctx, q)
	return args.Get(0).(contributors.QueryResult), args.Error(1)
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
