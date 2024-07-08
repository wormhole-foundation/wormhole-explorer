package prices

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type getPriceByTimeResponse struct {
	CoingeckoID string `json:"coingeckoId"`
	Symbol      string `json:"symbol"`
	Price       string `json:"price"`
	DateTime    string `json:"dateTime"`
}

type PricesApi struct {
	client            *resty.Client
	log               *zap.Logger
	coingeckoHaderKey string
	coingeckoApiKey   string
}

func NewPricesApi(coingeckoURL, coingeckoHeaderKey, coingeckoApiKey string, log *zap.Logger) *PricesApi {
	return &PricesApi{
		client:            resty.New().SetBaseURL(coingeckoURL),
		log:               log,
		coingeckoHaderKey: coingeckoHeaderKey,
		coingeckoApiKey:   coingeckoApiKey,
	}
}

// GetPriceByTime
// Deprecated: use GetPriceAtTime
func (n *PricesApi) GetPriceByTime(ctx context.Context, coingeckoID string, dateTime time.Time) (decimal.Decimal, error) {
	url := fmt.Sprintf("/api/coingecko/prices/%s/%s", coingeckoID, dateTime.Format(time.RFC3339))
	req := n.client.R()

	resp, err := req.
		SetContext(ctx).
		SetResult(&getPriceByTimeResponse{}).
		Get(url)

	if err != nil {
		return decimal.Zero, err
	}

	if resp.IsError() {
		return decimal.Zero, fmt.Errorf("status code: %s. %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*getPriceByTimeResponse)
	if result == nil {
		return decimal.Zero, fmt.Errorf("empty response")
	}

	return decimal.NewFromString(result.Price)
}

type getPriceAtTimeResponse struct {
	Id         string `json:"id"`
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	MarketData struct {
		CurrentPrice struct {
			Usd string `json:"usd"`
		} `json:"current_price"`
	} `json:"market_data"`
}

// GetPriceAtTime fetches the price of a token at a specific time
func (n *PricesApi) GetPriceAtTime(ctx context.Context, coingeckoID string, dateTime time.Time) (decimal.Decimal, error) {
	url := fmt.Sprintf("/api/v3/coins/%s/history?localization=false&date=%s", coingeckoID, dateTime.Format("02-01-2006"))
	req := n.client.R()

	if n.coingeckoHaderKey != "" && n.coingeckoApiKey != "" {
		req.SetHeader(n.coingeckoHaderKey, n.coingeckoApiKey)
	}

	resp, err := req.
		SetContext(ctx).
		SetResult(&getPriceByTimeResponse{}).
		Get(url)

	if err != nil {
		return decimal.Zero, err
	}

	if resp.IsError() {
		return decimal.Zero, fmt.Errorf("status code: %s. %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*getPriceAtTimeResponse)
	if result == nil {
		return decimal.Zero, fmt.Errorf("empty response")
	}

	return decimal.NewFromString(result.MarketData.CurrentPrice.Usd)
}
