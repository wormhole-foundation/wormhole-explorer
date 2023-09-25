package metrics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/token"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/metric"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/prices"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type VaaConverter struct {
	MissingTokens            map[sdk.Address]sdk.ChainID
	MissingTokensCounter     map[sdk.Address]int
	PriceCache               *prices.CoinPricesCache
	Metrics                  metrics.Metrics
	GetTransferredTokenByVaa token.GetTransferredTokenByVaa
}

func NewVaaConverter(priceCache *prices.CoinPricesCache, GetTransferredTokenByVaa token.GetTransferredTokenByVaa) *VaaConverter {
	return &VaaConverter{
		MissingTokens:            make(map[sdk.Address]sdk.ChainID),
		MissingTokensCounter:     make(map[sdk.Address]int),
		PriceCache:               priceCache,
		Metrics:                  metrics.NewNoopMetrics(),
		GetTransferredTokenByVaa: GetTransferredTokenByVaa,
	}
}

func (c *VaaConverter) Convert(ctx context.Context, vaaBytes []byte) (string, error) {

	// Parse the VAA and payload
	vaa, err := sdk.Unmarshal(vaaBytes)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling vaa: %v", err)
	}
	transferredToken, err := c.GetTransferredTokenByVaa(ctx, vaa)
	if err != nil {
		return "", fmt.Errorf("error decoding payload: %v", err)
	}

	// Look up token metadata
	tokenMetadata, ok := domain.GetTokenByAddress(transferredToken.TokenChain, transferredToken.TokenAddress.String())
	if !ok {

		// if not found, add to missing tokens
		c.MissingTokens[transferredToken.TokenAddress] = transferredToken.TokenChain
		c.MissingTokensCounter[transferredToken.TokenAddress] = c.MissingTokensCounter[transferredToken.TokenAddress] + 1

		return "", fmt.Errorf("unknown token: %s %s", transferredToken.TokenChain.String(), transferredToken.TokenAddress.String())
	}

	// Generate a data point for the VAA volume metric
	var point *write.Point
	{
		p := metric.MakePointForVaaVolumeParams{
			Vaa: vaa,
			TokenPriceFunc: func(_ string, timestamp time.Time) (decimal.Decimal, error) {

				// fetch the historic price from cache
				price, err := c.PriceCache.GetPriceByTime(tokenMetadata.CoingeckoID, timestamp)
				if err != nil {
					return decimal.NewFromInt(0), err
				}

				return price, nil
			},
			Metrics:          c.Metrics,
			TransferredToken: transferredToken,
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
