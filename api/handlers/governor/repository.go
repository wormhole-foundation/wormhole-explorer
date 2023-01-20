// Package governor handle the request of governor data from governor endpoint defined in the api.
package governor

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/certusone/wormhole/node/pkg/vaa"
	"github.com/pkg/errors"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const minGuardianNum = 13

// Repository definition.
type Repository struct {
	db          *mongo.Database
	logger      *zap.Logger
	collections struct {
		governorConfig *mongo.Collection
		governorStatus *mongo.Collection
	}
}

// NewRepository create a new Repository.
func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger: logger.With(zap.String("module", "GovernorRepository")),
		collections: struct {
			governorConfig *mongo.Collection
			governorStatus *mongo.Collection
		}{
			governorConfig: db.Collection("governorConfig"),
			governorStatus: db.Collection("governorStatus"),
		},
	}
}

// GovernorQuery respresent a query for the governors mongodb documents.
type GovernorQuery struct {
	pagination.Pagination
	id string
}

// QueryGovernor create a new GovernorQuery with default pagination values.
func QueryGovernor() *GovernorQuery {
	page := pagination.FirstPage()
	return &GovernorQuery{Pagination: *page}
}

// SetID set the id field of the GovernorQuery struct.
func (q *GovernorQuery) SetID(id string) *GovernorQuery {
	q.id = id
	return q
}

// SetPagination set the pagination field of the GovernorQuery struct.
func (q *GovernorQuery) SetPagination(p *pagination.Pagination) *GovernorQuery {
	q.Pagination = *p
	return q
}

func (q *GovernorQuery) toBSON() *bson.D {
	r := bson.D{}
	if q.id != "" {
		r = append(r, bson.E{Key: "_id", Value: q.id})
	}
	return &r
}

// FindGovConfigurations get a list of *GovConfig.
func (r *Repository) FindGovConfigurations(ctx context.Context, q *GovernorQuery) ([]*GovConfig, error) {
	if q == nil {
		q = QueryGovernor()
	}
	sort := bson.D{{Key: q.SortBy, Value: q.GetSortInt()}}
	projection := bson.D{
		{Key: "createdAt", Value: 1},
		{Key: "updatedAt", Value: 1},
		{Key: "nodename", Value: "$parsedConfig.nodename"},
		{Key: "counter", Value: "$parsedConfig.counter"},
		{Key: "chains", Value: "$parsedConfig.chains"},
		{Key: "tokens", Value: "$parsedConfig.tokens"},
	}
	options := options.Find().SetProjection(projection).SetLimit(q.PageSize).SetSkip(q.Offset).SetSort(sort)
	cur, err := r.collections.governorConfig.Find(ctx, q.toBSON(), options)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Find command to get governor configurations",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	var govConfigs []*GovConfig
	err = cur.All(ctx, &govConfigs)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*GovConfig", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	return govConfigs, err
}

// FindGovConfiguration get a *GovConfig. The q parameter define the filter to apply to the query.
func (r *Repository) FindGovConfiguration(ctx context.Context, q *GovernorQuery) (*GovConfig, error) {
	if q == nil {
		return nil, errs.ErrMalformedQuery
	}
	var govConfig GovConfig
	projection := bson.D{
		{Key: "createdAt", Value: 1},
		{Key: "updatedAt", Value: 1},
		{Key: "nodename", Value: "$parsedConfig.nodename"},
		{Key: "counter", Value: "$parsedConfig.counter"},
		{Key: "chains", Value: "$parsedConfig.chains"},
		{Key: "tokens", Value: "$parsedConfig.tokens"},
	}
	options := options.FindOne().SetProjection(projection)
	err := r.collections.governorConfig.FindOne(ctx, q.toBSON(), options).Decode(&govConfig)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get governor configuration",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	return &govConfig, err
}

// FindGovernorStatus get a list of *GovStatus.
func (r *Repository) FindGovernorStatus(ctx context.Context, q *GovernorQuery) ([]*GovStatus, error) {
	if q == nil {
		q = QueryGovernor()
	}
	sort := bson.D{{Key: q.SortBy, Value: q.GetSortInt()}}
	projection := bson.D{
		{Key: "createdAt", Value: 1},
		{Key: "updatedAt", Value: 1},
		{Key: "nodename", Value: "$parsedStatus.nodename"},
		{Key: "chains", Value: "$parsedStatus.chains"},
	}
	options := options.Find().SetProjection(projection).SetLimit(q.PageSize).SetSkip(q.Offset).SetSort(sort)
	cur, err := r.collections.governorStatus.Find(ctx, q.toBSON(), options)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Find command to get all governor status",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	var govStatus []*GovStatus
	err = cur.All(ctx, &govStatus)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*GovStatus", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	return govStatus, err
}

// FindOneGovernorStatus get a *GovStatus. The q parameter define the filter to apply to the query.
func (r *Repository) FindOneGovernorStatus(ctx context.Context, q *GovernorQuery) (*GovStatus, error) {
	if q == nil {
		return nil, errs.ErrMalformedQuery
	}
	var govConfig GovStatus
	projection := bson.D{
		{Key: "createdAt", Value: 1},
		{Key: "updatedAt", Value: 1},
		{Key: "nodename", Value: "$parsedStatus.nodename"},
		{Key: "chains", Value: "$parsedStatus.chains"},
	}
	options := options.FindOne().SetProjection(projection)
	err := r.collections.governorStatus.FindOne(ctx, q.toBSON(), options).Decode(&govConfig)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get governor status",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	return &govConfig, err
}

// NotionalLimitQuery
type NotionalLimitQuery struct {
	pagination.Pagination
	id      string
	chainID vaa.ChainID
}

// QueryNotionalLimit create a new NotionalLimitQuery with default pagination values.
func QueryNotionalLimit() *NotionalLimitQuery {
	page := pagination.FirstPage()
	return &NotionalLimitQuery{Pagination: *page}
}

// SetID set the id field of the NotionalLimitQuery struct.
func (q *NotionalLimitQuery) SetID(id string) *NotionalLimitQuery {
	q.id = id
	return q
}

// SetChain set the chainID field of the NotionalLimitQuery struct.
func (q *NotionalLimitQuery) SetChain(chainID vaa.ChainID) *NotionalLimitQuery {
	q.chainID = chainID
	return q
}

// SetPagination set the Pagination field of the NotionalLimitQuery struct.
func (q *NotionalLimitQuery) SetPagination(p *pagination.Pagination) *NotionalLimitQuery {
	q.Pagination = *p
	return q
}

// FindNotionalLimit get a list *NotionalLimit.
func (r *Repository) FindNotionalLimit(ctx context.Context, q *NotionalLimitQuery) ([]*NotionalLimit, error) {
	// agreggation stages to get notionalLimit for each chainID.
	matchStage1 := bson.D{{Key: "$match", Value: bson.D{}}}

	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chains", Value: "$parsedConfig.chains"},
		}},
	}

	unwindStage3 := bson.D{
		{Key: "$unwind", Value: "$chains"},
	}

	sortStage4 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "chains.chainid", Value: 1},
			{Key: "chains.notionallimit", Value: -1},
			{Key: "chains.bigtransactionsize", Value: -1},
		}},
	}

	groupStage5 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$chains.chainid"},
			{Key: "notionalLimits", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "notionalLimit", Value: "$chains.notionallimit"},
					{Key: "maxTransactionSize", Value: "$chains.bigtransactionsize"},
				}},
			}},
		}},
	}

	projectStage6 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainId", Value: "$_id"},
			{Key: "notionalLimit", Value: bson.M{
				"$arrayElemAt": []interface{}{"$notionalLimits", minGuardianNum - 1},
			}},
		}},
	}

	projectStage7 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainId", Value: 1},
			{Key: "notionalLimit", Value: "$notionalLimit.notionalLimit"},
			{Key: "maxTransactionSize", Value: "$notionalLimit.maxTransactionSize"},
		}},
	}

	sortStage8 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "chainId", Value: 1},
		}},
	}

	// define aggregate pipeline
	pipeLine := mongo.Pipeline{
		matchStage1,
		projectStage2,
		unwindStage3,
		sortStage4,
		groupStage5,
		projectStage6,
		projectStage7,
		sortStage8,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorConfig.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get notional limit",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decodes to NotionalLimit.
	var notionalLimits []*NotionalLimit
	err = cur.All(ctx, &notionalLimits)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*NotionalLimit", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// check records exists.
	if len(notionalLimits) == 0 {
		return nil, errs.ErrNotFound
	}

	return notionalLimits, nil
}

// GetNotionalLimitByChainID get a list *NotionalLimitDetail.
func (r *Repository) GetNotionalLimitByChainID(ctx context.Context, q *NotionalLimitQuery) ([]*NotionalLimitDetail, error) {
	// agreggation stages to get notionalLimit by chainID.
	matchStage1 := bson.D{{Key: "$match", Value: bson.D{}}}

	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: "$parsedConfig.nodename"},
			{Key: "parsedConfig.chains", Value: bson.D{
				{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$parsedConfig.chains"},
					{Key: "as", Value: "chain"},
					{Key: "cond", Value: bson.D{
						{Key: "$eq", Value: bson.A{"$$chain.chainid", q.chainID}},
					}},
				}},
			}},
		}},
	}

	projectStage3 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: 1},
			{Key: "notionalLimits", Value: bson.M{
				"$arrayElemAt": []interface{}{"$parsedConfig.chains", 0},
			}},
		}},
	}

	projectStage4 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: 1},
			{Key: "chainId", Value: "$notionalLimits.chainid"},
			{Key: "notionalLimit", Value: "$notionalLimits.notionallimit"},
			{Key: "maxTransactionSize", Value: "$notionalLimits.bigtransactionsize"},
		}},
	}

	pipeLine := mongo.Pipeline{
		matchStage1,
		projectStage2,
		projectStage3,
		projectStage4,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorConfig.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get notional limit by chainID",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decode to []NotionalLimitRecord.
	var notionalLimits []*NotionalLimitDetail
	err = cur.All(ctx, &notionalLimits)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*NotionalLimitDetail", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	return notionalLimits, nil
}

// GetAvailableNotional get a list of *NotionalAvailable.
func (r *Repository) GetAvailableNotional(ctx context.Context, q *NotionalLimitQuery) ([]*NotionalAvailable, error) {
	// stage.
	matchStage1 := bson.D{{Key: "$match", Value: bson.D{}}}

	// project.
	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chains", Value: "$parsedStatus.chains"},
		}},
	}

	// deconstructs.
	unwindStage3 := bson.D{
		{Key: "$unwind", Value: "$chains"},
	}

	// sort.
	sortStage4 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "chains.chainid", Value: 1},
			{Key: "chains.remainingavailablenotional", Value: -1},
		}},
	}

	// group.
	groupStage5 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$chains.chainid"},
			{Key: "availableNotionals", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "availableNotional", Value: "$chains.remainingavailablenotional"},
				}},
			}},
		}},
	}

	// project.
	projectStage6 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainId", Value: "$_id"},
			{Key: "availableNotional", Value: bson.M{
				"$arrayElemAt": []interface{}{"$availableNotionals", minGuardianNum - 1},
			}},
		}},
	}

	// projects.
	projectStage7 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainId", Value: 1},
			{Key: "availableNotional", Value: "$availableNotional.availableNotional"},
		}},
	}

	// sort stage
	sortStage8 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "chainId", Value: 1},
		}},
	}

	pipeLine := mongo.Pipeline{
		matchStage1,
		projectStage2,
		unwindStage3,
		sortStage4,
		groupStage5,
		projectStage6,
		projectStage7,
		sortStage8,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorStatus.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get available notional",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decode to []NotionalLimitRecord.
	var notionalAvailables []*NotionalAvailable
	err = cur.All(ctx, &notionalAvailables)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*NotionalAvailable", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// check exists records
	if len(notionalAvailables) == 0 {
		return nil, errs.ErrNotFound
	}

	return notionalAvailables, nil
}

// GetAvailableNotionalByChainID get a list of *NotionalAvailableDetail.
func (r *Repository) GetAvailableNotionalByChainID(ctx context.Context, q *NotionalLimitQuery) ([]*NotionalAvailableDetail, error) {
	// stage definitions.
	matchStage1 := bson.D{{Key: "$match", Value: bson.D{}}}

	// projection
	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: "$parsedStatus.nodename"},
			{Key: "parsedStatus.chains", Value: bson.D{
				{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$parsedStatus.chains"},
					{Key: "as", Value: "chain"},
					{Key: "cond", Value: bson.D{
						{Key: "$eq", Value: bson.A{"$$chain.chainid", q.chainID}},
					}},
				}},
			}},
		}},
	}

	// projection
	projectStage3 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: 1},
			{Key: "availableNotional", Value: bson.M{
				"$arrayElemAt": []interface{}{"$parsedStatus.chains", 0},
			}},
		}},
	}

	// projection
	projectStage4 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: 1},
			{Key: "chainId", Value: "$availableNotional.chainid"},
			{Key: "availableNotional", Value: "$availableNotional.remainingavailablenotional"},
		}},
	}

	pipeLine := mongo.Pipeline{
		matchStage1,
		projectStage2,
		projectStage3,
		projectStage4,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorStatus.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get available notional by chainID",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decode to []NotionalLimitRecord.
	var notionalAvailability []*NotionalAvailableDetail
	err = cur.All(ctx, &notionalAvailability)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*NotionalAvailableDetail", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	return notionalAvailability, nil
}

// GetMaxNotionalAvailableByChainID get a *MaxNotionalAvailableRecord.
func (r *Repository) GetMaxNotionalAvailableByChainID(ctx context.Context, q *NotionalLimitQuery) (*MaxNotionalAvailableRecord, error) {
	// stage definitions.
	matchStage1 := bson.D{{Key: "$match", Value: bson.D{}}}

	// projection
	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: "$parsedStatus.nodename"},
			{Key: "parsedStatus.chains", Value: bson.D{
				{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$parsedStatus.chains"},
					{Key: "as", Value: "chain"},
					{Key: "cond", Value: bson.D{
						{Key: "$eq", Value: bson.A{"$$chain.chainid", q.chainID}},
					}},
				}},
			}},
		}},
	}

	// projection
	projectStage3 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: 1},
			{Key: "availableNotional", Value: bson.M{
				"$arrayElemAt": []interface{}{"$parsedStatus.chains", 0},
			}},
		}},
	}

	// projection
	projectStage4 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: 1},
			{Key: "chainId", Value: "$availableNotional.chainid"},
			{Key: "availableNotional", Value: "$availableNotional.remainingavailablenotional"},
			{Key: "emitters", Value: "$availableNotional.emitters"},
		}},
	}

	// TODO: CHECK
	sortStage5 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "availableNotional", Value: -1},
		}},
	}

	pipeLine := mongo.Pipeline{
		matchStage1,
		projectStage2,
		projectStage3,
		projectStage4,
		sortStage5,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorStatus.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get maximun available notional by chainID",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decode to []NotionalLimitRecord.
	var rows []*MaxNotionalAvailableRecord
	err = cur.All(ctx, &rows)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*MaxNotionalAvailableRecord", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// check exists records
	if len(rows) == 0 {
		return nil, errs.ErrNotFound
	}

	if len(rows) < minGuardianNum {
		return nil, errs.ErrNotFound
	}

	maxNotionalLimit := rows[minGuardianNum-1]
	return maxNotionalLimit, nil
}

// EnqueuedVaaQuery respresent a query for enqueuedVaa queries.
type EnqueuedVaaQuery struct {
	pagination.Pagination
	id      string
	chainID vaa.ChainID
}

// QueryEnqueuedVaa create a new EnqueuedVaaQuery with default pagination values.
func QueryEnqueuedVaa() *EnqueuedVaaQuery {
	page := pagination.FirstPage()
	return &EnqueuedVaaQuery{Pagination: *page}
}

// SetID set the id field of the EnqueuedVaaQuery struct.
func (q *EnqueuedVaaQuery) SetID(id string) *EnqueuedVaaQuery {
	q.id = id
	return q
}

// SetChain set the chainID field of the EnqueuedVaaQuery struct.
func (q *EnqueuedVaaQuery) SetChain(chainID vaa.ChainID) *EnqueuedVaaQuery {
	q.chainID = chainID
	return q
}

// SetPagination set the pagination field of the EnqueuedVaaQuery struct.
func (q *EnqueuedVaaQuery) SetPagination(p *pagination.Pagination) *EnqueuedVaaQuery {
	q.Pagination = *p
	return q
}

// GetEnqueueVass get a list of *EnqueuedVaas.
func (r *Repository) GetEnqueueVass(ctx context.Context, q *EnqueuedVaaQuery) ([]*EnqueuedVaas, error) {
	// match stage.
	matchStage1 := bson.D{{Key: "$match", Value: bson.D{}}}

	// match project.
	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chains", Value: "$parsedStatus.chains"},
		}},
	}

	// match unwind.
	unwindStage3 := bson.D{
		{Key: "$unwind", Value: "$chains"},
	}

	// match project.
	projectStage4 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "chainId", Value: "$chains.chainid"},
			{Key: "emitters", Value: "$chains.emitters"},
		}},
	}

	// match group.
	groupStage5 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$chainId"},
			{Key: "emitters", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "emitterAddress", Value: bson.M{
						"$arrayElemAt": []interface{}{"$emitters.emitteraddress", 0},
					}},
					{Key: "enqueuedVaas", Value: bson.M{
						"$arrayElemAt": []interface{}{"$emitters.enqueuedvaas", 0},
					}},
				}},
			}},
		}},
	}

	pipeLine := mongo.Pipeline{
		matchStage1,
		projectStage2,
		unwindStage3,
		projectStage4,
		groupStage5,
	}

	cur, err := r.collections.governorStatus.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get enqueued vaas",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	var rows []struct {
		ID       vaa.ChainID `bson:"_id"`
		Emitters []*struct {
			Address      string `bson:"emitterAddress"`
			EnqueuedVaas []*struct {
				Sequence      string     `bson:"sequence"`
				ReleaseTime   *time.Time `bson:"releasetime"`
				NotionalValue int64      `bson:"notionalValue"`
				TxHash        string     `bson:"txhash"`
			} `bson:"enqueuedVaas"`
		} `bson:"emitters"`
	}

	// decode query response.
	err = cur.All(ctx, &rows)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to rows", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	if len(rows) == 0 {
		return nil, errs.ErrNotFound
	}

	// TODO: Change this logic to mongo query code.
	// proecess and build the response.
	keys := map[string]string{}
	enqueuedVaas := []*EnqueuedVaa{}
	for _, row := range rows {
		chainID := row.ID
		emiiterAddress := row.Emitters
		for _, ea := range emiiterAddress {
			emitterAddress := ea.Address
			enqueuedVaa := ea.EnqueuedVaas
			for _, v := range enqueuedVaa {
				key := fmt.Sprintf("%s/%s/%s", emitterAddress, v.Sequence, v.TxHash)
				if _, ok := keys[key]; !ok {
					enqueuedVaa := EnqueuedVaa{
						ChainID:        chainID,
						EmitterAddress: emitterAddress,
						Sequence:       v.Sequence,
						NotionalValue:  v.NotionalValue,
						TxHash:         v.TxHash,
					}
					enqueuedVaas = append(enqueuedVaas, &enqueuedVaa)
					keys[key] = key
				}
			}
		}
	}

	if len(enqueuedVaas) == 0 {
		return nil, errs.ErrNotFound
	}

	// group by chainID.
	enqueuedVaasGroupedByChainID := map[vaa.ChainID][]*EnqueuedVaa{}
	for _, f := range enqueuedVaas {
		if _, ok := enqueuedVaasGroupedByChainID[f.ChainID]; !ok {
			enqueuedVaasGroupedByChainID[f.ChainID] = []*EnqueuedVaa{f}
		} else {
			fr := enqueuedVaasGroupedByChainID[f.ChainID]
			fr = append(fr, f)
			enqueuedVaasGroupedByChainID[f.ChainID] = fr
		}
	}

	// create response.
	response := []*EnqueuedVaas{}
	for k, v := range enqueuedVaasGroupedByChainID {
		r := EnqueuedVaas{
			ChainID:     k,
			EnqueuedVaa: v,
		}
		response = append(response, &r)
	}

	return response, nil
}

// GetEnqueueVassByChainID get a list of *EnqueuedVaaDetail by chainID.
func (r *Repository) GetEnqueueVassByChainID(ctx context.Context, q *EnqueuedVaaQuery) ([]*EnqueuedVaaDetail, error) {
	// stage definitions.
	matchStage1 := bson.D{{Key: "$match", Value: bson.D{}}}

	// project stage.
	projectStage2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: "$parsedStatus.nodename"},
			{Key: "parsedStatus.chains", Value: bson.D{
				{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$parsedStatus.chains"},
					{Key: "as", Value: "chain"},
					{Key: "cond", Value: bson.D{
						{Key: "$eq", Value: bson.A{"$$chain.chainid", q.chainID}},
					}},
				}},
			}},
		}},
	}

	// project stage.
	projectStage3 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "updatedAt", Value: 1},
			{Key: "nodeName", Value: 1},
			{Key: "emitters", Value: "$parsedStatus.chains.emitters"},
		}},
	}

	// unwind stage.
	unwindStage4 := bson.D{
		{Key: "$unwind", Value: "$emitters"},
	}

	// group stage.
	groupStage5 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.M{
				"$arrayElemAt": []interface{}{"$emitters.emitteraddress", 0},
			}},
			{Key: "enqueuedVaas", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "enqueuedVaa", Value: "$emitters.enqueuedvaas"},
				}},
			}},
		}},
	}

	pipeline := mongo.Pipeline{
		matchStage1,
		projectStage2,
		projectStage3,
		unwindStage4,
		groupStage5,
	}

	cur, err := r.collections.governorStatus.Aggregate(ctx, pipeline)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get enqueued vaas by chainID",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decode query response.
	var rows []*struct {
		ID           string `bson:"_id"`
		EnqueuedVaas []*struct {
			EnqueuedVaas [][]*struct {
				Sequence      string `bson:"sequence"`
				ReleaseTime   int64  `bson:"releasetime"`
				NotionalValue int64  `bson:"notionalValue"`
				TxHash        string `bson:"txhash"`
			} `bson:"enqueuedVaa"`
		} `bson:"enqueuedVaas"`
	}
	err = cur.All(ctx, &rows)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to rows", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// TODO: Change this logic to mongo query code.

	// build response.
	keys := map[string]string{}
	response := []*EnqueuedVaaDetail{}
	for _, row := range rows {
		emitterAddress := row.ID
		enqueuedVaas := row.EnqueuedVaas
		for _, ev := range enqueuedVaas {
			for _, v := range ev.EnqueuedVaas[0] {
				key := fmt.Sprintf("%s/%s/%s", emitterAddress, v.Sequence, v.TxHash)
				if _, ok := keys[key]; !ok {
					fr := EnqueuedVaaDetail{
						ChainID:        q.chainID,
						EmitterAddress: emitterAddress,
						Sequence:       v.Sequence,
						NotionalValue:  v.NotionalValue,
						TxHash:         v.TxHash,
						ReleaseTime:    v.ReleaseTime,
					}
					response = append(response, &fr)
					keys[key] = key
				}
			}
		}
	}

	if len(response) == 0 {
		return nil, errs.ErrNotFound
	}

	// sort response by sequence.
	sort.Slice(response, func(i, j int) bool {
		return response[i].Sequence < response[j].Sequence
	})
	return response, nil
}

// GetGovernorLimit get a list of *GovernorLimit.
func (r *Repository) GetGovernorLimit(ctx context.Context, q *GovernorQuery) ([]*GovernorLimit, error) {
	// lookup.
	lookupStage1 := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "governorStatus"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "status"},
		}},
	}

	// unwind.
	unwindStage2 := bson.D{
		{Key: "$unwind", Value: "$status"},
	}

	projectStage3 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "configChains", Value: "$parsedConfig.chains"},
			{Key: "statusChains", Value: "$status.parsedStatus.chains"},
		}},
	}

	unwindStage4 := bson.D{
		{Key: "$unwind", Value: "$configChains"},
	}

	unwindStage5 := bson.D{
		{Key: "$unwind", Value: "$statusChains"},
	}

	matchStage6 := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{
				{Key: "$eq", Value: bson.A{"$configChains.chainid", "$statusChains.chainid"}},
			}},
		}},
	}

	groupStage7 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$configChains.chainid"},
			{Key: "notionalLimits", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "notionalLimit", Value: "$configChains.notionallimit"},
				}},
			}},
			{Key: "maxTransactionSizes", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "maxTransactionSize", Value: "$configChains.bigtransactionsize"},
				}},
			}},
			{Key: "availableNotionals", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "availableNotional", Value: "$statusChains.remainingavailablenotional"},
				}},
			}},
		}},
	}

	projectStage8 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "notionalLimits", Value: bson.D{
				{Key: "$sortArray", Value: bson.D{
					{Key: "input", Value: "$notionalLimits"},
					{Key: "sortBy", Value: bson.D{
						{Key: "notionalLimit", Value: -1},
					}},
				}},
			}},
			{Key: "maxTransactionSizes", Value: bson.D{
				{Key: "$sortArray", Value: bson.D{
					{Key: "input", Value: "$maxTransactionSizes"},
					{Key: "sortBy", Value: bson.D{
						{Key: "maxTransactionSize", Value: -1},
					}},
				}},
			}},
			{Key: "availableNotionals", Value: bson.D{
				{Key: "$sortArray", Value: bson.D{
					{Key: "input", Value: "$availableNotionals"},
					{Key: "sortBy", Value: bson.D{
						{Key: "availableNotional", Value: -1},
					}},
				}},
			}},
		}},
	}

	projectStage9 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "notionalLimit", Value: bson.M{
				"$arrayElemAt": []interface{}{"$notionalLimits", minGuardianNum - 1},
			}},
			{Key: "maxTransactionSize", Value: bson.M{
				"$arrayElemAt": []interface{}{"$maxTransactionSizes", minGuardianNum - 1},
			}},
			{Key: "availableNotional", Value: bson.M{
				"$arrayElemAt": []interface{}{"$availableNotionals", minGuardianNum - 1},
			}},
		}},
	}

	projectStage10 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainId", Value: "$_id"},
			{Key: "notionalLimit", Value: "$notionalLimit.notionalLimit"},
			{Key: "maxTransactionSize", Value: "$maxTransactionSize.maxTransactionSize"},
			{Key: "availableNotional", Value: "$availableNotional.availableNotional"},
		}},
	}

	sortStage11 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "chainId", Value: 1},
		}},
	}

	// define aggregate pipeline
	pipeLine := mongo.Pipeline{
		lookupStage1,
		unwindStage2,
		projectStage3,
		unwindStage4,
		unwindStage5,
		matchStage6,
		groupStage7,
		projectStage8,
		projectStage9,
		projectStage10,
		sortStage11,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorConfig.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get governor limit",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decodes to RawDocRecord.
	var governorLimits []*GovernorLimit
	err = cur.All(ctx, &governorLimits)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*GovernorLimit", zap.Error(err), zap.Any("q", q),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// check exists records
	if len(governorLimits) == 0 {
		return nil, errs.ErrNotFound
	}

	return governorLimits, nil
}

// GetAvailNotionByChain get the limits by chainID.
// In this version returns the minimum value of the availableNotional per chainID by analyzing the data of all guardian nodes.
func (r *Repository) GetAvailNotionByChain(ctx context.Context) ([]*AvailableNotionalByChain, error) {
	lookupStage1 := bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "governorStatus"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "status"},
		}},
	}

	unwindStage2 := bson.D{
		{Key: "$unwind", Value: "$status"},
	}

	projectStage3 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "configChains", Value: "$parsedConfig.chains"},
			{Key: "statusChains", Value: "$status.parsedStatus.chains"},
		}},
	}

	unwindStage4 := bson.D{
		{Key: "$unwind", Value: "$configChains"},
	}

	unwindStage5 := bson.D{
		{Key: "$unwind", Value: "$statusChains"},
	}

	matchStage6 := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{
				{Key: "$eq", Value: bson.A{"$configChains.chainid", "$statusChains.chainid"}},
			}},
		}},
	}

	groupStage7 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$configChains.chainid"},
			{Key: "notionalLimits", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "notionalLimit", Value: "$configChains.notionallimit"},
					{Key: "maxTransactionSize", Value: "$configChains.bigtransactionsize"},
					{Key: "availableNotional", Value: "$statusChains.remainingavailablenotional"},
				}},
			}},
		}},
	}

	projectStage8 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "governorLimit", Value: bson.D{
				{Key: "$sortArray", Value: bson.D{
					{Key: "input", Value: "$notionalLimits"},
					{Key: "sortBy", Value: bson.D{
						{Key: "availableNotional", Value: 1},
					}},
				}},
			}},
		}},
	}

	projectStage9 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "governorLimit", Value: bson.M{
				"$arrayElemAt": []interface{}{"$governorLimit", 0},
			}},
		}},
	}

	projectStage10 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainId", Value: "$_id"},
			{Key: "notionalLimit", Value: "$governorLimit.notionalLimit"},
			{Key: "maxTransactionSize", Value: "$governorLimit.maxTransactionSize"},
			{Key: "availableNotional", Value: "$governorLimit.availableNotional"},
		}},
	}

	sortStage11 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "chainId", Value: 1},
		}},
	}

	// define aggregate pipeline
	pipeLine := mongo.Pipeline{
		lookupStage1,
		unwindStage2,
		projectStage3,
		unwindStage4,
		unwindStage5,
		matchStage6,
		groupStage7,
		projectStage8,
		projectStage9,
		projectStage10,
		sortStage11,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorConfig.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get governor limit",
			zap.Error(err), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decodes to GovernorLimitV2.
	var availbleNotional []*AvailableNotionalByChain
	err = cur.All(ctx, &availbleNotional)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*AvailableNotionalByChain", zap.Error(err),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// check exists records
	if len(availbleNotional) == 0 {
		return nil, errs.ErrNotFound
	}

	return availbleNotional, nil
}

// GetTokenList get token lists.
func (r *Repository) GetTokenList(ctx context.Context) ([]*TokenList, error) {
	projectStage1 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "tokens", Value: "$parsedConfig.tokens"},
		}},
	}
	unwindStage2 := bson.D{
		{Key: "$unwind", Value: "$tokens"},
	}
	projectStage3 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "originaddress", Value: "$tokens.originaddress"},
			{Key: "originchainid", Value: "$tokens.originchainid"},
			{Key: "price", Value: "$tokens.price"},
		}},
	}
	groupStage4 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "originaddress", Value: "$originaddress"},
				{Key: "originchainid", Value: "$originchainid"},
			}},
			{Key: "prices", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "price", Value: "$price"},
				}},
			}},
		}},
	}
	unwindStage5 := bson.D{
		{Key: "$unwind", Value: "$prices"},
	}

	groupStage6 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "_id", Value: "$_id"},
				{Key: "prices", Value: "$prices"},
			}},
			{Key: "count", Value: bson.D{
				{Key: "$sum", Value: 1},
			}},
		}},
	}
	sortStage7 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "_id._id.originchainid", Value: 1},
			{Key: "_id._id.originaddress", Value: 1},
			{Key: "count", Value: -1},
		}},
	}
	groupStage8 := bson.D{
		{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "originchainid", Value: "$_id._id.originchainid"},
				{Key: "originaddress", Value: "$_id._id.originaddress"},
			}},
			{Key: "price", Value: bson.D{
				{Key: "$first", Value: "$_id.prices"},
			}},
			{Key: "count", Value: bson.D{
				{Key: "$first", Value: "$count"},
			}},
		}},
	}
	sortStage9 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "_id.originchainid", Value: 1},
			{Key: "_id.originaddress", Value: 1},
		}},
	}
	projectStage10 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "originchainid", Value: "$_id.originchainid"},
			{Key: "originaddress", Value: "$_id.originaddress"},
			{Key: "price", Value: "$price.price"},
		}},
	}

	// define aggregate pipeline
	pipeLine := mongo.Pipeline{
		projectStage1,
		unwindStage2,
		projectStage3,
		groupStage4,
		unwindStage5,
		groupStage6,
		sortStage7,
		groupStage8,
		sortStage9,
		projectStage10,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorConfig.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get token list",
			zap.Error(err), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decodes to RawDocRecord.
	var tokens []*TokenList
	err = cur.All(ctx, &tokens)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*TokenList", zap.Error(err),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// check exists records
	if len(tokens) == 0 {
		return nil, errs.ErrNotFound
	}

	return tokens, nil
}

// GetEnqueuedVaas get enqueued vaas.
func (r *Repository) GetEnqueuedVaas(ctx context.Context) ([]*EnqueuedVaaItem, error) {
	projectStage1 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chains", Value: "$parsedStatus.chains"},
		}},
	}

	unwindStage2 := bson.D{
		{Key: "$unwind", Value: "$chains"},
	}

	matchStage3 := bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{
				{Key: "$gt", Value: bson.A{"chains.emitters.totalenqueuedvaas", 0}},
			}},
		}},
	}

	projectStage4 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainid", Value: "$chains.chainid"},
			{Key: "emitters", Value: "$chains.emitters"},
		}},
	}

	unwindStage5 := bson.D{
		{Key: "$unwind", Value: "$emitters"},
	}

	projectStage6 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainid", Value: "$chainid"},
			{Key: "emitteraddress", Value: "$emitters.emitteraddress"},
			{Key: "enqueuedvaas", Value: "$emitters.enqueuedvaas"},
		}},
	}

	unwindStage7 := bson.D{
		{Key: "$unwind", Value: "$enqueuedvaas"},
	}

	projectStage8 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainid", Value: 1},
			{Key: "emitteraddress", Value: 1},
			{Key: "notionalvalue", Value: "$enqueuedvaas.notionalvalue"},
			{Key: "releasetime", Value: "$enqueuedvaas.releasetime"},
			{Key: "sequence", Value: "$enqueuedvaas.sequence"},
			{Key: "txhash", Value: "$enqueuedvaas.txhash"},
		}},
	}

	sortStage9 := bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "chainId", Value: 1},
			{Key: "emitteraddress", Value: 1},
			{Key: "sequence", Value: 1},
		}},
	}

	// define aggregate pipeline
	pipeLine := mongo.Pipeline{
		projectStage1,
		unwindStage2,
		matchStage3,
		projectStage4,
		unwindStage5,
		projectStage6,
		unwindStage7,
		projectStage8,
		sortStage9,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorStatus.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get enqueuedVAA",
			zap.Error(err), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	// decodes to []*EnqueuedVaaItem.
	var enqueuedVAA []*EnqueuedVaaItem
	err = cur.All(ctx, &enqueuedVAA)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*EnqueuedVaaItem", zap.Error(err),
			zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}

	return enqueuedVAA, nil
}

type EnqueuedResponse struct {
	ChainID        int64  `bson:"chainid" json:"chainID"`
	EmitterAddress string `bson:"emitteraddress" json:"emitterAddress"`
	Sequence       string `bson:"sequence" json:"sequence"`
}

// IsVaaEnqueued check vaa is enqueued.
func (r *Repository) IsVaaEnqueued(ctx context.Context, chainID vaa.ChainID, emitter vaa.Address, sequence string) (bool, error) {
	projectStage1 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chains", Value: "$parsedStatus.chains"},
		}},
	}
	unwindStage2 := bson.D{
		{Key: "$unwind", Value: "$chains"},
	}
	projectStage3 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainid", Value: "$chains.chainid"},
			{Key: "availableNotion", Value: "$chains.remainingavailablenotional"},
			{Key: "emitters", Value: "$chains.emitters"},
		}},
	}
	unwindStage4 := bson.D{
		{Key: "$unwind", Value: "$emitters"},
	}
	unwindStage5 := bson.D{
		{Key: "$unwind", Value: "$emitters.enqueuedvaas"},
	}

	matchStage6 := bson.D{
		{"$match", bson.D{
			{"chainid", chainID},
			{"emitters.emitteraddress", fmt.Sprintf("0x%s", emitter.String())},
			{"emitters.enqueuedvaas.sequence", sequence},
		}},
	}

	projectStage7 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "chainid", Value: "$chainid"},
			{Key: "emitteraddress", Value: "$emitters.emitteraddress"},
			{Key: "sequence", Value: "$emitters.enqueuedvaas.sequence"},
		}},
	}

	// define aggregate pipeline
	pipeLine := mongo.Pipeline{
		projectStage1,
		unwindStage2,
		projectStage3,
		unwindStage4,
		unwindStage5,
		matchStage6,
		projectStage7,
	}

	// execute aggregate operations.
	cur, err := r.collections.governorStatus.Aggregate(ctx, pipeLine)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute Aggregate command to get token list",
			zap.Error(err), zap.String("requestID", requestID))
		return false, errors.WithStack(err)
	}

	// decodes to RawDocRecord.
	var response []*EnqueuedResponse
	err = cur.All(ctx, &response)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed decoding cursor to []*EnqueuedResponse", zap.Error(err),
			zap.String("requestID", requestID))
		return false, errors.WithStack(err)
	}

	// check exists records
	if len(response) == 0 {
		return false, nil
	}

	return true, nil
}
