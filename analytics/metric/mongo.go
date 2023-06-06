package metric

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// TransferPriceDoc models a document in the `transferPrices` collection
type TransferPriceDoc struct {
	ID        string    `bson:"_id"`
	Timestamp time.Time `bson:"timestamp"`
	Symbol    string    `bson:"symbol"`
	Price     string    `bson:"price"`
}

func upsertTransferPrices(
	logger *zap.Logger,
	vaa *sdk.VAA,
	transferPrices *mongo.Collection,
	tokenPriceFunc func(symbol domain.Symbol, timestamp time.Time) (decimal.Decimal, error),
) error {

	// Do not generate this metric for PythNet VAAs
	if vaa.EmitterChain == sdk.ChainIDPythNet {
		return nil
	}

	// Decode the VAA payload
	payload, err := sdk.DecodeTransferPayloadHdr(vaa.Payload)
	if err != nil {
		return nil
	}

	// Get the token metadata
	//
	// This is complementary data about the token that is not present in the VAA itself.
	tokenMeta, ok := domain.GetTokenByAddress(payload.OriginChain, payload.OriginAddress.String())
	if !ok {
		return nil
	}

	// Try to obtain the token notional value from the cache
	notionalUSD, err := tokenPriceFunc(tokenMeta.Symbol, vaa.Timestamp)
	if err != nil {
		logger.Warn("failed to obtain notional for this token",
			zap.String("vaaId", vaa.MessageID()),
			zap.String("tokenAddress", payload.OriginAddress.String()),
			zap.Uint16("tokenChain", uint16(payload.OriginChain)),
			zap.Any("tokenMetadata", tokenMeta),
			zap.Error(err),
		)
		return nil
	}

	// Upsert the `TransferPrices` collection
	update := bson.M{
		"$set": TransferPriceDoc{
			ID:        vaa.MessageID(),
			Timestamp: vaa.Timestamp,
			Symbol:    tokenMeta.Symbol.String(),
			Price:     notionalUSD.Truncate(8).String(),
		},
	}
	_, err = transferPrices.UpdateByID(
		context.Background(),
		vaa.MessageID(),
		update,
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("failed to update transfer price collection: %w", err)
	}

	return nil
}
