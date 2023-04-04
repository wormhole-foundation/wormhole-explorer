package vaa

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
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
		vaas               *mongo.Collection
		vaasPythnet        *mongo.Collection
		invalidVaas        *mongo.Collection
		vaaCount           *mongo.Collection
		globalTransactions *mongo.Collection
	}
}

// NewRepository create a new Repository.
func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger: logger.With(zap.String("module", "VaaRepository")),
		collections: struct {
			vaas               *mongo.Collection
			vaasPythnet        *mongo.Collection
			invalidVaas        *mongo.Collection
			vaaCount           *mongo.Collection
			globalTransactions *mongo.Collection
		}{
			vaas:               db.Collection("vaas"),
			vaasPythnet:        db.Collection("vaasPythnet"),
			invalidVaas:        db.Collection("invalid_vaas"),
			vaaCount:           db.Collection("vaaCounts"),
			globalTransactions: db.Collection("globalTransactions"),
		},
	}
}

// Find searches the database for VAAs.
// The input parameter [q *VaaQuery] define the filters to apply in the query.
func (r *Repository) Find(ctx context.Context, q *VaaQuery) ([]*VaaDoc, error) {

	var err error
	var cur *mongo.Cursor
	if q.chainId == vaa.ChainIDPythNet {
		cur, err = r.collections.vaasPythnet.Find(ctx, q.toBSON(), q.findOptions())
	} else {
		cur, err = r.collections.vaas.Find(ctx, q.toBSON(), q.findOptions())
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

	// Clear the `Payload` and `AppId` fields.
	// (this function doesn't return those fields currently, but may do so in the future by adding new parameters)
	for i := range vaas {
		vaas[i].Payload = nil
		vaas[i].AppId = ""
	}

	// If no results were found, return an empty slice instead of nil.
	if vaas == nil {
		vaas = make([]*VaaDoc, 0)
	}

	return vaas, err
}

// FindVaasBySolanaTxHash searches the database for VAAs that match a given Solana transaction hash.
func (r *Repository) FindVaasBySolanaTxHash(
	ctx context.Context,
	txHash string,
	includeParsedPayload bool,
) ([]*VaaDoc, error) {

	r.logger.Info("FindVaasBySolanaTxHash",
		zap.String("txHash", txHash),
		zap.Bool("includeParsedPayload", includeParsedPayload),
	)

	// Find globalTransactions that match the given Solana TxHash
	cur, err := r.collections.globalTransactions.Find(
		ctx,
		bson.D{bson.E{"originTx.nativeTxHash", txHash}},
		nil,
	)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed to find globalTransactions by Solana TxHash",
			zap.Error(err),
			zap.String("requestID", requestID),
		)
		return nil, errors.WithStack(err)
	}

	// Read results from cursor
	var globalTxs []transactions.GlobalTransactionDoc
	err = cur.All(ctx, &globalTxs)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed to decode cursor to []GlobalTransactionDoc",
			zap.Error(err),
			zap.String("requestID", requestID),
		)
		return nil, errors.WithStack(err)
	}

	r.logger.Info("Results",
		zap.Any("globalTxs", globalTxs),
	)

	// If no results were found, return an empty slice instead of nil.
	if len(globalTxs) == 0 {
		result := make([]*VaaDoc, 0)
		return result, nil
	}
	if len(globalTxs) > 1 {
		return nil, fmt.Errorf("expected at most one transaction, but found %d", len(globalTxs))
	}

	// Find VAAs that match the given VAA ID
	q := Query().SetID(globalTxs[0].ID)
	var vaas []*VaaDoc
	if includeParsedPayload {
		vaas, err = r.FindVaasWithPayload(ctx, q)
	} else {
		vaas, err = r.Find(ctx, q)
	}

	r.logger.Info("Results",
		zap.Any("vaas", vaas),
	)

	return vaas, err
}

// FindVaasWithPayload returns VAAs that include a parsed payload.
// The input parameter `q` defines the filters to be applied in the query.
func (r *Repository) FindVaasWithPayload(
	ctx context.Context,
	q *VaaQuery,
) ([]*VaaDoc, error) {

	// build a query pipeline based on input parameters
	var pipeline mongo.Pipeline
	{
		// specify sorting criteria
		pipeline = append(pipeline, bson.D{
			{"$sort", bson.D{bson.E{q.SortBy, q.GetSortInt()}}},
		})

		// filter by emitterChain
		if q.chainId != 0 {
			pipeline = append(pipeline, bson.D{
				{"$match", bson.D{bson.E{"emitterChain", q.chainId}}},
			})
		}

		// filter by emitterAddr
		if q.emitter != "" {
			pipeline = append(pipeline, bson.D{
				{"$match", bson.D{bson.E{"emitterAddr", q.emitter}}},
			})
		}

		// filter by sequence
		if q.sequence != "" {
			pipeline = append(pipeline, bson.D{
				{"$match", bson.D{bson.E{"sequence", q.sequence}}},
			})
		}

		// filter by txHash
		if q.txHash != "" {
			pipeline = append(pipeline, bson.D{
				{"$match", bson.D{bson.E{"txHash", q.txHash}}},
			})
		}

		// left outer join on the `parsedVaa` collection
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "parsedVaa"},
				{"localField", "_id"},
				{"foreignField", "_id"},
				{"as", "payload"},
			}},
		})

		// add parsed payload fields
		pipeline = append(pipeline, bson.D{
			{"$addFields", bson.D{
				{"payload", bson.M{"$arrayElemAt": []interface{}{"$payload.result", 0}}},
				{"appId", bson.M{"$arrayElemAt": []interface{}{"$payload.appId", 0}}},
			}},
		})

		// filter by appId
		if q.appId != "" {
			pipeline = append(pipeline, bson.D{
				{"$match", bson.D{bson.E{"appId", q.appId}}},
			})
		}

		// skip initial results
		if q.Pagination.Skip != 0 {
			pipeline = append(pipeline, bson.D{
				{"$skip", q.Pagination.Skip},
			})
		}

		// limit size of results
		pipeline = append(pipeline, bson.D{
			{"$limit", q.Pagination.Limit},
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
	var vaasWithPayload []*VaaDoc
	err = cur.All(ctx, &vaasWithPayload)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*VaaDoc",
			zap.Error(err),
			zap.Any("q", q),
			zap.String("requestID", requestID),
		)
		return nil, errors.WithStack(err)
	}

	// If no results were found, return an empty slice instead of nil.
	if vaasWithPayload == nil {
		vaasWithPayload = make([]*VaaDoc, 0)
	}

	return vaasWithPayload, nil
}

// GetVaaCount get a count of vaa by chainID.
func (r *Repository) GetVaaCount(ctx context.Context, q *VaaQuery) ([]*VaaStats, error) {

	cur, err := r.collections.vaaCount.Find(ctx, q.toBSON(), q.findOptions())
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
	id       string
	chainId  vaa.ChainID
	emitter  string
	sequence string
	txHash   string
	appId    string
}

// Query create a new VaaQuery with default pagination vaues.
func Query() *VaaQuery {
	p := pagination.Default()
	return &VaaQuery{Pagination: *p}
}

// SetChain sets the id field of the VaaQuery struct.
func (q *VaaQuery) SetID(id string) *VaaQuery {
	q.id = id
	return q
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

func (q *VaaQuery) SetAppId(appId string) *VaaQuery {
	q.appId = appId
	return q
}

func (q *VaaQuery) toBSON() *bson.D {
	r := bson.D{}
	if q.id != "" {
		r = append(r, bson.E{"_id", q.id})
	}
	if q.chainId > 0 {
		r = append(r, bson.E{"emitterChain", q.chainId})
	}
	if q.emitter != "" {
		r = append(r, bson.E{"emitterAddr", q.emitter})
	}
	if q.sequence != "" {
		r = append(r, bson.E{"sequence", q.sequence})
	}
	if q.txHash != "" {
		r = append(r, bson.E{"txHash", q.txHash})
	}
	return &r
}

func (q *VaaQuery) findOptions() *options.FindOptions {

	sort := bson.D{{q.SortBy, q.GetSortInt()}}

	return options.
		Find().
		SetSort(sort).
		SetLimit(q.Limit).
		SetSkip(q.Skip)
}
