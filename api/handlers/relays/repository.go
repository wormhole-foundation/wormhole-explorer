package relays

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
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
	response := make(RelayResponse)
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
