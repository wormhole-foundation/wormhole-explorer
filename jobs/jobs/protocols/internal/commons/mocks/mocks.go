package mocks

import (
	"context"
	"errors"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/test-go/testify/mock"
	"net/http"
)

// MockWriterApi mock influxdb WriterApiBlocking interface
type MockWriterApi struct {
	mock.Mock
}

func (m *MockWriterApi) WriteRecord(ctx context.Context, line ...string) error {
	args := m.Called(ctx, line)
	return args.Error(0)
}

func (m *MockWriterApi) WritePoint(ctx context.Context, point ...*write.Point) error {
	args := m.Called(ctx, point)
	return args.Error(0)
}

func (m *MockWriterApi) EnableBatching() {
}

func (m *MockWriterApi) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockHttpClient func(req *http.Request) (*http.Response, error)

func (m MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m(req)
}

type MockFailReadCloser struct {
}

func (m *MockFailReadCloser) Read(_ []byte) (n int, err error) {
	return 0, errors.New("mocked_fail_read")
}

func (m *MockFailReadCloser) Close() error {
	return nil
}
