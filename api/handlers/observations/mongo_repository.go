// Package observations handle the request of observations data from governor endpoint defined in the api.
package observations

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.uber.org/zap"
)

// MongoRepository definition.
type MongoRepository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		observations *mongo.Collection
	}
}

// NewMongoRepository create a new Repository.
func NewMongoRepository(db *mongo.Database, logger *zap.Logger) *MongoRepository {
	return &MongoRepository{db: db,
		logger:      logger.With(zap.String("module", "MongoObservationsRepository")),
		collections: struct{ observations *mongo.Collection }{observations: db.Collection("observations")},
	}
}

// Find get a list of ObservationDoc pointers.
// The input parameter [q *ObservationQuery] define the filters to apply in the query.
func (r *MongoRepository) Find(ctx context.Context, q *ObservationQuery) ([]*ObservationDoc, error) {

	// Sort observations in descending timestamp order
	sort := bson.D{{"indexedAt", -1}}

	cur, err := r.collections.observations.Find(ctx, q.toBSON(), options.Find().SetLimit(q.Limit).SetSkip(q.Skip).SetSort(sort))
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Find command to get observations",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	var obs []*ObservationDoc
	err = cur.All(ctx, &obs)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*ObservationDoc", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// If no results were found, return an empty slice instead of nil.
	if obs == nil {
		obs = make([]*ObservationDoc, 0)
	}

	return obs, err
}

// Find get ObservationDoc pointer.
// The input parameter [q *ObservationQuery] define the filters to apply in the query.
func (r *MongoRepository) FindOne(ctx context.Context, q *ObservationQuery) (*ObservationDoc, error) {
	var obs ObservationDoc
	err := r.collections.observations.FindOne(ctx, q.toBSON()).Decode(&obs)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get observations",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	return &obs, err
}

func (q *ObservationQuery) toBSON() *bson.D {
	r := bson.D{}
	if q.chainId > 0 {
		r = append(r, bson.E{"emitterChain", q.chainId})
	}
	if q.emitter != "" {
		r = append(r, bson.E{"emitterAddr", q.emitter})
	}
	if q.sequence != "" {
		r = append(r, bson.E{"sequence", q.sequence})
	}
	if len(q.hash) > 0 {
		r = append(r, bson.E{"hash", q.hash})
	}
	if q.guardianAddr != "" {
		r = append(r, bson.E{"guardianAddr", q.guardianAddr})
	}
	if q.txHash != nil {
		nativeTxHash := q.txHash.String()
		r = append(r, bson.E{"nativeTxHash", nativeTxHash})
	}

	return &r
}
