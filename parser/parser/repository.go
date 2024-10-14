package parser

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// repository errors
var ErrDocNotFound = errors.New("NOT FOUND")

const ParsedVAACollection = "parsedVaa"

// MongoRepository definitions.
type MongoRepository struct {
	db          *mongo.Database
	log         *zap.Logger
	collections struct {
		parsedVaa *mongo.Collection
	}
}

// NewRepository create a new respository instance.
func NewMongoRepository(db *mongo.Database, log *zap.Logger) *MongoRepository {
	return &MongoRepository{db, log, struct {
		parsedVaa *mongo.Collection
	}{
		parsedVaa: db.Collection(ParsedVAACollection),
	}}
}

// UpsertParsedVaa saves vaa information and parsed result.
func (s *MongoRepository) UpsertParsedVaa(ctx context.Context, parsedVAA ParsedVaaUpdate) error {
	update := bson.M{
		"$set":         parsedVAA,
		"$setOnInsert": indexedAt(*parsedVAA.UpdatedAt),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}

	opts := options.Update().SetUpsert(true)
	var err error
	_, err = s.collections.parsedVaa.UpdateByID(ctx, parsedVAA.ID, update, opts)
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
