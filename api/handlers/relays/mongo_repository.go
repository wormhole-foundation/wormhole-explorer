package relays

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type MongoRepository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		relays *mongo.Collection
	}
}

func NewMongoRepository(db *mongo.Database, logger *zap.Logger) *MongoRepository {
	return &MongoRepository{db: db,
		logger: logger.With(zap.String("module", "MongoRelaysRepository")),
		collections: struct {
			relays *mongo.Collection
		}{
			relays: db.Collection("relays"),
		},
	}
}

func (r *MongoRepository) FindOne(ctx context.Context, q *RelaysQuery) (*RelayDoc, error) {
	var response RelayDoc
	err := r.collections.relays.FindOne(ctx, q.toBSON()).Decode(&response)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get relays",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	return &response, nil
}

func (q *RelaysQuery) toBSON() *bson.D {
	r := bson.D{}
	id := fmt.Sprintf("%d/%s/%s", q.chainId, q.emitter, q.sequence)
	r = append(r, bson.E{Key: "_id", Value: id})
	return &r
}
