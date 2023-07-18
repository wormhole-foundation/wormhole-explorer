package coingecko

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
)

// CoingeckoAPI is a client for the coingecko API
type CoingeckoAPI struct {
	url       string
	chunkSize int
	client    *http.Client
}

// NewCoingeckoAPI creates a new coingecko client
func NewCoingeckoAPI(url string) *CoingeckoAPI {
	return &CoingeckoAPI{
		url:       url,
		chunkSize: 200,
		client:    http.DefaultClient,
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

	// iterate over chunks of ids.
	for _, chunk := range chunksIds {
		notionalUrl := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", c.url, strings.Join(chunk, ","))

		req, err := http.NewRequest(http.MethodGet, notionalUrl, nil)
		if err != nil {
			return response, err
		}
		res, err := c.client.Do(req)
		if err != nil {
			return response, err
		}
		defer res.Body.Close()

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

// GetChainIDs returns the coingecko chain ids for the given p2p network.
func GetChainIDs(p2pNetwork string) []string {

	if p2pNetwork == domain.P2pMainNet {
		return domain.GetAllCoingeckoIDs()
	}

	// TODO: define chains ids for testnet.
	return []string{}
}
