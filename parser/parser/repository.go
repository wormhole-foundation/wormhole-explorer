package parser

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// repository errors
var ErrNotFound = errors.New("NOT FOUND")

// Repository definitions.
type Repository struct {
	db          *mongo.Database
	log         *zap.Logger
	collections struct {
		vaaParserFunctions *mongo.Collection
		parsedVaa          *mongo.Collection
	}
}

// NewRepository create a new respository.
func NewRepository(db *mongo.Database, log *zap.Logger) *Repository {
	return &Repository{db, log, struct {
		vaaParserFunctions *mongo.Collection
		parsedVaa          *mongo.Collection
	}{
		vaaParserFunctions: db.Collection("vaaParserFunctions"),
		parsedVaa:          db.Collection("parsedVaa"),
	}}
}

// GetVaaParserFunction get a vaa parser function by chainID and address.
func (r *Repository) GetVaaParserFunction(ctx context.Context, chainID uint16, address string) (*VaaParserFunctions, error) {
	filter := bson.D{bson.E{"emitterChain", chainID}, bson.E{"emitterAddress", address}}
	var vpf VaaParserFunctions
	err := r.collections.vaaParserFunctions.FindOne(ctx, filter).Decode(&vpf)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		r.log.Error("failed execute FindOne command to get vaaParserFunctions",
			zap.Error(err), zap.Uint16("chainID", uint16(chainID)), zap.String("address", address))
		return nil, err
	}
	return &vpf, nil
}

func (s *Repository) UpsertParsedVaa(ctx context.Context, e *queue.VaaEvent, r interface{}) error {
	now := time.Now()
	vaaDoc := ParsedVaaUpdate{
		ID:           e.ID(),
		EmitterChain: e.ChainID,
		EmitterAddr:  e.EmitterAddress,
		Sequence:     strconv.FormatUint(e.Sequence, 10),
		Result:       r,
		UpdatedAt:    &now,
	}

	update := bson.M{
		"$set":         vaaDoc,
		"$setOnInsert": indexedAt(now),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}

	opts := options.Update().SetUpsert(true)
	var err error
	_, err = s.collections.parsedVaa.UpdateByID(ctx, e.ID, update, opts)
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
