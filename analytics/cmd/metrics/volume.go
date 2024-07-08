package metrics

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"sync"
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
	MissingTokens            sync.Map
	MissingTokensCounter     sync.Map
	PriceCache               *prices.CoinPricesCache
	Metrics                  metrics.Metrics
	GetTransferredTokenByVaa token.GetTransferredTokenByVaa
	TokenProvider            *domain.TokenProvider
	zap.Logger
}

func NewVaaConverter(priceCache *prices.CoinPricesCache,
	GetTransferredTokenByVaa token.GetTransferredTokenByVaa,
	tokenProvider *domain.TokenProvider,
) *VaaConverter {
	return &VaaConverter{
		MissingTokens:            sync.Map{},
		MissingTokensCounter:     sync.Map{},
		PriceCache:               priceCache,
		Metrics:                  metrics.NewNoopMetrics(),
		GetTransferredTokenByVaa: GetTransferredTokenByVaa,
		TokenProvider:            tokenProvider,
	}
}

func (c *VaaConverter) Convert(ctx context.Context, vaaBytes []byte) (*token.TransferredToken, *write.Point, string, error) {

	// Parse the VAA and payload
	vaa, err := sdk.Unmarshal(vaaBytes)
	if err != nil {
		return nil, nil, "", fmt.Errorf("error unmarshaling vaa: %v", err)
	}
	transferredToken, err := c.GetTransferredTokenByVaa(ctx, vaa)
	if err != nil {
		return transferredToken, nil, "", fmt.Errorf("error decoding payload: %v", err)
	}

	if transferredToken == nil {
		return nil, nil, "", errors.New("transferred token is nil")
	}

	// Look up token metadata
	var tokenMetadata *domain.TokenMetadata
	tokenMetadata, ok := c.TokenProvider.GetTokenByAddress(transferredToken.TokenChain, transferredToken.TokenAddress.String())
	if !ok {
		// if not found, add to missing tokens
		c.MissingTokens.Store(transferredToken.TokenAddress, transferredToken.TokenChain)
		counter, found := c.MissingTokensCounter.Load(transferredToken.TokenAddress)
		missingTokenCounter := uint64(1)
		if found {
			missingTokenCounter = counter.(uint64)
			missingTokenCounter++
		}
		c.MissingTokensCounter.Store(transferredToken.TokenAddress, missingTokenCounter)
		return transferredToken, nil, "", fmt.Errorf("unknown token: %s %s", transferredToken.TokenChain.String(), transferredToken.TokenAddress.String())
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
			TokenProvider:    c.TokenProvider,
		}

		var err error
		point, err = metric.MakePointForVaaVolume(&p)
		if err != nil {
			return transferredToken, point, "", fmt.Errorf("failed to create data point for VAA volume metric: %v", err)
		}
		if point == nil {
			// Some VAAs don't generate any data points for this metric (e.g.: PythNet, non-token-bridge VAAs)
			return transferredToken, point, "", errors.New(fmt.Sprintf("can't generate point for VAA volume metric. vaaId:%s", vaa.MessageID()))
		}
	}

	// Convert the data point to line protocol
	result := convertPointToLineProtocol(point)
	return transferredToken, point, result, nil
}
