package txtracker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const DefaultTimeout = 30

var (
	ErrCallEndpoint  = errors.New("ERROR CALL ENPOINT")
	ErrBadRequest    = errors.New("BAD REQUEST")
	ErrInternalError = errors.New("INTERNAL ERROR")
)

// TxTrackerAPIClient tx tracker api client.
type TxTrackerAPIClient struct {
	Client  http.Client
	BaseURL string
	Logger  *zap.Logger
}

// NewTxTrackerAPIClient create new instances of TxTrackerAPIClient.
func NewTxTrackerAPIClient(timeout int64, baseURL string, logger *zap.Logger) (TxTrackerAPIClient, error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	if baseURL == "" {
		return TxTrackerAPIClient{}, errors.New("baseURL can not be empty")
	}

	return TxTrackerAPIClient{
		Client: http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		BaseURL: baseURL,
		Logger:  logger,
	}, nil
}

// ProcessVaaResponse represent a process vaa response.
type ProcessVaaResponse struct {
	From         string `json:"from"`
	NativeTxHash string `json:"nativeTxHash"`
	Attributes   any    `json:"attributes"`
}

// Process process vaa.
func (c *TxTrackerAPIClient) Process(vaaID string) (*ProcessVaaResponse, error) {
	endpointUrl := fmt.Sprintf("%s/vaa/process", c.BaseURL)

	// create request body.
	payload := struct {
		VaaID string `json:"id"`
	}{
		VaaID: vaaID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		c.Logger.Error("error marshalling payload", zap.Error(err), zap.String("vaaID", vaaID))
		return nil, err
	}

	response, err := c.Client.Post(endpointUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.Logger.Error("error call parse vaa endpoint", zap.Error(err), zap.String("vaaID", vaaID))
		return nil, ErrCallEndpoint
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		var processVaaResponse ProcessVaaResponse
		json.NewDecoder(response.Body).Decode(&processVaaResponse)
		return &processVaaResponse, nil
	case http.StatusInternalServerError:
		return nil, ErrInternalError
	default:
		return nil, ErrInternalError
	}
}

// CreateTxHashFunc represent a create tx hash function.
type CreateTxHashFunc func(vaaID, txHash string) (*TxHashResponse, error)

// TxHashResponse represent a create tx hash response.
type TxHashResponse struct {
	NativeTxHash string `json:"nativeTxHash"`
}

// CreateTxHash create tx hash.
func (c *TxTrackerAPIClient) CreateTxHash(vaaID, txHash string) (*TxHashResponse, error) {
	endpoint := fmt.Sprintf("%s/vaa/tx-hash", c.BaseURL)

	// create request body.
	payload := struct {
		VaaID  string `json:"id"`
		TxHash string `json:"txHash"`
	}{
		VaaID:  vaaID,
		TxHash: txHash,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		c.Logger.Error("error marshalling payload", zap.Error(err), zap.String("vaaID", vaaID), zap.String("txHash", txHash))
		return nil, err
	}

	response, err := c.Client.Post(endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.Logger.Error("error call create tx hash endpoint",
			zap.Error(err),
			zap.String("vaaID", vaaID),
			zap.String("txHash", txHash))
		return nil, ErrCallEndpoint
	}

	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		var txHashResponse TxHashResponse
		json.NewDecoder(response.Body).Decode(&txHashResponse)
		return &txHashResponse, nil
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusInternalServerError:
		return nil, ErrInternalError
	default:
		return nil, ErrInternalError
	}

}
