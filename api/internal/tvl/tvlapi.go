package tvl

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/tidwall/gjson"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
)

var (
	endpoints map[string]string = map[string]string{
		domain.P2pMainNet: "https://europe-west3-wormhole-315720.cloudfunctions.net/mainnet-notionaltvl",
		domain.P2pTestNet: "https://europe-west3-wormhole-315720.cloudfunctions.net/mainnet-notionaltvl",
	}
)

type TvlAPI struct {
	url    string
	client *http.Client
}

// NewCoingeckoAPI creates a new coingecko client
func NewTvlAPI(net string) *TvlAPI {
	return &TvlAPI{
		client: http.DefaultClient,
		url:    endpoints[net],
	}
}

func (c *TvlAPI) GetNotionalUSD(ctx context.Context, ids []string) (*string, error) {

	// Build the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, err
	}

	// Send it
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Extract TVL from the response
	tvl := gjson.Get(string(body), "AllTime.\\*.\\*.Notional")
	response := tvl.String()
	return &response, nil

}
