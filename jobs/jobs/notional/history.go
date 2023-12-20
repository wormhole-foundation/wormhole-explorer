package notional

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/coingecko"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type HistoryNotionalJob struct {
	coingeckoAPI     *coingecko.CoinGeckoAPI
	db               *mongo.Database
	p2pNetwork       string
	requestLimitTime time.Duration
	days             string
	logger           *zap.Logger
}

type PriceUpdate struct {
	ID          string    `bson:"_id" json:"id"`
	CoingeckoID string    `bson:"coingeckoId" json:"coingeckoId"`
	Price       string    `bson:"price" json:"price"`
	Datetime    time.Time `bson:"dateTime" json:"dateTime"`
	UpdatedAt   time.Time `bson:"updatedAt" json:"updatedAt"`
}

func NewHistoryNotionalJob(api *coingecko.CoinGeckoAPI, db *mongo.Database, p2pNetwork string, requestLimitTimeSeconds int, days string, logger *zap.Logger) *HistoryNotionalJob {
	return &HistoryNotionalJob{
		coingeckoAPI:     api,
		p2pNetwork:       p2pNetwork,
		requestLimitTime: time.Duration(requestLimitTimeSeconds) * time.Second,
		db:               db,
		days:             days,
		logger:           logger,
	}
}

func (h *HistoryNotionalJob) Run(ctx context.Context) error {

	prices := h.db.Collection("prices")

	// create token provider
	tokenProvider := domain.NewTokenProvider(h.p2pNetwork)
	tokens := tokenProvider.GetAllCoingeckoIDs()
	sort.StringSlice(tokens).Sort()

	h.logger.Info("found tokens", zap.Int("count", len(tokens)), zap.String("priceDays", h.days))
	for index, token := range tokens {
		log := h.logger.With(zap.String("coingeckoID", token), zap.Int("index", index+1), zap.Int("count", len(tokens)))
		r, err := h.coingeckoAPI.GetSymbolDailyPrice(token, h.days)
		if err != nil {
			log.Error("failed to get price", zap.Error(err))
			if errors.Is(err, coingecko.ErrTooManyRequests) {
				time.Sleep(h.requestLimitTime * 3)
			}
			time.Sleep(h.requestLimitTime)
			continue
		}
		log.Info("processing token", zap.Int("prices", len(r.Prices)))
		var lastDateTime time.Time
		for _, p := range r.Prices {
			dateTimeMilli := p[0].IntPart()
			dateTime := time.UnixMilli(dateTimeMilli).Truncate(24 * time.Hour).UTC()
			id := fmt.Sprintf("%s-%s", token, dateTime.Format(time.RFC3339))
			if dateTime.Equal(lastDateTime) {
				continue
			}
			update := &PriceUpdate{
				ID:          id,
				CoingeckoID: token,
				Price:       p[1].Truncate(8).String(),
				Datetime:    dateTime,
				UpdatedAt:   time.Now(),
			}

			err := h.upsertPrice(ctx, prices, update)
			if err != nil {
				log.Error("failed to upsert price", zap.Error(err))
			}
			lastDateTime = dateTime
		}
		time.Sleep(h.requestLimitTime)
	}

	return nil

}

// UpsertParsedVaa saves vaa information and parsed result.
func (h *HistoryNotionalJob) upsertPrice(ctx context.Context, collection *mongo.Collection, price *PriceUpdate) error {
	update := bson.M{
		"$set":         price,
		"$setOnInsert": indexedAt(time.Now()),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}

	opts := options.Update().SetUpsert(true)
	var err error
	_, err = collection.UpdateByID(ctx, price.ID, update, opts)
	return err
}

func indexedAt(t time.Time) IndexingTimestamps {
	return IndexingTimestamps{
		IndexedAt: t,
	}
}

type IndexingTimestamps struct {
	IndexedAt time.Time `bson:"indexedAt"`
}
