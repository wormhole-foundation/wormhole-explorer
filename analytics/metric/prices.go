package metric

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/token"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/storage"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

func UpsertTransferPrices(
	ctx context.Context,
	logger *zap.Logger,
	vaa *sdk.VAA,
	pricesRepository storage.PricesRepository,
	tokenPriceFunc func(tokenID, coinGeckoID string, timestamp time.Time) (decimal.Decimal, error),
	transferredToken *token.TransferredToken,
	tokenProvider *domain.TokenProvider,
	source, trackID string,
) error {

	// Do not generate this metric for PythNet VAAs
	if vaa.EmitterChain == sdk.ChainIDPythNet {
		return nil
	}

	// Do not generate this metric if the VAA is not a transfer
	if transferredToken == nil {
		return nil
	}

	// Get the token metadata
	//
	// This is complementary data about the token that is not present in the VAA itself.
	tokenMeta, ok := tokenProvider.GetTokenByAddress(transferredToken.TokenChain, transferredToken.TokenAddress.String())
	if !ok {
		return nil
	}

	// Try to obtain the token notional value from the cache
	notionalUSD, err := tokenPriceFunc(tokenMeta.GetTokenID(), tokenMeta.CoingeckoID, vaa.Timestamp)
	if err != nil {
		logger.Warn("failed to obtain notional for this token",
			zap.String("vaaId", vaa.MessageID()),
			zap.String("tokenAddress", transferredToken.TokenAddress.String()),
			zap.Uint16("tokenChain", uint16(transferredToken.TokenChain)),
			zap.Any("tokenMetadata", tokenMeta),
			zap.String("timestamp", vaa.Timestamp.Format(time.RFC3339)),
			zap.Error(err),
		)
		return nil
	}

	// Compute the amount with decimals
	var exp int32
	if tokenMeta.Decimals > 8 {
		exp = 8
	} else {
		exp = int32(tokenMeta.Decimals)
	}
	tokenAmount := decimal.NewFromBigInt(transferredToken.Amount, -exp)

	// Compute the amount in USD
	usdAmount := tokenAmount.Mul(notionalUSD)

	tp := storage.OperationPrice{
		Source:        source,
		TrackID:       trackID,
		Digest:        utils.NormalizeHex(vaa.HexDigest()),
		VaaID:         vaa.MessageID(),
		ChainID:       vaa.EmitterChain,
		Timestamp:     vaa.Timestamp,
		TokenChainID:  uint16(transferredToken.TokenChain),
		TokenAddress:  transferredToken.TokenAddress.String(),
		Symbol:        tokenMeta.Symbol.String(),
		TokenUSDPrice: notionalUSD.Truncate(8),
		TotalToken:    tokenAmount.Truncate(8),
		TotalUSD:      usdAmount.Truncate(8),
		CoinGeckoID:   tokenMeta.CoingeckoID,
		UpdatedAt:     time.Now(),
	}

	err = pricesRepository.Upsert(ctx, tp)

	if err != nil {
		return fmt.Errorf("failed to update transfer price collection: %w", err)
	}

	return nil
}
