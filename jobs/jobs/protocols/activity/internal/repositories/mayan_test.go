package repositories

import (
	"bytes"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/internal/commons/mocks"
	"go.uber.org/zap"
	"io"
	"net/http"
	"testing"
	"time"
)

func Test_HttpRestClientActivity_FailRequestCreation(t *testing.T) {

	a := NewMayanRestClient("protocol_test", "localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return nil, nil
		}))
	_, err := a.Get(nil, time.Now(), time.Now()) // passing ctx nil to force request creation error
	assert.NotNil(t, err)
}

func Test_HttpRestClientActivity_FailedRequestExecution(t *testing.T) {

	a := NewMayanRestClient("protocol_test", "localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("mocked_http_client_do")
		}))
	_, err := a.Get(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "mocked_http_client_do", err.Error())
}

func Test_HttpRestClientActivity_Status500(t *testing.T) {

	a := NewMayanRestClient("protocol_test", "localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
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

	a := NewMayanRestClient("protocol_test", "localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mocks.MockFailReadCloser{},
			}, nil
		}))
	_, err := a.Get(context.Background(), time.Now(), time.Now())
	assert.NotNil(t, err)
	assert.Equal(t, "failed reading response body from protocol Activities. url:localhost - status_code:200: mocked_fail_read", err.Error())
}

func Test_HttpRestClientActivity_Status200_FailedParsing(t *testing.T) {

	a := NewMayanRestClient("protocol_test", "localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
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

	a := NewMayanRestClient("protocol_test", "localhost", zap.NewNop(),
		mocks.MockHttpClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("{\"total_value_secure\":1640898.7106282723,\"total_value_transferred\":2600395.040031102,\"total_messages\":2225,\"activity\":[{\"emmiter_chain_id\":\"1\",\"destination_chain_id\":\"2\",\"txs\":88,\"total_usd\":648500.9762709612}],\"volume\":2761848.9678057004}")),
			}, nil
		}))
	resp, err := a.Get(context.Background(), time.Now(), time.Now())
	assert.Nil(t, err)
	assert.Equal(t, 1640898.7106282723, resp.TotalValueSecure)
	assert.Equal(t, 2600395.040031102, resp.TotalValueTransferred)
	assert.Equal(t, uint64(2225), resp.TotalMessages)
	assert.Equal(t, 2761848.9678057004, resp.Volume)
	assert.Equal(t, 1, len(resp.Activities))
	assert.Equal(t, uint64(1), resp.Activities[0].EmitterChainID)
	assert.Equal(t, uint64(2), resp.Activities[0].DestinationChainID)
	assert.Equal(t, uint64(88), resp.Activities[0].Txs)
	assert.Equal(t, 648500.9762709612, resp.Activities[0].TotalUSD)
}
