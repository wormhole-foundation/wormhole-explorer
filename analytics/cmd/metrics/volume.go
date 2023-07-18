package metrics

import (
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/metric"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/prices"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type VaaConverter struct {
	MissingTokens        map[sdk.Address]sdk.ChainID
	MissingTokensCounter map[sdk.Address]int
	PriceCache           *prices.CoinPricesCache
	Metrics              metrics.Metrics
}

func NewVaaConverter(priceCache *prices.CoinPricesCache) *VaaConverter {
	return &VaaConverter{
		MissingTokens:        make(map[sdk.Address]sdk.ChainID),
		MissingTokensCounter: make(map[sdk.Address]int),
		PriceCache:           priceCache,
		Metrics:              metrics.NewNoopMetrics(),
	}
}

func (c *VaaConverter) Convert(vaaBytes []byte) (string, error) {

	// Parse the VAA and payload
	vaa, err := sdk.Unmarshal(vaaBytes)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling vaa: %v", err)
	}
	payload, err := sdk.DecodeTransferPayloadHdr(vaa.Payload)
	if err != nil {
		return "", fmt.Errorf("error decoding payload: %v", err)
	}

	// Look up token metadata
	tokenMetadata, ok := domain.GetTokenByAddress(vaa.EmitterChain, payload.OriginAddress.String())
	if !ok {

		// if not found, add to missing tokens
		c.MissingTokens[payload.OriginAddress] = vaa.EmitterChain
		c.MissingTokensCounter[payload.OriginAddress] = c.MissingTokensCounter[payload.OriginAddress] + 1

		return "", fmt.Errorf("unknown token: %s %s", payload.OriginChain.String(), payload.OriginAddress.String())
	}

	// Generate a data point for the VAA volume metric
	var point *write.Point
	{
		p := metric.MakePointForVaaVolumeParams{
			Vaa: vaa,
			TokenPriceFunc: func(_ domain.Symbol, timestamp time.Time) (decimal.Decimal, error) {

				// fetch the historic price from cache
				price, err := c.PriceCache.GetPriceByTime(tokenMetadata.CoingeckoID, timestamp)
				if err != nil {
					return decimal.NewFromInt(0), err
				}

				return price, nil
			},
			Metrics: c.Metrics,
		}

		var err error
		point, err = metric.MakePointForVaaVolume(&p)
		if err != nil {
			return "", fmt.Errorf("failed to create data point for VAA volume metric: %v", err)
		}
		if point == nil {
			// Some VAAs don't generate any data points for this metric (e.g.: PythNet, non-token-bridge VAAs)
			return "", errors.New("can't generate point for VAA volume metric")
		}
	}

	// Convert the data point to line protocol
	result := convertPointToLineProtocol(point)
	return result, nil
}
