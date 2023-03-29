package consumer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Repository exposes operations over the `globalTransactions` collection.
type Repository struct {
	logger             *zap.Logger
	globalTransactions *mongo.Collection
}

// New creates a new repository.
func NewRepository(logger *zap.Logger, db *mongo.Database) *Repository {

	r := Repository{
		logger:             logger,
		globalTransactions: db.Collection("globalTransactions"),
	}

	return &r
}

// UpsertDocumentParams is a struct that contains the parameters for the upsertDocument method.
type UpsertDocumentParams struct {
	VaaId    string
	ChainId  sdk.ChainID
	TxHash   string
	TxDetail *chains.TxDetail
	TxStatus SourceTxStatus
}

func (r *Repository) UpsertDocument(ctx context.Context, params *UpsertDocumentParams) error {

	fields := bson.D{
		{Key: "chainId", Value: params.ChainId},
		{Key: "txHash", Value: params.TxHash},
		{Key: "status", Value: params.TxStatus},
	}

	if params.TxDetail != nil {
		fields = append(fields, primitive.E{Key: "timestamp", Value: params.TxDetail.Timestamp})
		fields = append(fields, primitive.E{Key: "signer", Value: params.TxDetail.Signer})

		// It is still to be defined whether we want to expose this field to the API consumers,
		// since it can be obtained from the original TxHash.
		//fields = append(fields, primitive.E{Key: "nativeTxHash", Value: txDetail.NativeTxHash})
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
	}

	opts := options.Update().SetUpsert(true)

	_, err := r.globalTransactions.UpdateByID(ctx, params.VaaId, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert source tx information: %w", err)
	}

	return nil
}

// CountIncompleteDocuments returns the number of documents that have destTx data, but don't have sourceTx data.
func (r *Repository) CountIncompleteDocuments(ctx context.Context) (uint64, error) {

	// Build the aggregation pipeline
	var pipeline mongo.Pipeline
	{
		// Look up transactions that either:
		// 1. have not been processed
		// 2. have been processed, but encountered an internal error
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{
				{"$or", bson.A{
					bson.D{{"originTx", bson.D{{"$exists", false}}}},
					bson.D{{"originTx.status", bson.M{"$eq": SourceTxStatusInternalError}}},
				}},
			}},
		})

		// Count the number of results
		pipeline = append(pipeline, bson.D{
			{"$count", "numGlobalTransactions"},
		})
	}

	// Execute the aggregation pipeline
	cur, err := r.globalTransactions.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return 0, err
	}

	// Read results from cursor
	var results []struct {
		NumGlobalTransactions uint64 `bson:"numGlobalTransactions"`
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

	return results[0].NumGlobalTransactions, nil
}

type GlobalTransaction struct {
	Id   string       `bson:"_id"`
	Vaas []vaa.VaaDoc `bson:"vaas"`
}

// GetIncompleteDocuments gets a batch of VAA IDs from the database.
func (r *Repository) GetIncompleteDocuments(
	ctx context.Context,
	maxId string,
	limit uint,
) ([]GlobalTransaction, error) {

	// Build the aggregation pipeline
	var pipeline mongo.Pipeline
	{
		// Specify sorting criteria
		pipeline = append(pipeline, bson.D{
			{"$sort", bson.D{bson.E{"_id", 1}}},
		})

		// filter out already processed documents
		//
		// We use the _id field as a pagination cursor
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{{"_id", bson.M{"$gt": maxId}}}},
		})

		// Look up transactions that either:
		// 1. have not been processed
		// 2. have been processed, but encountered an internal error
		pipeline = append(pipeline, bson.D{
			{"$match", bson.D{
				{"$or", bson.A{
					bson.D{{"originTx", bson.D{{"$exists", false}}}},
					bson.D{{"originTx.status", bson.M{"$eq": SourceTxStatusInternalError}}},
				}},
			}},
		})

		// Left join on the VAA collection
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "vaas"},
				{"localField", "_id"},
				{"foreignField", "_id"},
				{"as", "vaas"},
			}},
		})

		// Limit size of results
		pipeline = append(pipeline, bson.D{
			{"$limit", limit},
		})
	}

	// Execute the aggregation pipeline
	cur, err := r.globalTransactions.Aggregate(ctx, pipeline)
	if err != nil {
		r.logger.Error("failed execute aggregation pipeline", zap.Error(err))
		return nil, errors.WithStack(err)
	}

	// Read results from cursor
	var documents []GlobalTransaction
	err = cur.All(ctx, &documents)
	if err != nil {
		r.logger.Error("failed to decode cursor", zap.Error(err))
		return nil, errors.WithStack(err)
	}

	return documents, nil
}
