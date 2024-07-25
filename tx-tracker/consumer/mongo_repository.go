package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// DestinationTx representa a destination transaction.
type DestinationTx struct {
	ChainID     sdk.ChainID `bson:"chainId"`
	Status      string      `bson:"status"`
	Method      string      `bson:"method"`
	TxHash      string      `bson:"txHash"`
	From        string      `bson:"from"`
	To          string      `bson:"to"`
	BlockNumber string      `bson:"blockNumber"`
	Timestamp   *time.Time  `bson:"timestamp"`
	FeeDetail   *FeeDetail  `bson:"feeDetail"`
	UpdatedAt   *time.Time  `bson:"updatedAt"`
}

type FeeDetail struct {
	Fee    string            `bson:"fee"`
	RawFee map[string]string `bson:"rawFee"`
}

// TargetTxUpdate represents a transaction document.
type TargetTxUpdate struct {
	ID          string         `bson:"vaaId"`
	VaaID       string         `bson:"_id"`
	Destination *DestinationTx `bson:"destinationTx"`
	TrackID     string         `bson:"-"`
}

// Repository exposes operations over the `globalTransactions` collection.
type Repository struct {
	logger             *zap.Logger
	globalTransactions *mongo.Collection
	vaas               *mongo.Collection
	vaaIdTxHash        *mongo.Collection
}

// New creates a new repository.
func NewRepository(logger *zap.Logger, db *mongo.Database) *Repository {

	r := Repository{
		logger:             logger,
		globalTransactions: db.Collection("globalTransactions"),
		vaas:               db.Collection("vaas"),
		vaaIdTxHash:        db.Collection("vaaIdTxHash"),
	}

	return &r
}

// UpsertOriginTxParams is a struct that contains the parameters for the upsertDocument method.
type UpsertOriginTxParams struct {
	VaaId     string // {chain/address/sequence}
	Id        string // digest
	TrackID   string
	ChainId   sdk.ChainID
	TxDetail  *chains.TxDetail
	TxStatus  domain.SourceTxStatus
	Timestamp *time.Time
	Processed bool
}

func createChangesDoc(source, _type string, timestamp *time.Time) bson.D {
	return bson.D{
		{
			Key: "changes",
			Value: bson.D{
				{Key: "type", Value: _type},
				{Key: "source", Value: source},
				{Key: "timestamp", Value: timestamp},
			},
		},
	}
}

// UpsertOriginTx upserts a source transaction document.
func (r *Repository) UpsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error {

	now := time.Now()

	fields := bson.D{
		{Key: "chainId", Value: params.ChainId},
		{Key: "status", Value: params.TxStatus},
		{Key: "updatedAt", Value: now},
		{Key: "processed", Value: params.Processed},
	}

	if params.TxDetail != nil {
		fields = append(fields, primitive.E{Key: "nativeTxHash", Value: params.TxDetail.NativeTxHash})
		fields = append(fields, primitive.E{Key: "from", Value: params.TxDetail.From})
		if params.TxDetail.Attribute != nil {
			fields = append(fields, primitive.E{Key: "attribute", Value: params.TxDetail.Attribute})
		}
		if params.TxDetail.FeeDetail != nil {
			fields = append(fields, primitive.E{Key: "feeDetail", Value: params.TxDetail.FeeDetail})
		}
	}

	if params.Timestamp != nil {
		fields = append(fields, primitive.E{Key: "timestamp", Value: params.Timestamp})
	}

	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{
					Key:   "originTx",
					Value: fields,
				},
			},
		},
		{
			Key:   "$push",
			Value: createChangesDoc(params.TrackID, "originTx", &now),
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := r.globalTransactions.UpdateByID(ctx, params.VaaId, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert source tx information: %w", err)
	}

	return nil
}

// AlreadyProcessed returns true if the given VAA ID has already been processed.
func (r *Repository) AlreadyProcessed(ctx context.Context, vaaId string) (bool, error) {
	result := r.
		globalTransactions.
		FindOne(ctx, bson.D{
			{Key: "_id", Value: vaaId},
			{Key: "originTx", Value: bson.D{{Key: "$exists", Value: true}}},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "originTx.processed", Value: true}},
				bson.D{{Key: "originTx.processed", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
		})
	//  The originTx.processed will be true if the vaa was processed successfully.
	//  If exists and error getting the transactions from the rpcs, a partial originTx will save in the db and
	//  the originTx.processed will be false.

	var tx GlobalTransaction
	err := result.Decode(&tx)
	if err == mongo.ErrNoDocuments {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to decode already processed VAA id: %w", err)
	} else {
		return true, nil
	}
}

type GlobalTransaction struct {
	Id   string       `bson:"_id"`
	Vaas []vaa.VaaDoc `bson:"vaas"`
}

// VaaIdTxHash represents a vaaIdTxHash document.
type VaaIdTxHash struct {
	TxHash string `bson:"txHash"`
}

func (r *Repository) GetVaaIdTxHash(ctx context.Context, id string) (*VaaIdTxHash, error) {
	var v VaaIdTxHash
	err := r.vaaIdTxHash.FindOne(ctx, bson.M{"_id": id}).Decode(&v)
	return &v, err
}

func (r *Repository) UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error {
	update := bson.M{
		"$set":  globalTx,
		"$push": createChangesDoc(globalTx.TrackID, "destinationTx", globalTx.Destination.UpdatedAt),
	}

	_, err := r.globalTransactions.UpdateByID(ctx, globalTx.VaaID, update, options.Update().SetUpsert(true))
	if err != nil {
		r.logger.Error("Error inserting target tx in global transaction", zap.Error(err))
		return err
	}
	return err
}

// AlreadyProcessed returns true if the given VAA ID has already been processed.
func (r *Repository) GetTargetTx(ctx context.Context, vaaId string) (*TargetTxUpdate, error) {

	result := r.
		globalTransactions.
		FindOne(ctx, bson.D{
			{Key: "_id", Value: vaaId},
			{Key: "destinationTx", Value: bson.D{{Key: "$exists", Value: true}}},
		})

	var tx TargetTxUpdate
	err := result.Decode(&tx)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to decode already processed VAA id: %w", err)
	} else {
		return &tx, nil
	}
}

// CountDocumentsByTimeRange returns the number of documents that match the given time range.
func (r *Repository) CountDocumentsByVaas(
	ctx context.Context,
	emitterChainID sdk.ChainID,
	emitterAddress string,
	sequence string,
) (uint64, error) {

	// Build the aggregation pipeline
	var pipeline mongo.Pipeline
	{
		// filter by emitterChain
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{{Key: "emitterChain", Value: emitterChainID}}},
		})

		// filter by emitterAddr
		if emitterAddress != "" {
			pipeline = append(pipeline, bson.D{
				{Key: "$match", Value: bson.D{{Key: "emitterAddr", Value: emitterAddress}}},
			})
		}

		// filter by sequence
		if sequence != "" {
			pipeline = append(pipeline, bson.D{
				{Key: "$match", Value: bson.D{{Key: "sequence", Value: sequence}}},
			})
		}

		// Count the number of results
		pipeline = append(pipeline, bson.D{
			{Key: "$count", Value: "numDocuments"},
		})
	}

	// Execute the aggregation pipeline
	cur, err := r.vaas.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return 0, err
	}

	// Read results from cursor
	var results []struct {
		NumDocuments uint64 `bson:"numDocuments"`
	}
	err = cur.All(ctx, &results)
	if err != nil {
		r.logger.Error("failed to decode cursor", zap.Error(err))
		return 0, err
	}
	if len(results) == 0 {
		return 0, nil
	}
	if len(results) > 1 {
		r.logger.Error("too many results", zap.Int("numResults", len(results)))
		return 0, err
	}

	return results[0].NumDocuments, nil
}

// GetDocumentsByTimeRange iterates through documents within a specified time range.
func (r *Repository) GetDocumentsByVaas(
	ctx context.Context,
	lastId string,
	lastTimestamp *time.Time,
	limit uint,
	emitterChainID sdk.ChainID,
	emitterAddress string,
	sequence string,
) ([]GlobalTransaction, error) {

	// Build the aggregation pipeline
	var pipeline mongo.Pipeline
	{
		// Specify sorting criteria
		pipeline = append(pipeline, bson.D{
			{Key: "$sort", Value: bson.D{
				bson.E{Key: "timestamp", Value: -1},
				bson.E{Key: "_id", Value: 1},
			}},
		})

		// filter out already processed documents
		//
		// We use the timestap field as a pagination cursor
		if lastTimestamp != nil {
			pipeline = append(pipeline, bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "$or", Value: bson.A{
						bson.D{{Key: "timestamp", Value: bson.M{"$lt": *lastTimestamp}}},
						bson.D{{Key: "$and", Value: bson.A{
							bson.D{{Key: "timestamp", Value: bson.M{"$eq": *lastTimestamp}}},
							bson.D{{Key: "_id", Value: bson.M{"$gt": lastId}}},
						}}},
					}},
				}},
			})
		}

		// filter by emitterChain
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{{Key: "emitterChain", Value: emitterChainID}}},
		})

		// filter by emitterAddr
		if emitterAddress != "" {
			pipeline = append(pipeline, bson.D{
				{Key: "$match", Value: bson.D{{Key: "emitterAddr", Value: emitterAddress}}},
			})
		}

		// filter by sequence
		if sequence != "" {
			pipeline = append(pipeline, bson.D{
				{Key: "$match", Value: bson.D{{Key: "sequence", Value: sequence}}},
			})
		}

		// Limit size of results
		pipeline = append(pipeline, bson.D{
			{Key: "$limit", Value: limit},
		})
	}

	// Execute the aggregation pipeline
	cur, err := r.vaas.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return nil, errors.WithStack(err)
	}

	// Read results from cursor
	var documents []vaa.VaaDoc
	err = cur.All(ctx, &documents)
	if err != nil {
		r.logger.Error("failed to decode cursor", zap.Error(err))
		return nil, errors.WithStack(err)
	}

	// Build the result
	var globalTransactions []GlobalTransaction
	for i := range documents {
		globalTransaction := GlobalTransaction{
			Id:   documents[i].ID,
			Vaas: []vaa.VaaDoc{documents[i]},
		}
		globalTransactions = append(globalTransactions, globalTransaction)
	}

	return globalTransactions, nil
}

// SourceTxDoc represents a source transaction document.
type SourceTxDoc struct {
	ID       string `bson:"_id"`
	OriginTx *struct {
		ChainID      int    `bson:"chainId"`
		Status       string `bson:"status"`
		Processed    bool   `bson:"processed"`
		NativeTxHash string `bson:"nativeTxHash"`
		From         string `bson:"from"`
	} `bson:"originTx"`
}

// FindSourceTxById returns the source transaction document with the given ID.
func (r *Repository) FindSourceTxById(ctx context.Context, id string) (*SourceTxDoc, error) {
	var sourceTxDoc SourceTxDoc
	err := r.globalTransactions.FindOne(ctx, bson.M{"_id": id}).Decode(&sourceTxDoc)
	if err != nil {
		return nil, err
	}
	return &sourceTxDoc, err
}
