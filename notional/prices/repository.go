package prices

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type PriceDb struct {
	ID          string    `bson:"_id" json:"id"`
	CoingeckoID string    `bson:"coingeckoId" json:"coingeckoId"`
	Price       string    `bson:"price" json:"price"`
	Datetime    time.Time `bson:"dateTime" json:"dateTime"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}

type PriceRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
	logger     *zap.Logger
}

var ErrPriceNotFound = fmt.Errorf("price not found")

// NewPriceRepository creates a new price repository.
func NewPriceRepository(db *mongo.Database, logger *zap.Logger) *PriceRepository {
	return &PriceRepository{
		db:         db,
		collection: db.Collection("prices"),
		logger:     logger,
	}
}

// Upsert upserts a price.
func (p *PriceRepository) Upsert(ctx context.Context, coingeckoID string, price decimal.Decimal, dateTime time.Time) error {
	id := p.createID(coingeckoID, dateTime)
	model := &PriceDb{
		ID:          id,
		CoingeckoID: coingeckoID,
		Price:       price.Truncate(8).String(),
		Datetime:    dateTime,
		UpdatedAt:   time.Now(),
	}

	update := bson.M{
		"$set":         model,
		"$setOnInsert": repository.IndexedAt(time.Now()),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}

	opts := options.Update().SetUpsert(true)
	var err error
	_, err = p.collection.UpdateByID(ctx, model.ID, update, opts)
	return err
}

// Find finds a price.
func (p *PriceRepository) Find(ctx context.Context, coingeckoID string, dateTime time.Time) (*PriceDb, error) {
	id := p.createID(coingeckoID, dateTime)
	var price PriceDb
	err := p.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&price)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrPriceNotFound
		}
		return nil, err
	}
	return &price, err
}

func (p *PriceRepository) createID(coingeckoID string, dateTime time.Time) string {
	return fmt.Sprintf("%s-%s", coingeckoID, dateTime.UTC().Format(time.RFC3339))
}
