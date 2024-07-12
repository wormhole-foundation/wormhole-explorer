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
	client *resty.Client
	log    *zap.Logger
}

func NewPricesApi(coingeckoURL string, log *zap.Logger) *PricesApi {
	return &PricesApi{
		client: resty.New().SetBaseURL(coingeckoURL),
		log:    log,
	}
}

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
