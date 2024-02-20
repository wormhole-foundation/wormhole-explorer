package activity_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/mock"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/activity"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/activity/internal/repositories"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/internal/commons/mocks"
	"go.uber.org/zap"
	"testing"
	"time"
)

func Test_ProtocolsActivityJob_Succeed(t *testing.T) {
	var mockErr error
	activityFetcher := &mockActivityFetch{}
	act := repositories.ProtocolActivity{
		Activities: []repositories.Activity{
			{
				EmitterChainID:     1,
				DestinationChainID: 2,
				Txs:                150,
				TotalUSD:           250000,
			},
		},
	}

	activityFetcher.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(act, mockErr)
	activityFetcher.On("ProtocolName", mock.Anything).Return("protocol_test")
	mockWriterDB := &mocks.MockWriterApi{}
	mockWriterDB.On("WritePoint", mock.Anything, mock.Anything).Return(mockErr)

	job := activity.NewProtocolActivityJob(mockWriterDB, zap.NewNop(), "v1", activityFetcher)
	resultErr := job.Run(context.Background())
	assert.Nil(t, resultErr)
}

func Test_ProtocolsActivityJob_FailFetching(t *testing.T) {
	var mockErr error
	activityFetcher := &mockActivityFetch{}
	activityFetcher.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(repositories.ProtocolActivity{}, errors.New("mocked_error_fetch"))
	activityFetcher.On("ProtocolName", mock.Anything).Return("protocol_test")
	mockWriterDB := &mocks.MockWriterApi{}
	mockWriterDB.On("WritePoint", mock.Anything, mock.Anything).Return(mockErr)

	job := activity.NewProtocolActivityJob(mockWriterDB, zap.NewNop(), "v1", activityFetcher)
	resultErr := job.Run(context.Background())
	assert.NotNil(t, resultErr)
	assert.Equal(t, "mocked_error_fetch", resultErr.Error())
}

func Test_ProtocolsActivityJob_FailedUpdatingDB(t *testing.T) {
	var mockErr error
	activityFetcher := &mockActivityFetch{}
	activityFetcher.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(repositories.ProtocolActivity{}, mockErr)
	activityFetcher.On("ProtocolName", mock.Anything).Return("protocol_test")
	mockWriterDB := &mocks.MockWriterApi{}
	mockWriterDB.On("WritePoint", mock.Anything, mock.Anything).Return(errors.New("mocked_error_update_db"))

	job := activity.NewProtocolActivityJob(mockWriterDB, zap.NewNop(), "v1", activityFetcher)
	resultErr := job.Run(context.Background())
	assert.NotNil(t, resultErr)
	assert.Equal(t, "mocked_error_update_db", resultErr.Error())
}

type mockActivityFetch struct {
	mock.Mock
}

func (m *mockActivityFetch) Get(ctx context.Context, from, to time.Time) (repositories.ProtocolActivity, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(repositories.ProtocolActivity), args.Error(1)
}

func (m *mockActivityFetch) ProtocolName() string {
	args := m.Called()
	return args.String(0)
}
