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
		watcherBlock       *mongo.Collection
		globalTransactions *mongo.Collection
	}
}

// NewRepository create a new respository instance.
func NewRepository(db *mongo.Database, log *zap.Logger) *Repository {
	return &Repository{db, log, struct {
		watcherBlock       *mongo.Collection
		globalTransactions *mongo.Collection
	}{
		watcherBlock:       db.Collection("watcherBlock"),
		globalTransactions: db.Collection("globalTransactions"),
	}}
}

func indexedAt(t time.Time) IndexingTimestamps {
	return IndexingTimestamps{
		IndexedAt: t,
	}
}

func (s *Repository) UpsertGlobalTransaction(ctx context.Context, globalTransactions TransactionUpdate) error {
	update := bson.M{
		"$set":         globalTransactions,
		"$setOnInsert": indexedAt(time.Now()),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}

	_, err := s.collections.globalTransactions.UpdateByID(ctx, globalTransactions.ID, update, options.Update().SetUpsert(true))
	if err != nil {
		s.log.Error("Error inserting global transaction", zap.Error(err))
		return err
	}

	return err

}

func (s *Repository) UpdateWatcherBlock(ctx context.Context, watcherBlock WatcherBlock) error {
	update := bson.M{
		"$set":         watcherBlock,
		"$setOnInsert": indexedAt(time.Now()),
	}
	_, err := s.collections.watcherBlock.UpdateByID(ctx, watcherBlock.ID, update, options.Update().SetUpsert(true))
	if err != nil {
		s.log.Error("Error inserting watcher block", zap.Error(err))
		return err
	}
	return err
}

func (s *Repository) GetCurrentBlock(ctx context.Context, blockchain string) (int64, error) {
	var block WatcherBlock
	err := s.collections.watcherBlock.FindOne(ctx, bson.M{"_id": blockchain}).Decode(&block)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil
		}
		return 0, err
	}
	return block.BlockNumber, nil
}
