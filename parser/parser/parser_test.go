package parser

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

var timeout int64 = 10

type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip interface implementation.
func (r RoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return r(request), nil
}

// NewParserVAAAPITestClient create a
func NewParserVAAAPITestClient(rountTripFunc RoundTripFunc) ParserVAAAPIClient {
	parserVaaClient, _ := NewParserVAAAPIClient(timeout, "", zap.NewExample())
	parserVaaClient.Client = http.Client{
		Timeout:   time.Duration(10) * time.Second,
		Transport: RoundTripFunc(rountTripFunc),
	}
	return parserVaaClient
}

// TestSuccessVAAParser test success vaa parser.
func TestSuccessVAAParser(t *testing.T) {
	parserVaaClient := NewParserVAAAPITestClient(func(request *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader(`{"appID": "PORTAL_TOKEN_BRIDGE","chainId": 4, "address": "000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585", "sequence": "226769", "result": {"fee": 0,"type": "Transfer","payloadId": 1,"amount": 10000000,"tokenAddress": "0x000000000000000000000000dac17f958d2ee523a2206206994597c13d831ec7","tokenChain": 2,"toAddress": "0x0000000000000000000000000ff664edd699bd85610c2782d9dbbbad704b6fc5","chain": 5}}`)),
		}
	})
	var chainID uint16 = 4
	address := "000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585"
	sequence := "226769"
	parserVaaResponse, err := parserVaaClient.Parse(chainID, address, sequence, []byte{})
	if err != nil {
		t.Error("expected err zero value, got %w", err)
	}
	if parserVaaResponse == nil {
		t.Error("expected parserVaaResponse value, got nil")
	} else {
		if chainID != parserVaaResponse.ChainID {
			t.Errorf("expected chainID %v, got %v", chainID, parserVaaResponse.ChainID)
		}
		if parserVaaResponse.EmitterAddress != address {
			t.Errorf("expected address %s, got %s", address, parserVaaResponse.EmitterAddress)
		}
		if parserVaaResponse.Sequence != sequence {
			t.Errorf("expected sequence %s, got %s", sequence, parserVaaResponse.Sequence)
		}
	}
}

// TestNotFoundVaaParser test vaa parser not found.
func TestNotFoundVaaParser(t *testing.T) {
	parserVaaClient := NewParserVAAAPITestClient(func(request *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusNotFound,
		}
	})
	parserVaaResponse, err := parserVaaClient.Parse(4, "000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585", "226769", []byte{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %s", err.Error())
	}
	if parserVaaResponse != nil {
		t.Error("expected parserVaaResponse zero value, got %w", parserVaaResponse)
	}
}

// TestBadRequestVaaParser test vaa parser bad request.
func TestBadRequestVaaParser(t *testing.T) {
	parserVaaClient := NewParserVAAAPITestClient(func(request *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
		}
	})
	parserVaaResponse, err := parserVaaClient.Parse(4, "000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585", "226769", []byte{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got %s", err.Error())
	}
	if parserVaaResponse != nil {
		t.Error("expected parserVaaResponse zero value, got %w", parserVaaResponse)
	}
}

// TestUnprocessableVaaParser test vaa parser unprocessable request.
func TestUnprocessableVaaParser(t *testing.T) {
	parserVaaClient := NewParserVAAAPITestClient(func(request *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
		}
	})
	parserVaaResponse, err := parserVaaClient.Parse(4, "000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585", "26769", []byte{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, ErrUnproceesableEntity) {
		t.Errorf("expected ErrUnproceesableEntity, got %s", err.Error())
	}
	if parserVaaResponse != nil {
		t.Error("expected parserVaaResponse zero value, got %w", parserVaaResponse)
	}
}

// TestInternalErrorVaaParser test vaa parser internal error request.
func TestInternalErrorVaaParser(t *testing.T) {
	parserVaaClient := NewParserVAAAPITestClient(func(request *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
		}
	})
	parserVaaResponse, err := parserVaaClient.Parse(4, "000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585", "226769", []byte{})
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, ErrInternalError) {
		t.Errorf("expected ErrInternalError, got %s", err.Error())
	}
	if parserVaaResponse != nil {
		t.Error("expected parserVaaResponse zero value, got %w", parserVaaResponse)
	}
}
