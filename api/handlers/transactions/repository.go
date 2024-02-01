package transactions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/common"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/tvl"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const queryTemplateChainActivity = `
from(bucket: "%s")
  |> range(start: %s)
  |> filter(fn: (r) => r._measurement == "%s" and r._field == "%s")
  |> last()
  |> group(columns: ["emitter_chain", "destination_chain"])
  |> sum()
`

const queryTemplateChainActivityWithApps = `
from(bucket: "%s")
  |> range(start: %s)
  |> filter(fn: (r) => r._measurement == "%s" and r._field == "%s")
  |> filter(fn: (r) => contains(value: r.app_id, set: %s))
  |> last()
  |> group(columns: ["emitter_chain", "destination_chain"])
  |> sum()
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
  |> filter(fn: (r) => r._measurement == "vaa_volume_v2")
  |> filter(fn:(r) => r._field == "volume")
  |> group()
  |> sum(column: "_value")
`

const queryTemplateMessages24h = `
import "date"

// Get historic count from the summarized metric.
summarized = from(bucket: "%s")
  |> range(start: -24h)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count_all_messages_5m")
  |> group()
  |> sum()

// Get the current count from the unsummarized metric.
// This assumes that the summarization task runs exactly every 5 minutes
startOfInterval = date.truncate(t: now(), unit: 5m)
raw = from(bucket: "%s")
  |> range(start: startOfInterval)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count_all_messages")
  |> filter(fn: (r) => r["_field"] == "count")
  |> group()
  |> count()

// Merge all results, compute the sum, return the top 7 volumes.
union(tables: [summarized, raw])
  |> group()
  |> sum()
`

const queryTemplateTopAssets = `
import "date"

// Get historic volumes from the summarized metric.
summarized = from(bucket: "%s")
  |> range(start: -%s)
  |> filter(fn: (r) => r["_measurement"] == "asset_volumes_24h_v2")
  |> group(columns: ["emitter_chain", "token_address", "token_chain"])

// Get the current day's volume from the unsummarized metric.
// This assumes that the summarization task runs exactly once per day at 00:00hs
startOfDay = date.truncate(t: now(), unit: 1d)
raw = from(bucket: "%s")
  |> range(start: startOfDay)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume_v2")
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

from(bucket: "%s")
  |> range(start: -%s)
  |> filter(fn: (r) => r._measurement == "%s" and r._field == "count")
  |> last()
  |> group(columns: ["emitter_chain", "destination_chain"])
  |> sum()
  |> group()
  |> top(columns: ["_value"], n: 100)
`

type repositoryCollections struct {
	vaas               *mongo.Collection
	vaasPythnet        *mongo.Collection
	parsedVaa          *mongo.Collection
	globalTransactions *mongo.Collection
}

type Repository struct {
	tvl                     *tvl.Tvl
	influxCli               influxdb2.Client
	queryAPI                api.QueryAPI
	bucketInfiniteRetention string
	bucket30DaysRetention   string
	bucket24HoursRetention  string
	db                      *mongo.Database
	collections             repositoryCollections
	supportedChainIDs       map[sdk.ChainID]string
	logger                  *zap.Logger
}

func NewRepository(
	tvl *tvl.Tvl,
	client influxdb2.Client,
	org string,
	bucket24HoursRetention, bucket30DaysRetention, bucketInfiniteRetention string,
	db *mongo.Database,
	logger *zap.Logger,
) *Repository {

	r := Repository{
		tvl:                     tvl,
		influxCli:               client,
		queryAPI:                client.QueryAPI(org),
		bucket24HoursRetention:  bucket24HoursRetention,
		bucket30DaysRetention:   bucket30DaysRetention,
		bucketInfiniteRetention: bucketInfiniteRetention,
		db:                      db,
		collections: repositoryCollections{
			vaas:               db.Collection("vaas"),
			vaasPythnet:        db.Collection("vaasPythnet"),
			parsedVaa:          db.Collection("parsedVaa"),
			globalTransactions: db.Collection("globalTransactions"),
		},
		supportedChainIDs: domain.GetSupportedChainIDs(),
		logger:            logger,
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

	if timeSpan == nil {
		return nil, fmt.Errorf("invalid nil timeSpan")
	}

	var measurement string
	switch *timeSpan {
	case TimeSpan7Days:
		measurement = "chain_activity_7_days_3h_v2"
	case TimeSpan15Days:
		measurement = "chain_activity_15_days_3h_v2"
	case TimeSpan30Days:
		measurement = "chain_activity_30_days_3h_v2"
	}

	// Submit the query to InfluxDB
	query := fmt.Sprintf(queryTemplateTopChainPairs, r.bucket24HoursRetention, *timeSpan, measurement)
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
	var pairs []ChainPairDTO
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
		pair := ChainPairDTO{
			EmitterChain:      sdk.ChainID(emitterChain),
			DestinationChain:  sdk.ChainID(destinationChain),
			NumberOfTransfers: fmt.Sprintf("%d", rows[i].NumberOfTransfers),
		}

		// do not include invalid chain IDs in the response
		if !domain.ChainIdIsValid(pair.EmitterChain) || !domain.ChainIdIsValid(pair.DestinationChain) {
			continue
		}

		pairs = append(pairs, pair)

		// max number of elements
		if len(pairs) == 7 {
			break
		}
	}

	return pairs, nil
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
	query := r.buildChainActivityQuery(q)
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

	// https://github.com/wormhole-foundation/wormhole-explorer/issues/433
	// filter out results with wrong chain ids
	// this should be fixed in the InfluxDB
	var responseWithoutWrongChainId []ChainActivityResult
	for _, res := range response {
		chainSourceID, err := strconv.Atoi(res.ChainSourceID)
		if err != nil {
			continue
		}
		if _, ok := r.supportedChainIDs[sdk.ChainID(chainSourceID)]; !ok {
			continue
		}
		chainDestinationID, err := strconv.Atoi(res.ChainDestinationID)
		if err != nil {
			continue
		}
		if _, ok := r.supportedChainIDs[sdk.ChainID(chainDestinationID)]; !ok {
			continue
		}
		responseWithoutWrongChainId = append(responseWithoutWrongChainId, res)
	}
	return responseWithoutWrongChainId, nil
}

func (r *Repository) buildChainActivityQuery(q *ChainActivityQuery) string {

	var field string
	if q.IsNotional {
		field = "notional"
	} else {
		field = "count"
	}
	var measurement string
	switch q.TimeSpan {
	case ChainActivityTs7Days:
		measurement = "chain_activity_7_days_3h_v2"
	case ChainActivityTs30Days:
		measurement = "chain_activity_30_days_3h_v2"
	case ChainActivityTs90Days:
		measurement = "chain_activity_90_days_3h_v2"
	case ChainActivityTs1Year:
		measurement = "chain_activity_1_year_3h_v2"
	case ChainActivityTsAllTime:
		measurement = "chain_activity_all_time_3h_v2"
	default:
		measurement = "chain_activity_7_days_3h_v2"
	}
	//today without hours
	start := time.Now().Truncate(24 * time.Hour).UTC().Format(time.RFC3339)
	if q.HasAppIDS() {
		apps := `["` + strings.Join(q.GetAppIDs(), `","`) + `"]`
		return fmt.Sprintf(queryTemplateChainActivityWithApps, r.bucket24HoursRetention, start, measurement, field, apps)
	} else {
		return fmt.Sprintf(queryTemplateChainActivity, r.bucket24HoursRetention, start, measurement, field)
	}
}

func (r *Repository) GetScorecards(ctx context.Context) (*Scorecards, error) {

	// This function launches one goroutine for each scorecard.
	//
	// We use a `sync.WaitGroup` to block until all goroutines are done.
	var wg sync.WaitGroup

	var messages24h, tvl, totalTxCount, totalTxVolume, txCount24h, volume24h, totalPythMessage string

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		messages24h, err = r.getMessages24h(ctx)
		if err != nil {
			r.logger.Error("failed to query 24h messages", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		tvl, err = r.tvl.Get(ctx)
		if err != nil {
			r.logger.Error("failed to get tvl", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		totalTxCount, err = r.getTotalTxCount(ctx)
		if err != nil {
			r.logger.Error("failed to tx count", zap.Error(err))
		}

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		totalPythMessage, err = r.getTotalPythMessage(ctx)
		if err != nil {
			r.logger.Error("failed to get total pyth message", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		totalTxVolume, err = r.getTotalTxVolume(ctx)
		if err != nil {
			r.logger.Error("failed to get total tx volume", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		txCount24h, err = r.getTxCount24h(ctx)
		if err != nil {
			r.logger.Error("failed to get 24h transactions", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		volume24h, err = r.getVolume24h(ctx)
		if err != nil {
			r.logger.Error("failed to get 24h volume", zap.Error(err))
		}
	}()

	// Each of the queries synchronized by this wait group has a context timeout.
	//
	// Hence, this call to `wg.Wait()` will not block indefinitely as long as the
	// context timeouts are properly handled in each goroutine.
	wg.Wait()

	// totalPythMessagelegacyEmitter contain the last sequence for the legacy pyth emitter address
	// last vaa ==> 26/f8cd23c2ab91237730770bbea08d61005cdda0984348f3f6eecb559638c0bba0/965463498
	var totalPythMessagelegacyEmitter uint64 = 965463498
	uTotalTxCount, err := strconv.ParseUint(totalTxCount, 10, 64)
	if err != nil {
		uTotalTxCount = 0
	}
	uTotalPyth, err := strconv.ParseUint(totalPythMessage, 10, 64)
	if err != nil {
		uTotalPyth = 0
	}
	totalMessage := totalPythMessagelegacyEmitter + uTotalTxCount + uTotalPyth

	// Build the result and return
	scorecards := Scorecards{
		Messages24h:   messages24h,
		TotalMessages: strconv.FormatUint(totalMessage, 10),
		TotalTxCount:  totalTxCount,
		TotalTxVolume: totalTxVolume,
		Tvl:           tvl,
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

func (r *Repository) getMessages24h(ctx context.Context) (string, error) {

	// query 24h transactions
	query := fmt.Sprintf(queryTemplateMessages24h, r.bucket24HoursRetention, r.bucket24HoursRetention)
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

	// [QA] The transaction history graph shows the current data twice when filtered by 1W
	// https://github.com/wormhole-foundation/wormhole-explorer/issues/406
	for i := range response {
		if i > 0 {
			if q.TimeSpan == "1w" || q.TimeSpan == "1mo" {
				response[i].Time = response[i].Time.AddDate(0, 0, -1)
			} else if q.TimeSpan == "1d" {
				response[i].Time = response[i].Time.Add(-1 * time.Hour)
			}
		}
	}

	return response, nil
}

// getTotalPythMessage returns the last sequence for the pyth emitter address
func (r *Repository) getTotalPythMessage(ctx context.Context) (string, error) {
	pythEmitterAddr := "f8cd23c2ab91237730770bbea08d61005cdda0984348f3f6eecb559638c0bba0"
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
		Status: string(domain.SourceTxStatusConfirmed),
	}
	if record.EmitterChain != sdk.ChainIDSolana && record.EmitterChain != sdk.ChainIDAptos {
		originTx.TxHash = record.TxHash
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

// FindTransactionsInput is used to pass parameters to the `FindTransactions` method.
type FindTransactionsInput struct {
	// id specifies the VAA ID of the transaction to be found.
	id string
	// sort specifies whether the results should be sorted
	//
	// If set to true, the results will be sorted by descending timestamp and ID.
	// If set to false, the results will not be sorted.
	sort       bool
	pagination *pagination.Pagination
}

// FindTransactions returns transactions matching a specified search criteria.
func (r *Repository) FindTransactions(
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

// ListTransactionsByAddress returns a sorted list of transactions for a given address.
//
// Pagination is implemented using a keyset cursor pattern, based on the (timestamp, ID) pair.
func (r *Repository) ListTransactionsByAddress(
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
