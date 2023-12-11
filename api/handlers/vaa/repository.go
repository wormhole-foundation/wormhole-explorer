package vaa

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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
		parsedVaa          *mongo.Collection
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
			parsedVaa          *mongo.Collection
			vaasPythnet        *mongo.Collection
			invalidVaas        *mongo.Collection
			vaaCount           *mongo.Collection
			globalTransactions *mongo.Collection
		}{
			vaas:               db.Collection("vaas"),
			parsedVaa:          db.Collection("parsedVaa"),
			vaasPythnet:        db.Collection("vaasPythnet"),
			invalidVaas:        db.Collection("invalid_vaas"),
			vaaCount:           db.Collection("vaaCounts"),
			globalTransactions: db.Collection("globalTransactions"),
		},
	}
}

// FindVaasByTxHashWorkaround searches the database for VAAs that match a given transaction hash.
//
// This function exists to work around the issue that for Aptos and Solana, the real transaction
// hashes are stored in a different collection from other chains.
//
// When the `q.txHash` field is set, this function will look up transaction hashes in the `globalTransactions` collection.
// Then, if the transaction hash is not found, it will fall back to searching in the `vaas` collection.
//
// Take into consideration that multiple VAAs could share the same transaction ID.
func (r *Repository) FindVaasByTxHashWorkaround(
	ctx context.Context,
	query *VaaQuery,
) ([]*VaaDoc, error) {

	// Find globalTransactions that match the given TxHash
	cur, err := r.collections.globalTransactions.Find(
		ctx,
		bson.D{
			{"$or", bson.A{
				bson.D{{"originTx.nativeTxHash", bson.M{"$eq": query.txHash}}},
				bson.D{{"originTx.nativeTxHash", bson.M{"$eq": "0x" + query.txHash}}},
				bson.D{{"originTx.attribute.value.originTxHash", bson.M{"$eq": query.txHash}}},
				bson.D{{"originTx.attribute.value.originTxHash", bson.M{"$eq": "0x" + query.txHash}}},
			}},
		},
		nil,
	)

	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed to find globalTransactions by TxHash",
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

	// If no documents were found, look up the transaction hash in the `vaas` collection instead.
	if len(globalTxs) == 0 {
		return r.FindVaas(ctx, query)
	}

	// Find VAAs that match the given VAA ID
	q := *query // making a copy to avoid modifying the struct passed by the caller
	var ids []string
	for i := range globalTxs {
		ids = append(ids, globalTxs[i].ID)
	}
	q.SetIDs(ids)
	// Disable txHash filter, but keep all the other filters.
	// We have to do this because the transaction hashes in the `globalTransactions` collection
	// may be different that the transaction hash in the `vaas` collection. This is the case
	// for Aptos and Solana VAAs.
	q.txHash = ""
	return r.FindVaas(ctx, &q)
}

// FindVaasByEmitterAndToChain searches the database for VAAs that match a given emitter chain, address and toChain.
func (r *Repository) FindVaasByEmitterAndToChain(
	ctx context.Context,
	query *VaaQuery,
	toChain sdk.ChainID,
) ([]*VaaDoc, error) {

	// build a query pipeline based on input parameters
	var pipeline mongo.Pipeline
	{
		// filter by emitterChain, emitterAddr, and toChain
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{bson.E{"emitterChain", query.chainId}}},
		})
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{bson.E{"emitterAddr", query.emitter}}},
		})
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{bson.E{"rawStandardizedProperties.toChain", toChain}}},
		})

		// specify sorting criteria
		pipeline = append(pipeline, bson.D{{"$sort", bson.D{bson.E{"indexedAt", query.GetSortInt()}}}})

		// skip initial results
		if query.Pagination.Skip != 0 {
			pipeline = append(pipeline, bson.D{{"$skip", query.Pagination.Skip}})
		}

		// limit size of results
		pipeline = append(pipeline, bson.D{{"$limit", query.Pagination.Limit}})
	}

	// execute the aggregation pipeline
	cur, err := r.collections.parsedVaa.Aggregate(ctx, pipeline)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get vaa by emitter and toChain",
			zap.Error(err),
			zap.Any("q", query),
			zap.Any("toChain", toChain),
			zap.String("requestID", requestID),
		)
		return nil, errors.WithStack(err)
	}

	// read results from cursor
	var vaas []struct {
		ID string `bson:"_id"`
	}
	err = cur.All(ctx, &vaas)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed to decode cursor",
			zap.Error(err),
			zap.Any("q", query),
			zap.Any("toChain", toChain),
			zap.String("requestID", requestID),
		)
		return nil, errors.WithStack(err)
	}

	// if no results were found, return an empty slice instead of nil.
	if len(vaas) == 0 {
		return make([]*VaaDoc, 0), nil
	}

	// call FindVaas with the IDs we've found
	q := *query // make a copy to avoid modifying the struct passed by the caller
	for _, vaa := range vaas {
		q.ids = append(q.ids, vaa.ID)
	}
	return r.FindVaas(ctx, &q)
}

// FindVaas searches the database for VAAs matching the given filters.
//
// When the `q.txHash` field is set, this function will look up transaction hashes in the `vaas` collection.
//
// Take into consideration that multiple VAAs could share the same transaction ID.
func (r *Repository) FindVaas(
	ctx context.Context,
	q *VaaQuery,
) ([]*VaaDoc, error) {

	// build a query pipeline based on input parameters
	var pipeline mongo.Pipeline
	{
		// specify sorting criteria
		pipeline = append(pipeline, bson.D{
			{"$sort", bson.D{q.getSortPredicate()}},
		})

		// filter by VAA ids (potentially more than one)
		if len(q.ids) > 0 {
			var array bson.A
			for _, id := range q.ids {
				predicate := bson.D{bson.E{"_id", id}}
				array = append(array, predicate)
			}
			pipeline = append(pipeline, bson.D{
				{"$match", bson.D{{"$or", array}}},
			})
		}

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
				{"payload", bson.M{"$arrayElemAt": []interface{}{"$payload.parsedPayload", 0}}},
				{"appId", bson.M{"$arrayElemAt": []interface{}{"$payload.appId", 0}}},
			}},
		})

		// left outer join on the `globalTransaction` collection
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "globalTransactions"},
				{"localField", "_id"},
				{"foreignField", "_id"},
				{"as", "globalTransaction"},
			}},
		})

		// add globalTransaction fields
		pipeline = append(pipeline, bson.D{
			{"$addFields", bson.D{
				{"nativeTxHash", bson.M{"$arrayElemAt": []interface{}{"$globalTransaction.originTx.nativeTxHash", 0}}},
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
	if q.chainId == sdk.ChainIDPythNet {
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

	// If the payload field was not requested, remove it from the results.
	if !q.includeParsedPayload && q.appId == "" {
		for i := range vaasWithPayload {
			vaasWithPayload[i].Payload = nil
		}
	}

	// Set remaining fields on the returned structs
	for _, vaa := range vaasWithPayload {

		// For Solana and Aptos VAAs, overwrite the txHash found in the `vaas` collection
		// with the one from the `globalTransactions` collection.
		//
		// We have to do this because the value that comes from the gossip network is not
		// the real transaction hash in the case of those chains.
		//
		// If there is no transaction hash in the `globalTransactions` collection, we have no
		// option but to set the field to nil.
		if (vaa.EmitterChain == sdk.ChainIDSolana) || (vaa.EmitterChain == sdk.ChainIDAptos) {
			if vaa.NativeTxHash == "" {
				vaa.TxHash = nil
			} else {
				vaa.TxHash = &vaa.NativeTxHash
			}
		} else {
			if vaa.TxHash == nil && vaa.NativeTxHash != "" {
				txHash := utils.Remove0x(vaa.NativeTxHash)
				vaa.TxHash = &txHash
			}
		}

		// Set the `EmitterNativeAddr` field
		vaa.EmitterNativeAddr, err = domain.TranslateEmitterAddress(vaa.EmitterChain, vaa.EmitterAddr)
		if err != nil {
			r.logger.Warn("failed to translate emitter address for VAA",
				zap.Stringer("emitterChain", vaa.EmitterChain),
				zap.String("emitterAddr", vaa.EmitterAddr),
				zap.Error(err),
			)
		}
	}

	return vaasWithPayload, nil
}

// GetVaaCount get a count of vaa by chainID.
func (r *Repository) GetVaaCount(ctx context.Context, q *VaaQuery) ([]*VaaStats, error) {

	cur, err := r.collections.vaaCount.Find(ctx, bson.D{}, q.findOptions())
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
	ids                  []string
	chainId              sdk.ChainID
	emitter              string
	sequence             string
	txHash               string
	appId                string
	includeParsedPayload bool
}

// Query create a new VaaQuery with default pagination vaues.
func Query() *VaaQuery {
	p := pagination.Default()
	return &VaaQuery{Pagination: *p}
}

func (q *VaaQuery) SetIDs(ids []string) *VaaQuery {
	q.ids = ids
	return q
}

// SetChain set the chainId field of the VaaQuery struct.
func (q *VaaQuery) SetChain(chainID sdk.ChainID) *VaaQuery {
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

func (q *VaaQuery) IncludeParsedPayload(val bool) *VaaQuery {
	q.includeParsedPayload = val
	return q
}

func (q *VaaQuery) getSortPredicate() bson.E {
	return bson.E{"timestamp", q.GetSortInt()}
}

func (q *VaaQuery) findOptions() *options.FindOptions {

	sort := bson.D{q.getSortPredicate()}

	return options.
		Find().
		SetSort(sort).
		SetLimit(q.Limit).
		SetSkip(q.Skip)
}
