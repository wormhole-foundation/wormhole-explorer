package repository_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/internal/commons/mocks"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/repository"
	"go.uber.org/zap"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test_AllbridgeRestClientActivity_FailRequestCreation(t *testing.T) {

	a := repository.NewAllBridgeRestClient("localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return nil, nil
		}))
	_, err := a.GetActivity(nil, time.Now(), time.Now()) // passing ctx nil to force request creation error
	assert.NotNil(t, err)
}

func Test_AllbridgeRestClientActivity_FailedRequestExecution(t *testing.T) {

	a := repository.NewAllBridgeRestClient("localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("mocked_http_client_do")
		}))
	_, err := a.GetActivity(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "mocked_http_client_do", err.Error())
}

func Test_AllbridgeRestClientActivity_Status500(t *testing.T) {

	a := repository.NewAllBridgeRestClient("localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("response_body_test")),
			}, nil
		}))
	_, err := a.GetActivity(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "failed retrieving protocol Activities from baseURL:localhost/wormhole/activity - status_code:500 - response_body:response_body_test", err.Error())
}

func Test_AllbridgeRestClientActivity_Status200_FailedReadBody(t *testing.T) {

	a := repository.NewAllBridgeRestClient("localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mocks.MockFailReadCloser{},
			}, nil
		}))
	_, err := a.GetActivity(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "failed reading response body from protocol Activities. baseURL:localhost - status_code:200: mocked_fail_read", err.Error())
}

func Test_AllbridgeRestClientActivity_Status200_FailedParsing(t *testing.T) {

	a := repository.NewAllBridgeRestClient("localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("this should be a json")),
			}, nil
		}))
	_, err := a.GetActivity(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "failed unmarshalling response body from protocol Activities. baseURL:localhost - status_code:200 - response_body:this should be a json: invalid character 'h' in literal true (expecting 'r')", err.Error())
}

func Test_AllbridgeRestClientActivity_Status200_Succeed(t *testing.T) {

	a := repository.NewAllBridgeRestClient("localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("{\"activity\":[{\"emitter_chain_id\":5,\"destination_chain_id\":1,\"txs\":\"1827\",\"total_usd\":\"445743.185719500000\"}],\"total_value_secure\":\"0\",\"total_value_transferred\":\"5734947.136079277\"}")),
			}, nil
		}))
	resp, err := a.GetActivity(context.Background(), time.Now(), time.Now())
	assert.Nil(t, err)
	assert.Equal(t, float64(0), resp.TotalValueSecure)
	assert.Equal(t, 5734947.136079277, resp.TotalValueTransferred)
	assert.Equal(t, 1, len(resp.Activities))
	assert.Equal(t, uint64(5), resp.Activities[0].EmitterChainID)
	assert.Equal(t, uint64(1), resp.Activities[0].DestinationChainID)
	assert.Equal(t, uint64(1827), resp.Activities[0].Txs)
	assert.Equal(t, 445743.185719500000, resp.Activities[0].TotalUSD)
}
