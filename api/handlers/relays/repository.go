package relays

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Repository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		relays *mongo.Collection
	}
}

func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger: logger.With(zap.String("module", "VaaRepository")),
		collections: struct {
			relays *mongo.Collection
		}{
			relays: db.Collection("relays"),
		},
	}
}

func (r *Repository) FindOne(ctx context.Context, q *RelaysQuery) (*RelayResponse, error) {
	r.logger.Info("q.toBSON()", zap.Any("q", q.toBSON()))
	result := r.collections.relays.FindOne(ctx, q.toBSON())
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, errors.New("not found")
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get relays",
			zap.Error(result.Err()), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(result.Err())
	}

	var m bson.M
	err := result.Decode(&m)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to bson.M", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// to not expose the internal bson.M type
	response := make(RelayResponse)
	for k, v := range m {
		response[k] = v
	}

	return &response, nil
}

type RelaysQuery struct {
	chainId  vaa.ChainID
	emitter  string
	sequence string
}

type RelayResponse map[string]interface{}

func (q *RelaysQuery) toBSON() *bson.D {
	r := bson.D{}
	id := fmt.Sprintf("%d/%s/%s", q.chainId, q.emitter, q.sequence)
	r = append(r, bson.E{"id", id})
	return &r
}

func (q *RelaysQuery) SetChain(chainId vaa.ChainID) *RelaysQuery {
	q.chainId = chainId
	return q
}

func (q *RelaysQuery) SetEmitter(emitter string) *RelaysQuery {
	q.emitter = emitter
	return q
}

func (q *RelaysQuery) SetSequence(sequence string) *RelaysQuery {
	q.sequence = sequence
	return q
}

func Query() *RelaysQuery {
	return &RelaysQuery{}
}
