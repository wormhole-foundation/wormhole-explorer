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
	url    string
	client *http.Client
}

// NewCoingeckoAPI creates a new coingecko client
func NewCoingeckoAPI(url string) *CoingeckoAPI {
	return &CoingeckoAPI{
		url:    url,
		client: http.DefaultClient,
	}
}

// NotionalUSD is the response from the coingecko API.
type NotionalUSD struct {
	Price *decimal.Decimal `json:"usd"`
}

// GetNotionalUSD returns the notional USD value for the given ids
// ids is a list of coingecko chain identifier.
func (c *CoingeckoAPI) GetNotionalUSD(ids []string) (map[string]NotionalUSD, error) {
	var response map[string]NotionalUSD
	notionalUrl := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", c.url, strings.Join(ids, ","))

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
	err = json.Unmarshal(body, &response)
	return response, err
}

// GetChainIDs returns the coingecko chain ids for the given p2p network.
func GetChainIDs(p2pNetwork string) []string {

	if p2pNetwork == domain.P2pMainNet {
		return domain.GetAllCoingeckoIDs()
	}

	// TODO: define chains ids for testnet.
	return []string{}
}
