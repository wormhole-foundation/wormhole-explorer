package address

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/common"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Repository struct {
	db     *mongo.Database
	logger *zap.Logger

	collections struct {
		parsedVaa *mongo.Collection
	}
}

func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger: logger.With(zap.String("module", "AddressRepository")),
		collections: struct {
			parsedVaa *mongo.Collection
		}{
			parsedVaa: db.Collection("parsedVaa"),
		},
	}
}

type GetAddressOverviewParams struct {
	Address string
	Skip    int64
	Limit   int64
}

func (r *Repository) GetAddressOverview(ctx context.Context, params *GetAddressOverviewParams) (*AddressOverview, error) {

	ids, err := common.FindVaasIdsByFromAddressOrToAddress(ctx, r.db, params.Address)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		var result []*vaa.VaaDoc
		return &AddressOverview{Vaas: result}, nil
	}

	// build a query pipeline based on input parameters
	var pipeline mongo.Pipeline
	{
		// filter by list ids
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}}})

		// specify sorting criteria
		pipeline = append(pipeline, bson.D{
			{"$sort", bson.D{bson.E{"indexedAt", -1}}},
		})

		// left outer join on the `vaas` collection
		pipeline = append(pipeline, bson.D{
			{"$lookup", bson.D{
				{"from", "vaas"},
				{"localField", "_id"},
				{"foreignField", "_id"},
				{"as", "vaas"},
			}},
		})

		// skip initial results
		if params.Skip != 0 {
			pipeline = append(pipeline, bson.D{
				{"$skip", params.Skip},
			})
		}

		// limit size of results
		pipeline = append(pipeline, bson.D{
			{"$limit", params.Limit},
		})
	}

	// execute the aggregation pipeline
	cur, err := r.collections.parsedVaa.Aggregate(ctx, pipeline)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get vaa with payload",
			zap.Error(err),
			zap.Any("params", params),
			zap.String("requestID", requestID),
		)
		return nil, err
	}

	// read results from cursor
	var documents []struct {
		ID   string       `bson:"_id"`
		Vaas []vaa.VaaDoc `bson:"vaas"`
	}
	err = cur.All(ctx, &documents)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed to decode cursor for account activity",
			zap.Error(err),
			zap.Any("params", params),
			zap.String("requestID", requestID),
		)
		return nil, err
	}

	// build the result and return
	var vaas []*vaa.VaaDoc
	for i := range documents {
		if len(documents[i].Vaas) != 1 {
			r.logger.Warn("expected exactly 1 vaa document",
				zap.Int("numVaas", len(documents[i].Vaas)),
				zap.String("_id", documents[i].ID),
			)
		}
		vaas = append(vaas, &documents[i].Vaas[0])
	}
	return &AddressOverview{Vaas: vaas}, nil
}
