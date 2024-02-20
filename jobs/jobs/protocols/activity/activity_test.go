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
				EmitterChainID:     "0x1",
				DestinationChainID: "0x2",
				Txs:                150,
				TotalUSD:           250000,
			},
		},
	}

	activityFetcher.On("Get", mock.Anything).Return(act, mockErr)
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
	activityFetcher.On("Get", mock.Anything).Return(repositories.ProtocolActivity{}, errors.New("mocked_error_fetch"))
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
	activityFetcher.On("Get", mock.Anything).Return(repositories.ProtocolActivity{}, mockErr)
	activityFetcher.On("ProtocolName", mock.Anything).Return("protocol_test")
	mockWriterDB := &mocks.MockWriterApi{}
	mockWriterDB.On("WritePoint", mock.Anything, mock.Anything).Return(errors.New("mocked_error_update_db"))

	job := activity.NewProtocolActivityJob(mockWriterDB, zap.NewNop(), "v1", activityFetcher)
	resultErr := job.Run(context.Background())
	assert.NotNil(t, resultErr)
	assert.Equal(t, "mocked_error_update_db", resultErr.Error())
}

/*
func Test_HttpRestClientActivity_FailRequestCreation(t *testing.T) {

	a := repositories.NewHttpRestClientActivity("protocol_test", "localhost", zap.NewNop(),
		stats.mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return nil, nil
		}))
	_, err := a.Get(nil, time.Now(), time.Now()) // passing ctx nil to force request creation error
	assert.NotNil(t, err)
}

func Test_HttpRestClientActivity_FailedRequestExecution(t *testing.T) {

	a := repositories.NewHttpRestClientActivity("protocol_test", "localhost", zap.NewNop(),
		stats.mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("mocked_http_client_do")
		}))
	_, err := a.Get(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "mocked_http_client_do", err.Error())
}

func Test_HttpRestClientActivity_Status500(t *testing.T) {

	a := repositories.NewHttpRestClientActivity("protocol_test", "localhost", zap.NewNop(),
		stats.mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("response_body_test")),
			}, nil
		}))
	_, err := a.Get(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "failed retrieving protocol Activities from url:localhost - status_code:500 - response_body:response_body_test", err.Error())
}

func Test_HttpRestClientActivity_Status200_FailedReadBody(t *testing.T) {

	a := repositories.NewHttpRestClientActivity("protocol_test", "localhost", zap.NewNop(),
		stats.mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &stats.mockFailReadCloser{},
			}, nil
		}))
	_, err := a.Get(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "failed reading response body from protocol Activities. url:localhost - status_code:200: mocked_fail_read", err.Error())
}

func Test_HttpRestClientActivity_Status200_FailedParsing(t *testing.T) {

	a := repositories.NewHttpRestClientActivity("protocol_test", "localhost", zap.NewNop(),
		stats.mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("this should be a json")),
			}, nil
		}))
	_, err := a.Get(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "failed unmarshalling response body from protocol Activities. url:localhost - status_code:200 - response_body:this should be a json: invalid character 'h' in literal true (expecting 'r')", err.Error())
}

func Test_HttpRestClientActivity_Status200_Succeed(t *testing.T) {

	a := repositories.NewHttpRestClientActivity("protocol_test", "localhost", zap.NewNop(),
		stats.mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("{\"total_value_secured\":\"123\",\"total_value_transferred\":\"456\",\"Activities\":[{\"emitter_chain_id\":\"0x123\",\"destination_chain_id\":\"0x999\",\"txs\":\"4\",\"total_usd\":\"5000\"}]}")),
			}, nil
		}))
	resp, err := a.Get(context.Background(), time.Now(), time.Now())
	assert.Nil(t, err)
	assert.Equal(t, "123", resp.TotalValueSecure)
	assert.Equal(t, "456", resp.TotalValueTransferred)
	assert.Equal(t, 1, len(resp.Groups))
	assert.Equal(t, "0x123", resp.Groups[0].EmitterChainID)
	assert.Equal(t, "0x999", resp.Groups[0].DestinationChainID)
	assert.Equal(t, "4", resp.Groups[0].Txs)
	assert.Equal(t, "5000", resp.Groups[0].TotalUSD)
}

*/

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
