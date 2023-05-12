package transactions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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

const queryTemplateTxCount24h = `
from(bucket: "%s")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "vaa_count")
  |> group(columns: ["_measurement"])
  |> count()
`

const queryTemplateVolume24h = `
from(bucket: "%s")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "vaa_volume")
  |> filter(fn:(r) => r._field == "volume")
  |> drop(columns: ["_measurement", "app_id", "destination_address", "destination_chain", "token_address", "token_chain"])
  |> sum(column: "_value")
`

const queryTemplateTopAssets = `
import "date"

// Get historic volumes from the summarized metric.
summarized = from(bucket: "%s")
  |> range(start: -%s)
  |> filter(fn: (r) => r["_measurement"] == "asset_volumes_24h")
  |> group(columns: ["emitter_chain", "token_address", "token_chain"])

// Get the current day's volume from the unsummarized metric.
// This assumes that the summarization task runs exactly once per day at 00:00hs
startOfDay = date.truncate(t: now(), unit: 1d)
raw = from(bucket: "%s")
  |> range(start: startOfDay)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group(columns: ["emitter_chain", "token_address", "token_chain"])

// Merge all results, compute the sum, return the top 7 volumes.
union(tables: [summarized, raw])
  |> group(columns: ["emitter_chain", "token_address", "token_chain"])
  |> sum()
  |> group()
  |> top(columns: ["_value"], n: 7)
`

const queryTemplateTopChainPairs = `
import "date"

// Get historic number of transfers from the summarized metric.
summarized = from(bucket: "%s")
  |> range(start: -%s)
  |> filter(fn: (r) => r["_measurement"] == "chain_pair_transfers_24h")
  |> group(columns: ["emitter_chain", "destination_chain"])
  |> sum()

// Get the current day's number of transfers from the unsummarized metric.
// This assumes that the summarization task runs exactly once per day at 00:00hs
startOfDay = date.truncate(t: now(), unit: 1d)
raw = from(bucket: "%s")
  |> range(start: startOfDay)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group(columns: ["emitter_chain", "destination_chain"])
  |> drop(columns: ["app_id", "destination_address", "token_address", "token_chain", "_field"])
  |> group(columns: ["emitter_chain", "destination_chain"])
  |> count()

// Merge all results, compute the sum, return the top 7 volumes.
union(tables: [summarized, raw])
  |> group(columns: ["emitter_chain", "destination_chain"])
  |> sum()
  |> group()
  |> top(columns: ["_value"], n: 7)
`

type Repository struct {
	influxCli               influxdb2.Client
	queryAPI                api.QueryAPI
	bucketInfiniteRetention string
	bucket30DaysRetention   string
	bucket24HoursRetention  string
	db                      *mongo.Database
	collections             struct {
		globalTransactions *mongo.Collection
	}
	logger *zap.Logger
}

func NewRepository(
	client influxdb2.Client,
	org string,
	bucket24HoursRetention, bucket30DaysRetention, bucketInfiniteRetention string,
	db *mongo.Database,
	logger *zap.Logger,
) *Repository {

	r := Repository{
		influxCli:               client,
		queryAPI:                client.QueryAPI(org),
		bucket24HoursRetention:  bucket24HoursRetention,
		bucket30DaysRetention:   bucket30DaysRetention,
		bucketInfiniteRetention: bucketInfiniteRetention,
		db:                      db,
		collections:             struct{ globalTransactions *mongo.Collection }{globalTransactions: db.Collection("globalTransactions")},
		logger:                  logger,
	}

	return &r
}

func (r *Repository) GetTopAssets(ctx context.Context, timeSpan *TopStatisticsTimeSpan) ([]AssetDTO, error) {

	// Submit the query to InfluxDB
	query := fmt.Sprintf(queryTemplateTopAssets, r.bucket30DaysRetention, *timeSpan, r.bucketInfiniteRetention)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	// Scan query results
	type Row struct {
		EmitterChain string `mapstructure:"emitter_chain"`
		TokenChain   string `mapstructure:"token_chain"`
		TokenAddress string `mapstructure:"token_address"`
		Volume       uint64 `mapstructure:"_value"`
	}
	var rows []Row
	for result.Next() {
		var row Row
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	// Convert the rows into the response model
	var assets []AssetDTO
	for i := range rows {

		// parse emitter chain
		emitterChain, err := strconv.ParseUint(rows[i].EmitterChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert emitter chain field to uint16")
		}

		// parse token chain
		tokenChain, err := strconv.ParseUint(rows[i].TokenChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert token chain field to uint16")
		}

		// append the new item to the response
		asset := AssetDTO{
			EmitterChain: sdk.ChainID(emitterChain),
			TokenChain:   sdk.ChainID(tokenChain),
			TokenAddress: rows[i].TokenAddress,
			Volume:       convertToDecimal(rows[i].Volume),
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

func (r *Repository) GetTopChainPairs(ctx context.Context, timeSpan *TopStatisticsTimeSpan) ([]ChainPairDTO, error) {

	// Submit the query to InfluxDB
	query := fmt.Sprintf(queryTemplateTopChainPairs, r.bucket30DaysRetention, *timeSpan, r.bucketInfiniteRetention)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	// Scan query results
	type Row struct {
		EmitterChain      string `mapstructure:"emitter_chain"`
		DestinationChain  string `mapstructure:"destination_chain"`
		NumberOfTransfers int64  `mapstructure:"_value"`
	}
	var rows []Row
	for result.Next() {
		var row Row
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	// Convert the rows into the response model
	var assets []ChainPairDTO
	for i := range rows {

		// parse emitter chain
		emitterChain, err := strconv.ParseUint(rows[i].EmitterChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert emitter chain field to uint16")
		}

		// parse destination chain
		destinationChain, err := strconv.ParseUint(rows[i].DestinationChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert destination chain field to uint16")
		}

		// append the new item to the response
		asset := ChainPairDTO{
			EmitterChain:      sdk.ChainID(emitterChain),
			DestinationChain:  sdk.ChainID(destinationChain),
			NumberOfTransfers: fmt.Sprintf("%d", rows[i].NumberOfTransfers),
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

// convertToDecimal converts an integer amount to a decimal string, with 8 decimals of precision.
func convertToDecimal(amount uint64) string {

	// If the amount is less than 1, just use a format mask.
	if amount < 1_0000_0000 {
		return fmt.Sprintf("0.%08d", amount)
	}

	// If the amount is equal or greater than 1, we need to insert a dot 8 digits from the end.
	s := fmt.Sprintf("%d", amount)
	l := len(s)
	result := s[:l-8] + "." + s[l-8:]

	return result
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
		return fmt.Sprintf(queryTemplateWithApps, r.bucketInfiniteRetention, start, stop, apps, operation)
	}
	return fmt.Sprintf(queryTemplate, r.bucketInfiniteRetention, start, stop, operation)
}

func (r *Repository) GetScorecards(ctx context.Context) (*Scorecards, error) {

	totalTxCount, err := r.getTotalTxCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query total tx count by portal bridge")
	}

	totalTxVolume, err := r.getTotalTxVolume(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query tx volume by portal bridge")
	}

	txCount24h, err := r.getTxCount24h(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query 24h transactions: %w", err)
	}

	volume24h, err := r.getVolume24h(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query 24h volume: %w", err)
	}

	// build the result and return
	scorecards := Scorecards{
		TotalTxCount:  totalTxCount,
		TotalTxVolume: totalTxVolume,
		TxCount24h:    txCount24h,
		Volume24h:     volume24h,
	}

	return &scorecards, nil
}

func (r *Repository) getTotalTxCount(ctx context.Context) (string, error) {

	query := buildTotalTrxCountQuery(r.bucketInfiniteRetention, r.bucket30DaysRetention, time.Now())
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query total tx count by portal bridge", zap.Error(err))
		return "", err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query total tx count by portal bridge result has errors", zap.Error(err))
		return "", result.Err()
	}
	if !result.Next() {
		return "", errors.New("expected at least one record in query total tx count by portal bridge result")
	}
	row := struct {
		Value uint64 `mapstructure:"_value"`
	}{}
	if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return "", fmt.Errorf("failed to decode total tx count by portal bridge query response: %w", err)
	}
	return fmt.Sprintf("%d", row.Value), nil
}

func (r *Repository) getTotalTxVolume(ctx context.Context) (string, error) {

	query := buildTotalTrxVolumeQuery(r.bucketInfiniteRetention, r.bucket30DaysRetention, time.Now())
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query total tx volume by portal bridge", zap.Error(err))
		return "", err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query tx volume by portal bridge result has errors", zap.Error(err))
		return "", result.Err()
	}
	if !result.Next() {
		return "", errors.New("expected at least one record in query tx volume by portal bridge result")
	}
	row := struct {
		Value uint64 `mapstructure:"_value"`
	}{}
	if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return "", fmt.Errorf("failed to decode tx volume by portal bridge query response: %w", err)
	}
	return convertToDecimal(row.Value), nil
}

func (r *Repository) getTxCount24h(ctx context.Context) (string, error) {

	// query 24h transactions
	query := fmt.Sprintf(queryTemplateTxCount24h, r.bucket30DaysRetention)
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

func (r *Repository) getVolume24h(ctx context.Context) (string, error) {

	// query 24h volume
	query := fmt.Sprintf(queryTemplateVolume24h, r.bucketInfiniteRetention)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query 24h volume", zap.Error(err))
		return "", err
	}
	if result.Err() != nil {
		r.logger.Error("24h volume query result has errors", zap.Error(err))
		return "", result.Err()
	}
	if !result.Next() {
		return "", errors.New("expected at least one record in 24h volume query result")
	}

	// deserialize the row returned
	row := struct {
		Value uint64 `mapstructure:"_value"`
	}{}
	if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return "", fmt.Errorf("failed to decode 24h volume count query response: %w", err)
	}

	// convert the volume to a string and return
	volume := convertToDecimal(row.Value)
	return volume, nil
}

// GetTransactionCount get the last transactions.
func (r *Repository) GetTransactionCount(ctx context.Context, q *TransactionCountQuery) ([]TransactionCountResult, error) {
	query := buildLastTrxQuery(r.bucket30DaysRetention, time.Now(), q)
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

func (r *Repository) FindGlobalTransactionByID(ctx context.Context, q *GlobalTransactionQuery) (*GlobalTransactionDoc, error) {

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
			OriginTx: originTx,
		}
	case globalTransaction != nil && globalTransaction.OriginTx == nil:
		result = &GlobalTransactionDoc{
			OriginTx:      originTx,
			DestinationTx: globalTransaction.DestinationTx,
		}
	default:
		result = globalTransaction
		result.OriginTx.Timestamp = originTx.Timestamp
		result.OriginTx.ChainID = originTx.ChainID
	}

	return result, nil

}

// findOriginTxFromVaa uses data from the `vaas` collection to create an `OriginTx`.
func (r *Repository) findOriginTxFromVaa(ctx context.Context, q *GlobalTransactionQuery) (*OriginTx, error) {

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
		Timestamp: &record.Timestamp,
		TxHash:    record.TxHash,
		ChainID:   record.EmitterChain,
		Status:    string(domain.SourceTxStatusConfirmed),
	}
	return &originTx, nil
}

// findGlobalTransactionByID searches the `globalTransactions` collection by ID.
func (r *Repository) findGlobalTransactionByID(ctx context.Context, q *GlobalTransactionQuery) (*GlobalTransactionDoc, error) {

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
