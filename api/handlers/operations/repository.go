package operations

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Repository definition
type Repository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		vaas               *mongo.Collection
		parsedVaa          *mongo.Collection
		globalTransactions *mongo.Collection
	}
}

// NewRepository create a new Repository.
func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger: logger.With(zap.String("module", "OperationRepository")),
		collections: struct {
			vaas               *mongo.Collection
			parsedVaa          *mongo.Collection
			globalTransactions *mongo.Collection
		}{
			vaas:               db.Collection("vaas"),
			parsedVaa:          db.Collection("parsedVaa"),
			globalTransactions: db.Collection("globalTransactions"),
		},
	}
}

// FindById returns the operations for the given chainID/emitter/seq.
func (r *Repository) FindById(ctx context.Context, id string) (*OperationDto, error) {

	var pipeline mongo.Pipeline

	// filter vaas by id
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: id}}}})

	// lookup vaas
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "vaas"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "vaas"}}}})

	// lookup globalTransactions
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "globalTransactions"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "globalTransactions"}}}})

	// lookup transferPrices
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "transferPrices"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "transferPrices"}}}})

	// lookup parsedVaa
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "parsedVaa"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "parsedVaa"}}}})

	// add fields
	pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.D{
		{Key: "payload", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$parsedVaa.parsedPayload", 0}}}},
		{Key: "vaa", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$vaas", 0}}}},
		{Key: "standardizedProperties", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$parsedVaa.standardizedProperties", 0}}}},
		{Key: "symbol", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$transferPrices.symbol", 0}}}},
		{Key: "usdAmount", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$transferPrices.usdAmount", 0}}}},
		{Key: "tokenAmount", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$transferPrices.tokenAmount", 0}}}},
	}}})

	// unset
	pipeline = append(pipeline, bson.D{{Key: "$unset", Value: bson.A{"transferPrices", "parsedVaa"}}})

	// Execute the aggregation pipeline
	cur, err := r.collections.globalTransactions.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return nil, err
	}

	// Read results from cursor
	var operations []*OperationDto
	err = cur.All(ctx, &operations)
	if err != nil {
		r.logger.Error("failed to decode cursor", zap.Error(err))
		return nil, err
	}

	// Check if there is only one operation
	if len(operations) > 1 {
		r.logger.Error("invalid number of operations", zap.Int("count", len(operations)))
		return nil, fmt.Errorf("invalid number of operations")
	}

	if len(operations) == 0 {
		return nil, errors.ErrNotFound
	}

	return operations[0], nil
}

type mongoID struct {
	Id string `bson:"_id"`
}

// findOperationsIdByAddressOrTxHash returns all operations filtered by address or txHash.
func findOperationsIdByAddressOrTxHash(ctx context.Context, db *mongo.Database, q string, pagination *pagination.Pagination) ([]string, error) {
	qHexa := strings.ToLower(q)
	if !utils.StartsWith0x(q) {
		qHexa = "0x" + strings.ToLower(qHexa)
	}

	matchGlobalTransactions := bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "originTx.from", Value: bson.M{"$eq": qHexa}}},
		bson.D{{Key: "originTx.from", Value: bson.M{"$eq": q}}},
		bson.D{{Key: "originTx.nativeTxHash", Value: bson.M{"$eq": qHexa}}},
		bson.D{{Key: "originTx.nativeTxHash", Value: bson.M{"$eq": q}}},
		bson.D{{Key: "originTx.attribute.value.originTxHash", Value: bson.M{"$eq": qHexa}}},
		bson.D{{Key: "originTx.attribute.value.originTxHash", Value: bson.M{"$eq": q}}},
		bson.D{{Key: "destinationTx.txHash", Value: bson.M{"$eq": qHexa}}},
		bson.D{{Key: "destinationTx.txHash", Value: bson.M{"$eq": q}}},
	}}}}}

	matchParsedVaa := bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "standardizedProperties.toAddress", Value: bson.M{"$eq": qHexa}}},
		bson.D{{Key: "standardizedProperties.toAddress", Value: bson.M{"$eq": q}}},
	}}}}}

	globalTransactionFilter := bson.D{{Key: "$unionWith", Value: bson.D{{Key: "coll", Value: "globalTransactions"}, {Key: "pipeline", Value: bson.A{matchGlobalTransactions}}}}}
	parserFilter := bson.D{{Key: "$unionWith", Value: bson.D{{Key: "coll", Value: "parsedVaa"}, {Key: "pipeline", Value: bson.A{matchParsedVaa}}}}}
	group := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$_id"}}}}
	pipeline := []bson.D{globalTransactionFilter, parserFilter, group}

	cur, err := db.Collection("_operationsTemporal").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var documents []mongoID
	err = cur.All(ctx, &documents)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, doc := range documents {
		ids = append(ids, doc.Id)
	}
	return ids, nil
}

// QueryFilterIsVaaID checks if q is a vaaID.
func QueryFilterIsVaaID(ctx context.Context, q string) []string {
	// check if q is a vaaID
	isVaaID := regexp.MustCompile(`\d+/\w+/\d+`).MatchString(q)
	if isVaaID {
		return []string{q}
	}
	return []string{}
}

// FindAll returns all operations filtered by q.
func (r *Repository) FindAll(ctx context.Context, q string, pagination *pagination.Pagination) ([]*OperationDto, error) {

	var pipeline mongo.Pipeline

	// get all ids by that match q
	if q != "" {
		var ids []string
		// find all ids that match q (vaaID)
		ids = QueryFilterIsVaaID(ctx, q)
		if len(ids) == 0 {
			// find all ids that match q (address or txHash)
			var err error
			ids, err = findOperationsIdByAddressOrTxHash(ctx, r.db, q, pagination)
			if err != nil {
				return nil, err
			}

			if len(ids) == 0 {
				return []*OperationDto{}, nil
			}
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}}})
	}

	// sort
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{bson.E{Key: "originTx.timestamp", Value: pagination.GetSortInt()}}}})

	// Skip initial results
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: pagination.Skip}})

	// Limit size of results
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: pagination.Limit}})

	// lookup vaas
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "vaas"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "vaas"}}}})

	// lookup globalTransactions
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "globalTransactions"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "globalTransactions"}}}})

	// lookup transferPrices
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "transferPrices"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "transferPrices"}}}})

	// lookup parsedVaa
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "parsedVaa"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "parsedVaa"}}}})

	// add fields
	pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.D{
		{Key: "payload", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$parsedVaa.parsedPayload", 0}}}},
		{Key: "vaa", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$vaas", 0}}}},
		{Key: "standardizedProperties", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$parsedVaa.standardizedProperties", 0}}}},
		{Key: "symbol", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$transferPrices.symbol", 0}}}},
		{Key: "usdAmount", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$transferPrices.usdAmount", 0}}}},
		{Key: "tokenAmount", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$transferPrices.tokenAmount", 0}}}},
	}}})

	// unset
	pipeline = append(pipeline, bson.D{{Key: "$unset", Value: bson.A{"transferPrices", "parsedVaa"}}})

	// Execute the aggregation pipeline
	cur, err := r.collections.globalTransactions.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return nil, err
	}

	// Read results from cursor
	var operations []*OperationDto
	err = cur.All(ctx, &operations)
	if err != nil {
		r.logger.Error("failed to decode cursor", zap.Error(err))
		return nil, err
	}

	return operations, nil
}
