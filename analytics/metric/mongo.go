package metric

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/token"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// TransferPriceDoc models a document in the `transferPrices` collection
type TransferPriceDoc struct {
	// ID is the unique identifier of the VAA for which we are storing price information.
	ID string `bson:"_id"`
	// Timestamp is the timestamp of the VAA for which we are storing price information.
	Timestamp time.Time `bson:"timestamp"`
	// Symbol is the trading symbol of the token being transferred.
	Symbol string `bson:"symbol"`
	// SymbolPriceUsd is the price of the token in USD at the moment of the transfer.
	SymbolPriceUsd string `bson:"price"`
	// TokenAmount is the amount of the token being transferred.
	TokenAmount string `bson:"tokenAmount"`
	// UsdAmount is the value in USD of the token being transferred.
	UsdAmount string `bson:"usdAmount"`
	// TokenChain is the chain ID of the token being transferred.
	TokenChain uint16 `bson:"tokenChain"`
	// TokenAddress is the address of the token being transferred.
	TokenAddress string `bson:"tokenAddress"`
	// CoinGeckoID is the CoinGecko ID of the token being transferred.
	CoinGeckoID string `bson:"coinGeckoId"`
	// UpdatedAt is the timestamp the document was updated.
	UpdatedAt time.Time `bson:"updatedAt"`
}

func UpsertTransferPrices(
	ctx context.Context,
	logger *zap.Logger,
	vaa *sdk.VAA,
	transferPrices *mongo.Collection,
	tokenPriceFunc func(tokenID, coinGeckoID string, timestamp time.Time) (decimal.Decimal, error),
	transferredToken *token.TransferredToken,
	tokenProvider *domain.TokenProvider,
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

	// Upsert the `TransferPrices` collection
	update := bson.M{
		"$set": TransferPriceDoc{
			ID:             vaa.MessageID(),
			Timestamp:      vaa.Timestamp,
			Symbol:         tokenMeta.Symbol.String(),
			SymbolPriceUsd: notionalUSD.Truncate(8).String(),
			TokenAmount:    tokenAmount.Truncate(8).String(),
			UsdAmount:      usdAmount.Truncate(8).String(),
			TokenChain:     uint16(transferredToken.TokenChain),
			TokenAddress:   transferredToken.TokenAddress.String(),
			CoinGeckoID:    tokenMeta.CoingeckoID,
			UpdatedAt:      time.Now(),
		},
	}
	_, err = transferPrices.UpdateByID(
		ctx,
		vaa.MessageID(),
		update,
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("failed to update transfer price collection: %w", err)
	}

	return nil
}
