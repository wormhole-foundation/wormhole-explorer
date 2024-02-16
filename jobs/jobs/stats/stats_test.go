package stats_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/mock"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/stats"
	"go.uber.org/zap"
	"io"
	"net/http"
	"testing"
)

func Test_ContributorsStatsJob_Succeed(t *testing.T) {
	var mockErr error
	statsFetcher := &mockStatsFetch{}
	statsFetcher.On("Get", mock.Anything).Return(stats.Stats{}, mockErr)
	statsFetcher.On("ContributorName", mock.Anything).Return("contributor_test")
	mockWriterDB := &mockWriterApi{}
	mockWriterDB.On("WritePoint", mock.Anything, mock.Anything).Return(mockErr)

	job := stats.NewContributorsStatsJob(mockWriterDB, zap.NewNop(), "v1", statsFetcher)
	resultErr := job.Run(context.Background())
	assert.Nil(t, resultErr)
}

func Test_ContributorsStatsJob_FailFetching(t *testing.T) {
	var mockErr error
	statsFetcher := &mockStatsFetch{}
	statsFetcher.On("Get", mock.Anything).Return(stats.Stats{}, errors.New("mocked_error_fetch"))
	statsFetcher.On("ContributorName", mock.Anything).Return("contributor_test")
	mockWriterDB := &mockWriterApi{}
	mockWriterDB.On("WritePoint", mock.Anything, mock.Anything).Return(mockErr)

	job := stats.NewContributorsStatsJob(mockWriterDB, zap.NewNop(), "v1", statsFetcher)
	resultErr := job.Run(context.Background())
	assert.NotNil(t, resultErr)
	assert.Equal(t, "mocked_error_fetch", resultErr.Error())
}

func Test_ContributorsStatsJob_FailedUpdatingDB(t *testing.T) {
	var mockErr error
	statsFetcher := &mockStatsFetch{}
	statsFetcher.On("Get", mock.Anything).Return(stats.Stats{}, mockErr)
	statsFetcher.On("ContributorName", mock.Anything).Return("contributor_test")
	mockWriterDB := &mockWriterApi{}
	mockWriterDB.On("WritePoint", mock.Anything, mock.Anything).Return(errors.New("mocked_error_update_db"))

	job := stats.NewContributorsStatsJob(mockWriterDB, zap.NewNop(), "v1", statsFetcher)
	resultErr := job.Run(context.Background())
	assert.NotNil(t, resultErr)
	assert.Equal(t, "mocked_error_update_db", resultErr.Error())
}

func Test_HttpRestClientStats_FailRequestCreation(t *testing.T) {

	a := stats.NewHttpRestClientStats("contributor_test", "localhost", zap.NewNop(),
		mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return nil, nil
		}))
	_, err := a.Get(nil) // passing ctx nil to force request creation error
	assert.NotNil(t, err)
}

func Test_HttpRestClientStats_FailedRequestExecution(t *testing.T) {

	a := stats.NewHttpRestClientStats("contributor_test", "localhost", zap.NewNop(),
		mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("mocked_http_client_do")
		}))
	_, err := a.Get(context.Background())
	assert.NotNil(t, err)
	assert.Equal(t, "mocked_http_client_do", err.Error())
}

func Test_HttpRestClientStats_Status500(t *testing.T) {

	a := stats.NewHttpRestClientStats("contributor_test", "localhost", zap.NewNop(),
		mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("respones_body_test")),
			}, nil
		}))
	_, err := a.Get(context.Background())
	assert.NotNil(t, err)
	assert.Equal(t, "failed retrieving client stats from url:localhost - status_code:500 - response_body:respones_body_test", err.Error())
}

func Test_HttpRestClientStats_Status200_FailedReadBody(t *testing.T) {

	a := stats.NewHttpRestClientStats("contributor_test", "localhost", zap.NewNop(),
		mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockFailReadCloser{},
			}, nil
		}))
	_, err := a.Get(context.Background())
	assert.NotNil(t, err)
	assert.Equal(t, "failed reading response body from client stats. url:localhost - status_code:200: mocked_fail_read", err.Error())
}

func Test_HttpRestClientStats_Status200_FailedParsing(t *testing.T) {

	a := stats.NewHttpRestClientStats("contributor_test", "localhost", zap.NewNop(),
		mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("this should be a json")),
			}, nil
		}))
	_, err := a.Get(context.Background())
	assert.NotNil(t, err)
	assert.Equal(t, "failed unmarshalling response body from client stats. url:localhost - status_code:200 - response_body:this should be a json: invalid character 'h' in literal true (expecting 'r')", err.Error())
}

func Test_HttpRestClientStats_Status200_Succeed(t *testing.T) {

	a := stats.NewHttpRestClientStats("contributor_test", "localhost", zap.NewNop(),
		mockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("{\"total_value_locked\":\"123\",\"total_messages\":\"456\"}")),
			}, nil
		}))
	resp, err := a.Get(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "123", resp.TotalValueLocked)
	assert.Equal(t, "456", resp.TotalMessages)
}

// mock influxdb WriterApiBlocking interface
type mockWriterApi struct {
	mock.Mock
}

func (m *mockWriterApi) WriteRecord(ctx context.Context, line ...string) error {
	args := m.Called(ctx, line)
	return args.Error(0)
}

func (m *mockWriterApi) WritePoint(ctx context.Context, point ...*write.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}

func (m *mockWriterApi) EnableBatching() {
}

func (m *mockWriterApi) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type mockStatsFetch struct {
	mock.Mock
}

func (m *mockStatsFetch) Get(ctx context.Context) (stats.Stats, error) {
	args := m.Called(ctx)
	return args.Get(0).(stats.Stats), args.Error(1)
}

func (m *mockStatsFetch) ContributorName() string {
	args := m.Called()
	return args.String(0)
}

type mockHttpClient func(req *http.Request) (*http.Response, error)

func (m mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m(req)
}

type mockFailReadCloser struct {
}

func (m *mockFailReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("mocked_fail_read")
}

func (m *mockFailReadCloser) Close() error {
	return nil
}
