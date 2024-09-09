package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/analytics/config"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

// TransferPrice models a document in the `transferPrices` collection
type TransferPrice struct {
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

// MongoPricesRepository represents the repository for prices.
type MongoPricesRepository struct {
	metrics        metrics.Metrics
	transferPrices *mongo.Collection
	logger         *zap.Logger
}

func NewMongoPricesRepository(db *mongo.Database, metrics metrics.Metrics, logger *zap.Logger) *MongoPricesRepository {
	return &MongoPricesRepository{
		metrics:        metrics,
		transferPrices: db.Collection("transferPrices"),
		logger:         logger,
	}
}

func (r *MongoPricesRepository) Upsert(ctx context.Context, o OperationPrice) error {
	// Upsert the `TransferPrices` collection
	_, err := r.transferPrices.UpdateByID(
		ctx,
		o.VaaID,
		bson.M{"$set": TransferPrice{
			ID:             o.VaaID,
			Timestamp:      o.Timestamp,
			Symbol:         o.Symbol,
			SymbolPriceUsd: o.TokenUSDPrice.String(),
			TokenAmount:    o.TotalToken.String(),
			UsdAmount:      o.TotalUSD.String(),
			TokenChain:     o.TokenChainID,
			TokenAddress:   o.TokenAddress,
			CoinGeckoID:    o.CoinGeckoID,
			UpdatedAt:      time.Now(),
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("failed to update transfer price collection: %w", err)
	}
	r.metrics.IncTransferPricesInserted(config.DbLayerMongo)
	return nil
}

// MongoVaaRepository repository struct definition.
type MongoVaaRepository struct {
	db            *mongo.Database
	logger        *zap.Logger
	vaas          *mongo.Collection
	vaaRepository *repository.VaaRepository
}

// VaaDoc vaa document struct definition.
type VaaDoc struct {
	ID  string `bson:"_id" json:"id"`
	Vaa []byte `bson:"vaas" json:"vaa"`
}

// NewRepository create a new Repository.
func NewMongoVaaRepository(db *mongo.Database, logger *zap.Logger) *MongoVaaRepository {
	return &MongoVaaRepository{db: db,
		logger:        logger.With(zap.String("module", "MongoVaaRepository")),
		vaas:          db.Collection("vaas"),
		vaaRepository: repository.NewVaaRepository(db, logger),
	}
}

// FindById find a vaa by id.
func (r *MongoVaaRepository) FindByVaaID(ctx context.Context, id string) (*Vaa, error) {
	var vaaDoc VaaDoc
	err := r.vaas.FindOne(ctx, bson.M{"_id": id}).Decode(&vaaDoc)
	if err != nil {
		return nil, err
	}
	return &Vaa{
		ID:    vaaDoc.ID,
		VaaID: vaaDoc.ID,
		Vaa:   vaaDoc.Vaa,
	}, err
}

func (r *MongoVaaRepository) FindPage(ctx context.Context, query VaaPageQuery, pagination Pagination) ([]*Vaa, error) {
	q := repository.VaaQuery{
		StartTime:      query.StartTime,
		EndTime:        query.EndTime,
		EmitterChainID: query.EmitterChainID,
		EmitterAddress: query.EmitterAddress,
		Sequence:       query.Sequence,
	}

	p := repository.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		SortAsc:  pagination.SortAsc,
	}
	result, err := r.vaaRepository.FindPage(ctx, q, p)
	if err != nil {
		return nil, err
	}

	vaas := make([]*Vaa, 0, len(result))
	for _, v := range result {
		vaas = append(vaas, &Vaa{
			ID:    v.ID,
			VaaID: v.ID,
			Vaa:   v.Vaa,
		})
	}
	return vaas, nil
}
