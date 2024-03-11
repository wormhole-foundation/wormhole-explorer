package protocols_test

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/mock"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/internal/commons"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/internal/commons/mocks"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/repository"
	"go.uber.org/zap"
	"testing"
	"time"
)

func Test_ProtocolsStatsJob_Success(t *testing.T) {

	ctx := context.Background()
	from, _ := time.Parse(time.RFC3339, "2024-02-01T00:00:00Z")
	to, _ := time.Parse(time.RFC3339, "2024-02-02T00:00:00Z")

	mr := &mockProtocolRepo{}
	activity := repository.ProtocolActivity{
		TotalValueSecure:      10,
		TotalValueTransferred: 20,
		Volume:                30,
		TotalMessages:         40,
		Activities: []repository.Activity{
			{
				EmitterChainID:     1,
				DestinationChainID: 2,
				Txs:                50,
				TotalUSD:           60,
			},
			{
				EmitterChainID:     1,
				DestinationChainID: 2,
				Txs:                25,
				TotalUSD:           30,
			},
		},
	}

	stats := repository.Stats{
		TotalValueLocked: 70,
		TotalMessages:    80,
		Volume:           90,
	}

	mr.On("GetActivity", ctx, from, to).Return(activity, nil)
	mr.On("GetStats", ctx).Return(stats, nil)
	mr.On("ProtocolName").Return(commons.MayanProtocol)

	mockWriterDB := &mocks.MockWriterApi{}

	expectedStatsPoint := influxdb2.
		NewPointWithMeasurement(dbconsts.ProtocolsStatsMeasurementDaily).
		AddTag("protocol", commons.MayanProtocol).
		AddField("total_messages", stats.TotalMessages).
		AddField("total_value_locked", stats.TotalValueLocked).
		AddField("volume", stats.Volume).
		SetTime(from)

	expectedActivityPoint := influxdb2.NewPointWithMeasurement(dbconsts.ProtocolsActivityMeasurementDaily).
		AddTag("protocol", commons.MayanProtocol).
		AddField("total_value_secure", activity.TotalValueSecure).
		AddField("total_value_transferred", activity.TotalValueTransferred).
		AddField("volume", activity.Volume).
		AddField("txs", 75).
		AddField("total_usd", 90).
		SetTime(from)

	mockWriterDB.On("WritePoint", ctx, mock.MatchedBy(pointMatcher{Expected: expectedStatsPoint}.Matches)).Return(nil)
	mockWriterDB.On("WritePoint", ctx, mock.MatchedBy(pointMatcher{Expected: expectedActivityPoint}.Matches)).Return(nil).Times(1)

	job := protocols.NewStatsJob(mockWriterDB,
		from,
		to,
		dbconsts.ProtocolsActivityMeasurementDaily,
		dbconsts.ProtocolsStatsMeasurementDaily,
		[]repository.ProtocolRepository{mr},
		zap.NewNop())

	err := job.Run(ctx)
	assert.Nil(t, err)
	mockWriterDB.AssertNumberOfCalls(t, "WritePoint", 2)
}

func Test_ProtocolsStatsJob_FailedFetchingStats(t *testing.T) {

}

func Test_ProtocolsStatsJob_FailedFetchingActivity(t *testing.T) {

}

type mockProtocolRepo struct {
	mock.Mock
}

func (m *mockProtocolRepo) GetActivity(ctx context.Context, from, to time.Time) (repository.ProtocolActivity, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(repository.ProtocolActivity), args.Error(1)
}
func (m *mockProtocolRepo) GetStats(ctx context.Context) (repository.Stats, error) {
	args := m.Called(ctx)
	return args.Get(0).(repository.Stats), args.Error(1)
}
func (m *mockProtocolRepo) ProtocolName() string {
	args := m.Called()
	return args.String(0)
}

type pointMatcher struct {
	Expected *write.Point
}

func (p pointMatcher) Matches(x interface{}) bool {
	actual, ok := x.([]*write.Point)
	if !ok || len(actual) != 1 {
		return false
	}

	// Perform your comparison logic here
	// For example, check if the measurement name matches
	return actual[0].Name() == p.Expected.Name()
}

func (p pointMatcher) String() string {
	return "matches the expected point"
}
