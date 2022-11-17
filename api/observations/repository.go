package observations

import (
	"context"
	"errors"
	"github.com/certusone/wormhole/node/pkg/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Repository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		observations *mongo.Collection
	}
}

func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger:      logger.With(zap.String("module", "ObservationsRepository")),
		collections: struct{ observations *mongo.Collection }{observations: db.Collection("observations")},
	}
}

func (r *Repository) Find(ctx context.Context, q *ObservationQuery) ([]*ObservationDoc, error) {
	if q == nil {
		q = Query()
	}
	sort := bson.D{{q.SortBy, q.GetSortInt()}}
	cur, err := r.collections.observations.Find(ctx, q.toBSON(), options.Find().SetLimit(q.PageSize).SetSkip(q.Offset).SetSort(sort))
	if err != nil {
		return nil, err
	}
	var obs []*ObservationDoc
	err = cur.All(ctx, &obs)
	if err != nil {
		return nil, err
	}
	return obs, err
}

var (
	ErrWrongQuery = errors.New("MALFORMED_QUERY")
)

func (r *Repository) FindOne(ctx context.Context, q *ObservationQuery) (*ObservationDoc, error) {
	if q == nil {
		return nil, ErrWrongQuery
	}
	var obs ObservationDoc
	err := r.collections.observations.FindOne(ctx, q.toBSON()).Decode(&obs)
	if err != nil {
		return nil, err
	}
	return &obs, err
}

type ObservationQuery struct {
	pagination.Pagination
	chainId      vaa.ChainID
	emitter      string
	sequence     uint64
	guardianAddr string
	hash         []byte
	uint64
}

func Query() *ObservationQuery {
	page := pagination.FirstPage()
	return &ObservationQuery{Pagination: *page}
}

func (q *ObservationQuery) SetChain(chainID vaa.ChainID) *ObservationQuery {
	q.chainId = chainID
	return q
}

func (q *ObservationQuery) SetEmitter(emitter string) *ObservationQuery {
	q.emitter = emitter
	return q
}

func (q *ObservationQuery) SetSequence(seq uint64) *ObservationQuery {
	q.sequence = seq
	return q
}

func (q *ObservationQuery) SetGuardianAddr(guardianAddr string) *ObservationQuery {
	q.guardianAddr = guardianAddr
	return q
}

func (q *ObservationQuery) SetHash(hash []byte) *ObservationQuery {
	q.hash = hash
	return q
}

func (q *ObservationQuery) SetPagination(p *pagination.Pagination) *ObservationQuery {
	q.Pagination = *p
	return q
}

func (q *ObservationQuery) toBSON() *bson.D {
	r := bson.D{}
	if q.chainId > 0 {
		r = append(r, bson.E{"emitterChain", q.chainId})
	}
	if q.emitter != "" {
		r = append(r, bson.E{"emitterAddr", q.emitter})
	}
	if q.sequence > 0 {
		r = append(r, bson.E{"sequence", q.sequence})
	}
	if len(q.hash) > 0 {
		r = append(r, bson.E{"hash", q.hash})
	}
	if q.guardianAddr != "" {
		r = append(r, bson.E{"guardianAddr", q.guardianAddr})
	}

	return &r
}
