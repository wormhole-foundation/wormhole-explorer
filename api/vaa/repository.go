package vaa

import (
	"context"
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
		vaas        *mongo.Collection
		invalidVaas *mongo.Collection
	}
}

func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger: logger.With(zap.String("module", "VaaRepository")),
		collections: struct {
			vaas        *mongo.Collection
			invalidVaas *mongo.Collection
		}{vaas: db.Collection("vaas"), invalidVaas: db.Collection("invalid_vaas")}}
}

func (r *Repository) Find(ctx context.Context, q *VaaQuery) ([]*VaaDoc, error) {
	if q == nil {
		q = Query()
	}
	sort := bson.D{{q.SortBy, q.GetSortInt()}}
	cur, err := r.collections.vaas.Find(ctx, q.toBSON(), options.Find().SetLimit(q.PageSize).SetSkip(q.Offset).SetSort(sort))
	if err != nil {
		return nil, err
	}
	var vaas []*VaaDoc
	err = cur.All(ctx, &vaas)
	if err != nil {
		return nil, err
	}
	return vaas, err
}

func (r *Repository) FindOne(ctx context.Context, q *VaaQuery) (*VaaDoc, error) {
	var vaaDoc VaaDoc
	err := r.collections.vaas.FindOne(ctx, q.toBSON()).Decode(vaaDoc)
	if err != nil {
		return nil, err
	}
	return &vaaDoc, err
}

func (r *Repository) FindStats(ctx context.Context) ([]*VaaStats, error) {
	group := bson.D{
		{"$group", bson.D{
			{"_id", "$emitterChain"},
			{"Count", bson.D{{"$sum", 1}}},
		}},
	}
	c, err := r.collections.vaas.Aggregate(ctx, mongo.Pipeline{group})
	if err != nil {
		return nil, err
	}
	var stats []*VaaStats
	err = c.All(ctx, &stats)
	return stats, err
}

type VaaQuery struct {
	pagination.Pagination
	chainId  vaa.ChainID
	emitter  string
	sequence uint64
}

func Query() *VaaQuery {
	page := pagination.FirstPage()
	return &VaaQuery{Pagination: *page}
}

func (q *VaaQuery) SetChain(chainID vaa.ChainID) *VaaQuery {
	q.chainId = chainID
	return q
}

func (q *VaaQuery) SetEmitter(emitter string) *VaaQuery {
	q.emitter = emitter
	return q
}

func (q *VaaQuery) SetSequence(seq uint64) *VaaQuery {
	q.sequence = seq
	return q
}

func (q *VaaQuery) SetPagination(p *pagination.Pagination) *VaaQuery {
	q.Pagination = *p
	return q
}

func (q *VaaQuery) toBSON() *bson.D {
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
	return &r
}
