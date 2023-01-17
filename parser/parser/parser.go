package parser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const DefaultTimeout = 10

var (
	ErrCallEndpoint        = errors.New("ERROR CALL ENPOINT")
	ErrNotFound            = errors.New("NOT FOUND")
	ErrInternalError       = errors.New("INTERNAL ERROR")
	ErrUnproceesableEntity = errors.New("UNPROCESSABLE")
	ErrBadRequest          = errors.New("BAD REQUEST")
)

// ParseVaaResponse represent a parse vaa response.
type ParseVaaResponse struct {
	ChainID        uint16      `json:"chainId"`
	EmitterAddress string      `json:"address"`
	Sequence       string      `json:"sequence"`
	AppID          string      `json:"appId"`
	Result         interface{} `json:"result"`
}

// ParserVAAAPIClient parse vaa api client.
type ParserVAAAPIClient struct {
	Client  http.Client
	BaseURL string
	Logger  *zap.Logger
}

// NewParserVAAAPIClient create new instances of ParserVAAAPIClient.
func NewParserVAAAPIClient(timeout int64, baseURL string, logger *zap.Logger) (ParserVAAAPIClient, error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	if baseURL == "" {
		return ParserVAAAPIClient{}, errors.New("baseURL can not be empty")
	}

	return ParserVAAAPIClient{
		Client: http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		BaseURL: baseURL,
		Logger:  logger,
	}, nil
}

type ParseData struct {
	PayloadID int `bson:"payloadid"`
	Fields    interface{}
}

// ParseVaa invoke the endpoint to parse a VAA from the VAAParserAPI.
func (c ParserVAAAPIClient) Parse(chainID uint16, address, sequence string, vaa []byte) (*ParseVaaResponse, error) {
	endpointUrl := fmt.Sprintf("%s/vaa/parser/%v/%s/%v", c.BaseURL, chainID,
		address, sequence)

	// create request body.
	payload := struct {
		Payload []byte `json:"payload"`
	}{
		Payload: vaa,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		c.Logger.Error("error marshalling payload", zap.Error(err), zap.Uint16("chainID", chainID),
			zap.String("address", address), zap.String("sequence", sequence))
		return nil, err
	}

	response, err := c.Client.Post(endpointUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.Logger.Error("error call parse vaa endpoint", zap.Error(err), zap.Uint16("chainID", chainID),
			zap.String("address", address), zap.String("sequence", sequence))
		return nil, ErrCallEndpoint
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusCreated:
		var parsedVAA ParseVaaResponse
		json.NewDecoder(response.Body).Decode(&parsedVAA)
		return &parsedVAA, nil
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusUnprocessableEntity:
		return nil, ErrUnproceesableEntity
	default:
		return nil, ErrInternalError
	}
}
