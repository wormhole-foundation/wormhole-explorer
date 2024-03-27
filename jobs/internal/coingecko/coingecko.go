package coingecko

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// CoingeckoAPI is a client for the coingecko API
type CoingeckoAPI struct {
	url       string
	chunkSize int
	client    *http.Client
	logger    *zap.Logger
	headerKey string
	apiKey    string
}

// NewCoingeckoAPI creates a new coingecko client
func NewCoingeckoAPI(url string, headerKey, apiKey string, logger *zap.Logger) *CoingeckoAPI {
	return &CoingeckoAPI{
		url:       url,
		chunkSize: 200,
		client:    http.DefaultClient,
		headerKey: headerKey,
		apiKey:    apiKey,
		logger:    logger,
	}
}

// NotionalUSD is the response from the coingecko API.
type NotionalUSD struct {
	Price *decimal.Decimal `json:"usd"`
}

// GetNotionalUSD returns the notional USD value for the given ids
// ids is a list of coingecko chain identifier.
func (c *CoingeckoAPI) GetNotionalUSD(ids []string) (map[string]NotionalUSD, error) {
	response := map[string]NotionalUSD{}
	chunksIds := chunkChainIds(ids, c.chunkSize)

	c.logger.Info("fetching notional value of assets", zap.Int("total_chunks", len(chunksIds)))

	// iterate over chunks of ids.
	for i, chunk := range chunksIds {

		notionalUrl := fmt.Sprintf("%s/api/v3/simple/price?ids=%s&vs_currencies=usd", c.url, strings.Join(chunk, ","))

		req, err := http.NewRequest(http.MethodGet, notionalUrl, nil)
		if err != nil {
			return response, err
		}
		if c.headerKey != "" && c.apiKey != "" {
			req.Header.Add(c.headerKey, c.apiKey)
		}

		res, err := c.client.Do(req)
		if err != nil {
			return response, err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			c.logger.Error("failed to get notional value of assets", zap.Int("statusCode", res.StatusCode), zap.Int("chunk", i))
			return response, fmt.Errorf("failed to get notional value of assets, status code: %d", res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return response, err
		}

		chunkResponse := map[string]NotionalUSD{}
		err = json.Unmarshal(body, &chunkResponse)
		if err != nil {
			return response, err
		}

		// merge chunk response with response.
		for k, v := range chunkResponse {
			response[k] = v
		}
	}

	return response, nil
}

func chunkChainIds(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}
