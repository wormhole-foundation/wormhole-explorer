package storage

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// repository errors
var ErrDocNotFound = errors.New("NOT FOUND")

// Repository definitions.
type Repository struct {
	db          *mongo.Database
	log         *zap.Logger
	collections struct {
		redeemed    *mongo.Collection
		transaction *mongo.Collection
	}
}

// NewRepository create a new respository instance.
func NewRepository(db *mongo.Database, log *zap.Logger) *Repository {
	return &Repository{db, log, struct {
		redeemed    *mongo.Collection
		transaction *mongo.Collection
	}{
		redeemed:    db.Collection("redeemed"),
		transaction: db.Collection("transaction"),
	}}
}

func indexedAt(t time.Time) IndexingTimestamps {
	return IndexingTimestamps{
		IndexedAt: t,
	}
}

func (s *Repository) UpsertRedeemed(ctx context.Context, redeemed RedeemedUpdate) error {
	update := bson.M{
		"$set":         redeemed,
		"$setOnInsert": indexedAt(time.Now()),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}

	_, err := s.collections.redeemed.UpdateByID(ctx, redeemed.ID, update, options.Update().SetUpsert(true))
	if err != nil {
		s.log.Error("Error inserting redeemed", zap.Error(err))
		return err
	}

	return err

}
