package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// VaaRepository is a repository for VAA.
type VaaRepository struct {
	db     *mongo.Database
	logger *zap.Logger
	vaas   *mongo.Collection
}

// VaaDoc is a document for VAA.
type VaaDoc struct {
	ID               string     `bson:"_id" json:"id"`
	Vaa              []byte     `bson:"vaas" json:"vaa"`
	ChainID          uint16     `bson:"emitterChain"`
	EmitterAddress   string     `bson:"emitterAddr"`
	Sequence         string     `bson:"sequence"`
	GuardianSetIndex uint32     `bson:"guardianSetIndex"`
	IndexedAt        time.Time  `bson:"indexedAt"`
	Timestamp        *time.Time `bson:"timestamp"`
	UpdatedAt        *time.Time `bson:"updatedAt"`
	TxHash           string     `bson:"txHash"`
	Version          int        `bson:"version"`
	Revision         int        `bson:"revision"`
}

// NewVaaRepository create a new Vaa repository.
func NewVaaRepository(db *mongo.Database, logger *zap.Logger) *VaaRepository {
	return &VaaRepository{db: db,
		logger: logger.With(zap.String("module", "VaaRepository")),
		vaas:   db.Collection(Vaas),
	}
}

// FindById finds VAA by id.
func (r *VaaRepository) FindById(ctx context.Context, id string) (*VaaDoc, error) {
	var vaaDoc VaaDoc
	err := r.vaas.FindOne(ctx, bson.M{"_id": id}).Decode(&vaaDoc)
	return &vaaDoc, err
}

// FindPageByTimeRange finds VAA by time range.
func (r *VaaRepository) FindPageByTimeRange(ctx context.Context, startTime time.Time, endTime time.Time, page, pageSize int64, sortAsc bool) ([]*VaaDoc, error) {
	filter := bson.M{
		"timestamp": bson.M{
			"$gte": startTime,
			"$lt":  endTime,
		},
	}
	sort := -1
	if sortAsc {
		sort = 1
	}

	skip := page * pageSize
	opts := &options.FindOptions{Skip: &skip, Limit: &pageSize, Sort: bson.M{"timestamp": sort}}
	cur, err := r.vaas.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var vaas []*VaaDoc
	err = cur.All(ctx, &vaas)
	return vaas, err
}

// FindPage finds VAA by query and pagination.
func (r *VaaRepository) FindPage(ctx context.Context, query VaaQuery, pagination Pagination) ([]*VaaDoc, error) {

	filter := bson.M{}

	if query.StartTime != nil || query.EndTime != nil {
		rangeTimestamp := bson.M{}
		if query.StartTime != nil {
			rangeTimestamp["$gte"] = query.StartTime
		}
		if query.EndTime != nil {
			rangeTimestamp["$lt"] = query.EndTime
		}
		filter["timestamp"] = rangeTimestamp
	}

	if query.EmitterChainID != nil {
		filter["emitterChain"] = query.EmitterChainID
	}
	if query.EmitterAddress != nil {
		filter["emitterAddr"] = query.EmitterAddress
	}
	if query.Sequence != nil {
		filter["sequence"] = query.Sequence
	}

	sort := -1
	if pagination.SortAsc {
		sort = 1
	}

	skip := pagination.Page * pagination.PageSize
	opts := &options.FindOptions{Skip: &skip, Limit: &pagination.PageSize, Sort: bson.M{"timestamp": sort}}
	cur, err := r.vaas.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var vaas []*VaaDoc
	err = cur.All(ctx, &vaas)
	return vaas, err
}
