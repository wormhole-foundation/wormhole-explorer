package transactions

import (
	"context"
	"fmt"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const queryTemplate = `
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r._measurement == "vaa_volume" and r._field == "volume")
  |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
  |> %s(column: "volume")
`

const queryTemplateWithApps = `
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r._measurement == "vaa_volume")
  |> filter(fn: (r) => r._field == "volume" or  r._field == "app_id")
  |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
  |> filter(fn: (r) => contains(value: r.app_id, set: %s))
  |> %s(column: "volume")
`

const queryTemplateVaaCount = `
from(bucket: "%s")
  |> range(start: -%s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count")
  |> group()
  |> aggregateWindow(every: %s, fn: count, createEmpty: true)
  |> map(fn:(r) => ( {_time: r._time, count: r._value}))
`

const queryTemplateTotalTxCount = `
from(bucket: "%s")
  |> range(start: 2018-01-01T00:00:00Z)
  |> filter(fn: (r) => r._field == "total_vaa_count")
  |> last()
`

const queryTemplateTxCount24h = `
from(bucket: "%s")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "vaa_count")
  |> group(columns: ["_measurement"])
  |> count()
`

const queryTemplateMessages24h = `
from(bucket: "%s")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "vaa_count_including_pyth")
  |> group(columns: ["_measurement"])
  |> count()
`

type Repository struct {
	influxCli   influxdb2.Client
	queryAPI    api.QueryAPI
	bucket      string
	db          *mongo.Database
	collections struct {
		globalTransactions *mongo.Collection
	}
	logger *zap.Logger
}

func NewRepository(client influxdb2.Client, org, bucket string, db *mongo.Database, logger *zap.Logger) *Repository {
	queryAPI := client.QueryAPI(org)
	return &Repository{influxCli: client,
		queryAPI:    queryAPI,
		bucket:      bucket,
		db:          db,
		collections: struct{ globalTransactions *mongo.Collection }{globalTransactions: db.Collection("globalTransactions")},
		logger:      logger}
}

func (r *Repository) FindChainActivity(ctx context.Context, q *ChainActivityQuery) ([]ChainActivityResult, error) {
	query := r.buildFindVolumeQuery(q)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var response []ChainActivityResult
	for result.Next() {
		var row ChainActivityResult
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		response = append(response, row)
	}
	return response, nil
}

func (r *Repository) buildFindVolumeQuery(q *ChainActivityQuery) string {
	start := q.GetStart().UTC().Format(time.RFC3339)
	stop := q.GetEnd().UTC().Format(time.RFC3339)
	var operation string
	if q.IsNotional {
		operation = "sum"
	} else {
		operation = "count"
	}
	if q.HasAppIDS() {
		apps := `["` + strings.Join(q.GetAppIDs(), `","`) + `"]`
		return fmt.Sprintf(queryTemplateWithApps, r.bucket, start, stop, apps, operation)
	}
	return fmt.Sprintf(queryTemplate, r.bucket, start, stop, operation)
}

func (r *Repository) GetScorecards(ctx context.Context) (*Scorecards, error) {

	totalTxCount, err := r.getTotalTxCount(ctx)
	if err != nil {
		r.logger.Error("failed to query total transaction count", zap.Error(err))
	}

	txCount24h, err := r.getTxCount24h(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query 24h transactions: %w", err)
	}

	messages24h, err := r.getMessages24h(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query 24h messages: %w", err)
	}

	// build the result and return
	scorecards := Scorecards{
		TotalTxCount: totalTxCount,
		TxCount24h:   txCount24h,
		Messages24h:  messages24h,
	}

	return &scorecards, nil
}

func (r *Repository) getTotalTxCount(ctx context.Context) (string, error) {

	// query 24h transactions
	query := fmt.Sprintf(queryTemplateTotalTxCount, r.bucket)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query total transaction count", zap.Error(err))
		return "", err
	}
	if result.Err() != nil {
		r.logger.Error("total transaction count query result has errors", zap.Error(err))
		return "", result.Err()
	}
	if !result.Next() {
		return "", errors.New("expected at least one record in total transaction count query result")
	}

	// deserialize the row returned
	row := struct {
		Value uint64 `mapstructure:"_value"`
	}{}
	if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return "", fmt.Errorf("failed to decode total transaction count query response: %w", err)
	}

	return fmt.Sprint(row.Value), nil
}

func (r *Repository) getTxCount24h(ctx context.Context) (string, error) {

	// query 24h transactions
	query := fmt.Sprintf(queryTemplateTxCount24h, r.bucket)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query 24h transactions", zap.Error(err))
		return "", err
	}
	if result.Err() != nil {
		r.logger.Error("24h transactions query result has errors", zap.Error(err))
		return "", result.Err()
	}
	if !result.Next() {
		return "", errors.New("expected at least one record in 24h transactions query result")
	}

	// deserialize the row returned
	row := struct {
		Value uint64 `mapstructure:"_value"`
	}{}
	if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return "", fmt.Errorf("failed to decode 24h transaction count query response: %w", err)
	}

	return fmt.Sprint(row.Value), nil
}

func (r *Repository) getMessages24h(ctx context.Context) (string, error) {

	// query 24h messages
	query := fmt.Sprintf(queryTemplateMessages24h, r.bucket)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query 24h messages", zap.Error(err))
		return "", err
	}
	if result.Err() != nil {
		r.logger.Error("24h messages query result has errors", zap.Error(err))
		return "", result.Err()
	}
	if !result.Next() {
		return "", errors.New("expected at least one record in 24h messages query result")
	}

	// deserialize the row returned
	row := struct {
		Value uint64 `mapstructure:"_value"`
	}{}
	if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return "", fmt.Errorf("failed to decode 24h message count query response: %w", err)
	}

	return fmt.Sprint(row.Value), nil
}

// GetTransactionCount get the last transactions.
func (r *Repository) GetTransactionCount(ctx context.Context, q *TransactionCountQuery) ([]TransactionCountResult, error) {
	query := r.buildLastTrxQuery(q)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	response := []TransactionCountResult{}
	for result.Next() {
		var row TransactionCountResult
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		response = append(response, row)
	}
	return response, nil
}

func (r *Repository) buildLastTrxQuery(q *TransactionCountQuery) string {
	return fmt.Sprintf(queryTemplateVaaCount, r.bucket, q.TimeSpan, q.SampleRate)
}

func (r *Repository) FindGlobalTransactionByID(ctx context.Context, q GlobalTransactionQuery) (*GlobalTransactionDoc, error) {
	var globalTranstaction GlobalTransactionDoc
	err := r.db.Collection("globalTransactions").FindOne(ctx, bson.M{"_id": q.id}).Decode(&globalTranstaction)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		r.logger.Error("failed execute FindOne command to get global transaction",
			zap.Error(err), zap.Any("q", q), zap.String("requestID", requestID))
		return nil, errors.WithStack(err)
	}
	return &globalTranstaction, nil
}
