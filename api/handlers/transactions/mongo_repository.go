package transactions

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/common"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const pythEmitterAddr = "e101faedac5851e32b9b23b5f9411a8c2bac4aae3ed4dd7b811dd1a72ea4aa71"

type repositoryCollections struct {
	vaas               *mongo.Collection
	vaasPythnet        *mongo.Collection
	parsedVaa          *mongo.Collection
	globalTransactions *mongo.Collection
}

type MongoRepository struct {
	p2pNetwork  string
	db          *mongo.Database
	collections repositoryCollections
	logger      *zap.Logger
}

func NewMongoRepository(p2pNetwork string, db *mongo.Database, logger *zap.Logger) *MongoRepository {
	return &MongoRepository{
		p2pNetwork: p2pNetwork,
		db:         db,
		collections: repositoryCollections{
			vaas:               db.Collection("vaas"),
			vaasPythnet:        db.Collection("vaasPythnet"),
			parsedVaa:          db.Collection("parsedVaa"),
			globalTransactions: db.Collection("globalTransactions"),
		}, logger: logger,
	}
}

// ListTransactionsByAddress returns a sorted list of transactions for a given address.
//
// Pagination is implemented using a keyset cursor pattern, based on the (timestamp, ID) pair.
func (r *MongoRepository) ListTransactionsByAddress(
	ctx context.Context,
	address string,
	pagination *pagination.Pagination,
) ([]TransactionDto, error) {

	ids, err := common.FindVaasIdsByFromAddressOrToAddress(ctx, r.db, address)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []TransactionDto{}, nil
	}

	var pipeline mongo.Pipeline

	// filter by ids
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}}})

	// inner join on the `parsedVaa` collection
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "parsedVaa"},
		{Key: "localField", Value: "_id"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "parsedVaa"},
	}}})
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "parsedVaa", Value: bson.D{{Key: "$ne", Value: []any{}}}}}}})

	// sort by timestamp
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{bson.E{Key: "timestamp", Value: pagination.GetSortInt()}}}})

	// Skip initial results
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: pagination.Skip}})

	// Limit size of results
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: pagination.Limit}})

	// left outer join on the `transferPrices` collection
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "transferPrices"},
		{Key: "localField", Value: "_id"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "transferPrices"},
	}}})

	// left outer join on the `vaaIdTxHash` collection
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "vaaIdTxHash"},
		{Key: "localField", Value: "_id"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "vaaIdTxHash"},
	}}})

	// left outer join on the `globalTransactions` collection
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: "globalTransactions"},
		{Key: "localField", Value: "_id"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "globalTransactions"},
	}}})

	// add nested fields
	pipeline = append(pipeline, bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "txHash", Value: bson.M{"$arrayElemAt": []interface{}{"$vaaIdTxHash.txHash", 0}}},
			{Key: "payload", Value: bson.M{"$arrayElemAt": []interface{}{"$parsedVaa.parsedPayload", 0}}},
			{Key: "standardizedProperties", Value: bson.M{"$arrayElemAt": []interface{}{"$parsedVaa.standardizedProperties", 0}}},
			{Key: "symbol", Value: bson.M{"$arrayElemAt": []interface{}{"$transferPrices.symbol", 0}}},
			{Key: "usdAmount", Value: bson.M{"$arrayElemAt": []interface{}{"$transferPrices.usdAmount", 0}}},
			{Key: "tokenAmount", Value: bson.M{"$arrayElemAt": []interface{}{"$transferPrices.tokenAmount", 0}}},
		}},
	})

	// Execute the aggregation pipeline
	cur, err := r.collections.vaas.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return nil, err
	}

	// Read results from cursor
	var documents []TransactionDto
	err = cur.All(ctx, &documents)
	if err != nil {
		r.logger.Error("failed to decode cursor", zap.Error(err))
		return nil, err
	}

	return documents, nil
}

// FindTransactions returns transactions matching a specified search criteria.
func (r *MongoRepository) FindTransactions(
	ctx context.Context,
	input *FindTransactionsInput,
) ([]TransactionDto, error) {

	// Build the aggregation pipeline
	var pipeline mongo.Pipeline
	{
		// Specify sorting criteria
		if input.sort {
			pipeline = append(pipeline, bson.D{
				{"$sort", bson.D{
					bson.E{"timestamp", input.pagination.GetSortInt()},
					bson.E{"_id", -1},
				}},
			})
		}

		// Filter by ID
		if input.id != "" {
			pipeline = append(pipeline, bson.D{
				{"$match", bson.D{{"_id", input.id}}},
			})
		}

		// left outer join on the `transferPrices` collection
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "transferPrices"},
				{"localField", "_id"},
				{"foreignField", "_id"},
				{"as", "transferPrices"},
			}},
		})

		// left outer join on the `vaaIdTxHash` collection
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "vaaIdTxHash"},
				{"localField", "_id"},
				{"foreignField", "_id"},
				{"as", "vaaIdTxHash"},
			}},
		})

		// left outer join on the `parsedVaa` collection
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "parsedVaa"},
				{"localField", "_id"},
				{"foreignField", "_id"},
				{"as", "parsedVaa"},
			}},
		})

		// left outer join on the `globalTransactions` collection
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "globalTransactions"},
				{"localField", "_id"},
				{"foreignField", "_id"},
				{"as", "globalTransactions"},
			}},
		})

		// add nested fields
		pipeline = append(pipeline, bson.D{
			{"$addFields", bson.D{
				{"txHash", bson.M{"$arrayElemAt": []interface{}{"$vaaIdTxHash.txHash", 0}}},
				{"payload", bson.M{"$arrayElemAt": []interface{}{"$parsedVaa.parsedPayload", 0}}},
				{"standardizedProperties", bson.M{"$arrayElemAt": []interface{}{"$parsedVaa.standardizedProperties", 0}}},
				{"symbol", bson.M{"$arrayElemAt": []interface{}{"$transferPrices.symbol", 0}}},
				{"usdAmount", bson.M{"$arrayElemAt": []interface{}{"$transferPrices.usdAmount", 0}}},
				{"tokenAmount", bson.M{"$arrayElemAt": []interface{}{"$transferPrices.tokenAmount", 0}}},
			}},
		})

		// Unset unused fields
		pipeline = append(pipeline, bson.D{
			{"$unset", []interface{}{"transferPrices", "vaaTxIdHash", "parsedVaa"}},
		})

		// Skip initial results
		if input.pagination != nil {
			pipeline = append(pipeline, bson.D{
				{"$skip", input.pagination.Skip},
			})
		}

		// Limit size of results
		if input.pagination != nil {
			pipeline = append(pipeline, bson.D{
				{"$limit", input.pagination.Limit},
			})
		}
	}

	// Execute the aggregation pipeline
	cur, err := r.collections.vaas.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return nil, err
	}

	// Read results from cursor
	var documents []TransactionDto
	err = cur.All(ctx, &documents)
	if err != nil {
		r.logger.Error("failed to decode cursor", zap.Error(err))
		return nil, err
	}

	return documents, nil
}

// getTotalPythMessage returns the last sequence for the pyth emitter address
func (r *MongoRepository) getTotalPythMessage(ctx context.Context) (string, error) {
	if r.p2pNetwork != config.P2pMainNet {
		return "0", nil

	}

	var vaaPyth struct {
		ID       string `bson:"_id"`
		Sequence string `bson:"sequence"`
	}

	filter := bson.M{"emitterAddr": pythEmitterAddr}
	options := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	err := r.collections.vaasPythnet.FindOne(ctx, filter, options).Decode(&vaaPyth)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Warn("no pyth message found")
			return "0", nil
		}
		r.logger.Error("failed to get pyth message", zap.String("emitterAddr", pythEmitterAddr), zap.Error(err))
		return "", err
	}
	return vaaPyth.Sequence, nil
}

// findOriginTxFromVaa uses data from the `vaas` collection to create an `OriginTx`.
func (r *MongoRepository) findOriginTxFromVaa(ctx context.Context, q *GlobalTransactionQuery) (*OriginTx, error) {

	// query the `vaas` collection
	var record struct {
		Timestamp    time.Time   `bson:"timestamp"`
		TxHash       string      `bson:"txHash"`
		EmitterChain sdk.ChainID `bson:"emitterChain"`
	}
	err := r.db.
		Collection("vaas").
		FindOne(ctx, bson.M{"_id": q.id}).
		Decode(&record)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get global transaction from `vaas` collection",
			zap.Error(err),
			zap.Any("q", q),
			zap.String("requestID", requestID),
		)
		return nil, errors.WithStack(err)
	}

	// populate the result and return
	originTx := OriginTx{
		Status: string(domain.SourceTxStatusConfirmed),
	}
	if record.EmitterChain != sdk.ChainIDSolana && record.EmitterChain != sdk.ChainIDAptos {
		originTx.TxHash = record.TxHash
	}
	return &originTx, nil
}

func (r *MongoRepository) FindGlobalTransactionByID(ctx context.Context, q *GlobalTransactionQuery) (*GlobalTransactionDoc, error) {

	// Look up the global transaction
	globalTransaction, err := r.findGlobalTransactionByID(ctx, q)
	if err != nil && err != errs.ErrNotFound {
		return nil, fmt.Errorf("failed to find global transaction by id: %w", err)
	}

	// Look up the VAA
	originTx, err := r.findOriginTxFromVaa(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to find origin tx from the `vaas` collection: %w", err)
	}

	// If we found data in the `globalTransactions` collections, use it.
	// Otherwise, we can use data from the VAA collection to create an `OriginTx` object.
	//
	// Usually, `OriginTx`s will only exist in the `globalTransactions` collection for Solana,
	// which is gathered by the `tx-tracker` service.
	// For all the other chains, we'll end up using the data found in the `vaas` collection.
	var result *GlobalTransactionDoc
	switch {
	case globalTransaction == nil:
		result = &GlobalTransactionDoc{
			ID:       q.id,
			OriginTx: originTx,
		}
	case globalTransaction != nil && globalTransaction.OriginTx == nil:
		result = &GlobalTransactionDoc{
			ID:            q.id,
			OriginTx:      originTx,
			DestinationTx: globalTransaction.DestinationTx,
		}
	default:
		result = globalTransaction
	}

	return result, nil

}

// findGlobalTransactionByID searches the `globalTransactions` collection by ID.
func (r *MongoRepository) findGlobalTransactionByID(ctx context.Context, q *GlobalTransactionQuery) (*GlobalTransactionDoc, error) {

	var globalTranstaction GlobalTransactionDoc
	err := r.db.
		Collection("globalTransactions").
		FindOne(ctx, bson.M{"_id": q.id}).
		Decode(&globalTranstaction)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get global transaction from `globalTransactions` collection",
			zap.Error(err),
			zap.Any("q", q),
			zap.String("requestID", requestID),
		)
		return nil, errors.WithStack(err)
	}

	return &globalTranstaction, nil
}
