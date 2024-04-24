package guardian

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const DefaultTimeout = 10

// GuardianAPIClient guardian api client.
type GuardianAPIClient struct {
	Client  http.Client
	BaseURL string
	Logger  *zap.Logger
}

// NewGuardianAPIClient create new instances of GuardianAPIClient.
func NewGuardianAPIClient(timeout int64, baseURL string, logger *zap.Logger) (GuardianAPIClient, error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	if baseURL == "" {
		return GuardianAPIClient{}, errors.New("baseURL can not be empty")
	}

	return GuardianAPIClient{
		Client: http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
		BaseURL: baseURL,
		Logger:  logger,
	}, nil
}

// SignedVaa represent a guardianAPI signed vaa response.
type SignedVaa struct {
	VaaBytes []byte `json:"vaaBytes"`
}

// GetSignedVAA get signed vaa.
func (c *GuardianAPIClient) GetSignedVAA(vaaID string) (*SignedVaa, error) {
	endpointUrl := fmt.Sprintf("%s/v1/signed_vaa/%s", c.BaseURL, vaaID)
	resp, err := c.Client.Get(endpointUrl)
	if err != nil {
		c.Logger.Error("failed to call endpoint", zap.String("endpoint", endpointUrl), zap.Error(err))
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		c.Logger.Error("failed to call endpoint", zap.String("endpoint", endpointUrl), zap.Int("status_code", resp.StatusCode))
		return nil, errors.New("failed to call endpoint, status code is not 200")
	}

	var signedVaa SignedVaa
	err = json.NewDecoder(resp.Body).Decode(&signedVaa)
	if err != nil {
		c.Logger.Error("failed to decode response", zap.Error(err))
		return nil, err
	}
	return &signedVaa, nil
}
