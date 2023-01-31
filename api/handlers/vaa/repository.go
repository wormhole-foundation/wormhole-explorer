package vaa

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Repository definition
type Repository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		vaas        *mongo.Collection
		vaasPythnet *mongo.Collection
		invalidVaas *mongo.Collection
		vaaCount    *mongo.Collection
	}
}

// NewRepository create a new Repository.
func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger: logger.With(zap.String("module", "VaaRepository")),
		collections: struct {
			vaas        *mongo.Collection
			vaasPythnet *mongo.Collection
			invalidVaas *mongo.Collection
			vaaCount    *mongo.Collection
		}{vaas: db.Collection("vaas"), vaasPythnet: db.Collection("vaasPythnet"), invalidVaas: db.Collection("invalid_vaas"),
			vaaCount: db.Collection("vaaCounts")}}
}

// Find get a list of *VaaDoc.
// The input parameter [q *VaaQuery] define the filters to apply in the query.
func (r *Repository) Find(ctx context.Context, q *VaaQuery) ([]*VaaDoc, error) {
	var err error
	var cur *mongo.Cursor
	if q == nil {
		q = Query()
	}
	sort := bson.D{{
		Key:   q.SortBy,
		Value: q.GetSortInt(),
	}}
	if q.chainId == vaa.ChainIDPythNet {
		cur, err = r.collections.vaasPythnet.Find(ctx, q.toBSON(), options.Find().SetLimit(q.PageSize).SetSkip(q.Offset).SetSort(sort))
	} else {
		cur, err = r.collections.vaas.Find(ctx, q.toBSON(), options.Find().SetLimit(q.PageSize).SetSkip(q.Offset).SetSort(sort))
	}
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Find command to get vaas",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	var vaas []*VaaDoc
	err = cur.All(ctx, &vaas)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*VaaDoc", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	return vaas, err
}

// FindOne get *VaaDoc.
// The input parameter [q *VaaQuery] define the filters to apply in the query.
func (r *Repository) FindOne(ctx context.Context, q *VaaQuery) (*VaaDoc, error) {
	var vaaDoc VaaDoc
	var err error
	if q.chainId == vaa.ChainIDPythNet {
		err = r.collections.vaasPythnet.FindOne(ctx, q.toBSON()).Decode(&vaaDoc)
	} else {
		err = r.collections.vaas.FindOne(ctx, q.toBSON()).Decode(&vaaDoc)
	}
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get vaas",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	return &vaaDoc, err
}

// FindVaasWithPayload returns VAAs that include a parsed payload.
// The input parameter `q` defines the filters to be applied in the query.
func (r *Repository) FindVaasWithPayload(
	ctx context.Context,
	q *VaaQuery,
) ([]*VaaWithPayload, error) {

	// build a query pipeline based on input parameters
	var pipeline mongo.Pipeline
	{
		// filter by emitterChain
		if q.chainId != 0 {
			pipeline = append(pipeline, bson.D{
				{Key: "$match", Value: bson.D{bson.E{Key: "emitterChain", Value: q.chainId}}},
			})
		}

		// filter by emitterAddr
		if q.emitter != "" {
			pipeline = append(pipeline, bson.D{
				{Key: "$match", Value: bson.D{bson.E{Key: "emitterAddr", Value: q.emitter}}},
			})
		}

		// filter by sequence
		if q.sequence != "" {
			pipeline = append(pipeline, bson.D{
				{Key: "$match", Value: bson.D{bson.E{Key: "sequence", Value: q.sequence}}},
			})
		}

		// left outer join on the `parsedVaa` collection
		pipeline = append(pipeline, bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "parsedVaa"},
				{Key: "localField", Value: "_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "payload"},
			}},
		})

		// add parsed payload fields
		pipeline = append(pipeline, bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "payload", Value: bson.M{
					"$arrayElemAt": []interface{}{"$payload.result", 0},
				}},
			}},
		})

		// limit size of results
		pipeline = append(pipeline, bson.D{
			{Key: "$limit", Value: q.Pagination.PageSize},
		})
	}

	// execute the aggregation pipeline
	var err error
	var cur *mongo.Cursor
	if q.chainId == vaa.ChainIDPythNet {
		cur, err = r.collections.vaasPythnet.Aggregate(ctx, pipeline)
	} else {
		cur, err = r.collections.vaas.Aggregate(ctx, pipeline)
	}
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get vaa with payload",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// read results from cursor
	var vaasWithPayload []*VaaWithPayload
	err = cur.All(ctx, &vaasWithPayload)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*VaaWithPayload",
			zap.Error(err),
			zap.Any("q", q),
			zap.String("requestID", requestID),
		)
		return nil, errors.WithStack(err)
	}

	return vaasWithPayload, nil
}

// GetVaaCount get a count of vaa by chainID.
func (r *Repository) GetVaaCount(ctx context.Context, q *VaaQuery) ([]*VaaStats, error) {
	if q == nil {
		q = Query()
	}
	sort := bson.D{{
		Key:   q.SortBy,
		Value: q.GetSortInt(),
	}}
	cur, err := r.collections.vaaCount.Find(ctx, q.toBSON(), options.Find().SetLimit(q.PageSize).SetSkip(q.Offset).SetSort(sort))
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Find command to get vaaCount",
			zap.Error(err), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	var varCounts []*VaaStats
	err = cur.All(ctx, &varCounts)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*VaaStats", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	return varCounts, nil
}

// VaaQuery respresent a query for the vaa mongodb document.
type VaaQuery struct {
	pagination.Pagination
	chainId  vaa.ChainID
	emitter  string
	sequence string
	txHash   string
}

// Query create a new VaaQuery with default pagination vaues.
func Query() *VaaQuery {
	page := pagination.FirstPage()
	return &VaaQuery{Pagination: *page}
}

// SetChain set the chainId field of the VaaQuery struct.
func (q *VaaQuery) SetChain(chainID vaa.ChainID) *VaaQuery {
	q.chainId = chainID
	return q
}

// SetEmitter set the emitter field of the VaaQuery struct.
func (q *VaaQuery) SetEmitter(emitter string) *VaaQuery {
	q.emitter = emitter
	return q
}

// SetSequence set the sequence field of the VaaQuery struct.
func (q *VaaQuery) SetSequence(seq string) *VaaQuery {
	q.sequence = seq
	return q
}

// SetPagination set the pagination field of the VaaQuery struct.
func (q *VaaQuery) SetPagination(p *pagination.Pagination) *VaaQuery {
	q.Pagination = *p
	return q
}

// SetTxHash set the txHash field of the VaaQuery struct.
func (q *VaaQuery) SetTxHash(txHash string) *VaaQuery {
	q.txHash = txHash
	return q
}

func (q *VaaQuery) toBSON() *bson.D {
	r := bson.D{}
	if q.chainId > 0 {
		r = append(r, bson.E{Key: "emitterChain", Value: q.chainId})
	}
	if q.emitter != "" {
		r = append(r, bson.E{Key: "emitterAddr", Value: q.emitter})
	}
	if q.sequence != "" {
		r = append(r, bson.E{Key: "sequence", Value: q.sequence})
	}
	if q.txHash != "" {
		r = append(r, bson.E{Key: "txHash", Value: q.txHash})
	}
	return &r
}
