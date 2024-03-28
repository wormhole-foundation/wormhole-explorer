package operations

import (
	"context"
	"fmt"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

type OperationQuery struct {
	Pagination     pagination.Pagination
	TxHash         string
	Address        string
	SourceChainID  *vaa.ChainID
	TargetChainID  *vaa.ChainID
	AppID          string
	ExclusiveAppId bool
}

func buildQueryOperationsByChain(sourceChainID, targetChainID *vaa.ChainID) bson.D {

	var allMatch bson.A

	if sourceChainID != nil {
		matchSourceChain := bson.M{"rawStandardizedProperties.fromChain": *sourceChainID}
		allMatch = append(allMatch, matchSourceChain)
	}

	if targetChainID != nil {
		matchTargetChain := bson.M{"rawStandardizedProperties.toChain": *targetChainID}
		allMatch = append(allMatch, matchTargetChain)
	}

	if (sourceChainID != nil && targetChainID != nil) && (*sourceChainID == *targetChainID) {
		return bson.D{{Key: "$match", Value: bson.M{"$or": allMatch}}}
	}

	return bson.D{{Key: "$match", Value: bson.M{"$and": allMatch}}}
}

func buildQueryOperationsByAppID(appID string, exclusive bool) []bson.D {
	var result []bson.D

	if appID == "" {
		result = append(result, bson.D{{Key: "$match", Value: bson.M{}}})
		return result
	}

	if exclusive {
		result = append(result, bson.D{{Key: "$match", Value: bson.M{
			"$and": bson.A{
				bson.M{"rawStandardizedProperties.appIds": bson.M{"$eq": []string{appID}}},
				bson.M{"rawStandardizedProperties.appIds": bson.M{"$size": 1}},
			}}}})
		return result

	} else {
		result = append(result, bson.D{{Key: "$match", Value: bson.M{"rawStandardizedProperties.appIds": bson.M{"$in": []string{appID}}}}})
	}
	return result
}

// findOperationsIdByAddress returns all operations filtered by address.
func findOperationsIdByAddress(ctx context.Context, db *mongo.Database, address string, pagination *pagination.Pagination) ([]string, error) {
	addressHex := strings.ToLower(address)
	if !utils.StartsWith0x(address) {
		addressHex = "0x" + strings.ToLower(addressHex)
	}

	matchGlobalTransactions := bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "originTx.from", Value: bson.M{"$eq": addressHex}}},
		bson.D{{Key: "originTx.from", Value: bson.M{"$eq": address}}},
		bson.D{{Key: "originTx.attribute.value.originAddress", Value: bson.M{"$eq": addressHex}}},
		bson.D{{Key: "originTx.attribute.value.originAddress", Value: bson.M{"$eq": address}}},
	}}}}}

	matchParsedVaa := bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "standardizedProperties.toAddress", Value: bson.M{"$eq": addressHex}}},
		bson.D{{Key: "standardizedProperties.toAddress", Value: bson.M{"$eq": address}}},
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

// matchOperationByTxHash returns a mongo pipeline to match operations by txHash.
func (r *Repository) matchOperationByTxHash(ctx context.Context, txHash string) primitive.D {
	// build txHash field to search in mongo
	txHashHex := strings.ToLower(txHash)
	if !utils.StartsWith0x(txHash) {
		txHashHex = "0x" + strings.ToLower(txHashHex)
	}

	// build destination txHash field to search in mongo
	var qLowerWith0X, qHigherWith0X string
	qLower := strings.ToLower(txHash)
	qHigher := strings.ToUpper(txHash)
	if !utils.StartsWith0x(txHash) {
		qLowerWith0X = "0x" + strings.ToLower(qLower)
		qHigherWith0X = "0x" + strings.ToUpper(qHigher)
	}

	return bson.D{{Key: "$match", Value: bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "originTx.nativeTxHash", Value: bson.M{"$eq": txHashHex}}},
		bson.D{{Key: "originTx.nativeTxHash", Value: bson.M{"$eq": txHash}}},
		bson.D{{Key: "originTx.nativeTxHash", Value: bson.M{"$eq": qLower}}},
		bson.D{{Key: "originTx.nativeTxHash", Value: bson.M{"$eq": qHigher}}},
		bson.D{{Key: "originTx.attribute.value.originTxHash", Value: bson.M{"$eq": txHashHex}}},
		bson.D{{Key: "originTx.attribute.value.originTxHash", Value: bson.M{"$eq": txHash}}},
		bson.D{{Key: "originTx.attribute.value.originTxHash", Value: bson.M{"$eq": qLower}}},
		bson.D{{Key: "originTx.attribute.value.originTxHash", Value: bson.M{"$eq": qHigher}}},
		bson.D{{Key: "destinationTx.txHash", Value: bson.M{"$eq": txHash}}},
		bson.D{{Key: "destinationTx.txHash", Value: bson.M{"$eq": qLower}}},
		bson.D{{Key: "destinationTx.txHash", Value: bson.M{"$eq": qHigher}}},
		bson.D{{Key: "destinationTx.txHash", Value: bson.M{"$eq": qLowerWith0X}}},
		bson.D{{Key: "destinationTx.txHash", Value: bson.M{"$eq": qHigherWith0X}}},
	}}}}}
}

func (r *Repository) FindByChainAndAppId(ctx context.Context, query OperationQuery) ([]*OperationDto, error) {

	var pipeline mongo.Pipeline

	if query.SourceChainID != nil || query.TargetChainID != nil {
		matchBySourceTargetChain := buildQueryOperationsByChain(query.SourceChainID, query.TargetChainID)
		pipeline = append(pipeline, matchBySourceTargetChain)
	}

	if len(query.AppID) > 0 {
		matchByAppId := buildQueryOperationsByAppID(query.AppID, query.ExclusiveAppId)
		pipeline = append(pipeline, matchByAppId...)
	}

	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{
		bson.E{Key: "updatedAt", Value: query.Pagination.GetSortInt()},
		bson.E{Key: "_id", Value: -1},
	}}})

	// Skip initial results
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: query.Pagination.Skip}})

	// Limit size of results
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: query.Pagination.Limit}})

	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "vaas"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "vaas"}}}})

	// lookup transferPrices
	pipeline = append(pipeline, bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "transferPrices"}, {Key: "localField", Value: "_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "transferPrices"}}}})

	// add fields
	pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.D{
		{Key: "payload", Value: "$parsedPayload"},
		{Key: "vaa", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$vaas", 0}}}},
		{Key: "symbol", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$transferPrices.symbol", 0}}}},
		{Key: "usdAmount", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$transferPrices.usdAmount", 0}}}},
		{Key: "tokenAmount", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$transferPrices.tokenAmount", 0}}}},
	}}})

	// unset
	pipeline = append(pipeline, bson.D{{Key: "$unset", Value: bson.A{"transferPrices"}}})

	cur, err := r.collections.parsedVaa.Aggregate(ctx, pipeline)
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

// FindAll returns all operations filtered by q.
func (r *Repository) FindAll(ctx context.Context, query OperationQuery) ([]*OperationDto, error) {

	var pipeline mongo.Pipeline

	// filter operations by address or txHash
	if query.Address != "" {
		// find all ids that match by address
		ids, err := findOperationsIdByAddress(ctx, r.db, query.Address, &query.Pagination)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return []*OperationDto{}, nil
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}}})
	} else if query.TxHash != "" {
		// match operation by txHash (source tx and destination tx)
		matchByTxHash := r.matchOperationByTxHash(ctx, query.TxHash)
		pipeline = append(pipeline, matchByTxHash)
	}

	// sort
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{
		bson.E{Key: "originTx.timestamp", Value: query.Pagination.GetSortInt()},
		bson.E{Key: "_id", Value: -1},
	}}})

	// Skip initial results
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: query.Pagination.Skip}})

	// Limit size of results
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: query.Pagination.Limit}})

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
