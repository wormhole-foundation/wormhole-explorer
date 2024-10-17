package heartbeats

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// MongoHearbeatRepository definition.
type MongoHearbeatRepository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		heartbeats *mongo.Collection
	}
}

// NewMongoRepository create a new MongoHearbeatRepository.
func NewMongoRepository(db *mongo.Database, logger *zap.Logger) *MongoHearbeatRepository {
	return &MongoHearbeatRepository{db: db,
		logger:      logger.With(zap.String("module", "HeartbeatsRepository")),
		collections: struct{ heartbeats *mongo.Collection }{heartbeats: db.Collection("heartbeats")},
	}
}

// FindByIDS get a list of HeartbeatDoc pointer.
func (r *MongoHearbeatRepository) FindByIDs(ctx context.Context, ids []string) ([]*HeartbeatDoc, error) {
	in := bson.M{"_id": bson.M{"$in": ids}}
	cur, err := r.collections.heartbeats.Find(ctx, in)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Find command to get heartbeats",
			zap.Error(err), zap.Strings("ids", ids), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	var heartbeats []*HeartbeatDoc
	err = cur.All(ctx, &heartbeats)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*HeartbeatDoc", zap.Error(err),
			zap.Strings("ids", ids), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	return heartbeats, err
}
