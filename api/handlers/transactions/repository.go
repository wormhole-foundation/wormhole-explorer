package transactions

import (
	"context"
	errors2 "errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/tvl"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/mongo"
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

const queryTemplateVolume = `
from(bucket: "%s")
  |> range(start: -%s)
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

type Repository struct {
	mongoRepo               *MongoRepository
	postgresRepo            *PostgresRepository
	tvl                     getTvl
	p2pNetwork              string
	influxCli               influxdb2.Client
	queryAPI                influxQueryAPI
	bucketInfiniteRetention string
	bucket30DaysRetention   string
	bucket24HoursRetention  string
	supportedChainIDs       map[sdk.ChainID]string
	logger                  *zap.Logger
}

type influxQueryAPI interface {
	Query(ctx context.Context, query string) (influxQueryResult, error)
}

type influxQueryResult interface {
	Err() error
	Next() bool
	Record() *query.FluxRecord
}

type influxAdapter struct {
	influxAPI api.QueryAPI
}

func (i *influxAdapter) Query(ctx context.Context, query string) (influxQueryResult, error) {
	result, err := i.influxAPI.Query(ctx, query)
	return &influxResult{result}, err
}

type influxResult struct {
	result *api.QueryTableResult
}

func (i *influxResult) Err() error {
	return i.result.Err()
}

func (i *influxResult) Next() bool {
	return i.result.Next()
}

func (i *influxResult) Record() *query.FluxRecord {
	return i.result.Record()
}

type getTvl interface {
	Get(ctx context.Context) (string, error)
}

type offset string

const _24h offset = "24h"
const _7d offset = "7d"
const _30d offset = "30d"

func NewRepository(
	tvl *tvl.Tvl,
	p2pNetwork string,
	client influxdb2.Client,
	org string,
	bucket24HoursRetention, bucket30DaysRetention, bucketInfiniteRetention string,
	mongo *mongo.Database,
	db *db.DB,
	logger *zap.Logger,
) *Repository {
	r := Repository{
		mongoRepo:               NewMongoRepository(p2pNetwork, mongo, logger),
		postgresRepo:            NewPostgresRepository(p2pNetwork, db, logger),
		tvl:                     tvl,
		p2pNetwork:              p2pNetwork,
		influxCli:               client,
		queryAPI:                &influxAdapter{client.QueryAPI(org)},
		bucket24HoursRetention:  bucket24HoursRetention,
		bucket30DaysRetention:   bucket30DaysRetention,
		bucketInfiniteRetention: bucketInfiniteRetention,
		supportedChainIDs:       domain.GetSupportedChainIDs(),
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

	if q.TimeSpan == ChainActivityTs1Year || q.TimeSpan == ChainActivityTsAllTime {

		if field == "notional" {
			field = "volume"
		}

		var start string
		measurement := "chain_activity_1d"
		switch q.TimeSpan {
		case ChainActivityTs1Year:
			start = time.Now().AddDate(-1, 0, 0).Format(time.RFC3339)
		case ChainActivityTsAllTime:
			start = "1970-01-01T00:00:00Z"
		default:
			start = "1970-01-01T00:00:00Z"
		}

		hotfixQuery := `
			import "date"

			from(bucket: "%s")
				|> range(start: %s)
				|> filter(fn: (r) => r._measurement == "%s")
				|> filter(fn: (r) => r._field == "%s")
				|> drop(columns:["_time","to","_measurement","app_id"])
				|> group(columns:["emitter_chain","destination_chain"])
				|> sum()
`
		return fmt.Sprintf(hotfixQuery, r.bucketInfiniteRetention, start, measurement, field)

	}

	var measurement string
	switch q.TimeSpan {
	case ChainActivityTs7Days:
		measurement = "chain_activity_7_days_3h_v2"
	case ChainActivityTs30Days:
		measurement = "chain_activity_30_days_3h_v2"
	case ChainActivityTs90Days:
		measurement = "chain_activity_90_days_3h_v2"
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

func (r *Repository) GetScorecards(ctx context.Context, usePostgres bool) (*Scorecards, error) {

	// This function launches one goroutine for each scorecard.
	//
	// We use a `sync.WaitGroup` to block until all goroutines are done.
	var wg sync.WaitGroup

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	var resultErr error
	mutex := &sync.Mutex{}
	collectError := func(err error) {
		mutex.Lock()
		resultErr = errors2.Join(resultErr, err)
		mutex.Unlock()
	}

	handleErr := func(errMsgLog string, err error) {
		if err != nil {
			r.logger.Error(errMsgLog, zap.Error(err))
			collectError(err)
			cancel() // this will signal the rest of goroutines to exit also.
		}
	}

	var messages24h, totalValueLocked, totalTxCount, totalTxVolume, volume24h, volume7d, volume30d, totalPythMessage string

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		messages24h, err = r.getMessages24h(ctxWithCancel)
		handleErr("failed to get 24h messages", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		totalValueLocked, err = r.tvl.Get(ctxWithCancel)
		handleErr("failed to get tvl", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		totalTxCount, err = r.getTotalTxCount(ctxWithCancel)
		handleErr("failed to get total tx count", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		if usePostgres {
			totalPythMessage, err = r.postgresRepo.getTotalPythMessage(ctxWithCancel)
		} else {
			totalPythMessage, err = r.mongoRepo.getTotalPythMessage(ctxWithCancel)
		}
		handleErr("failed to get total pyth message", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		totalTxVolume, err = r.getTotalTxVolume(ctxWithCancel)
		handleErr("failed to get total tx volume", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		volume24h, err = r.getVolume(ctxWithCancel, _24h)
		handleErr("failed to get 24h volume", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		volume7d, err = r.getVolume(ctxWithCancel, _7d)
		handleErr("failed to get 7d volume", err)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		volume30d, err = r.getVolume(ctxWithCancel, _30d)
		handleErr("failed to get 30d volume", err)
	}()

	// Each of the queries synchronized by this wait group has a context timeout.
	//
	// Hence, this call to `wg.Wait()` will not block indefinitely as long as the
	// context timeouts are properly handled in each goroutine.
	wg.Wait()

	totalMessage := calculateTotalMessage(r.p2pNetwork, totalTxCount, totalPythMessage)
	// Build the result and return
	scorecards := Scorecards{
		Messages24h:   messages24h,
		TotalMessages: totalMessage,
		TotalTxCount:  totalTxCount,
		TotalTxVolume: totalTxVolume,
		Tvl:           totalValueLocked,
		Volume24h:     volume24h,
		Volume7d:      volume7d,
		Volume30d:     volume30d,
	}
	return &scorecards, resultErr
}

// calculateTotalMessage calculate the total message from the total tx count and the total pyth message
func calculateTotalMessage(p2pNetwork string, totalTxCount, totalPythMessage string) string {
	var totalPythMessagelegacyEmitter uint64 = 0
	if p2pNetwork == config.P2pMainNet {
		// totalPythMessagelegacyEmitter contain the last sequence for the legacy pyth emitter address
		// last vaa ==> 26/f8cd23c2ab91237730770bbea08d61005cdda0984348f3f6eecb559638c0bba0/965463498
		totalPythMessagelegacyEmitter = 965463498
	} else if p2pNetwork == config.P2pTestNet {
		// totalPythMessagelegacyEmitter contain the last sequence for the legacy pyth emitter address testnet
		// 26/a27839d641b07743c0cb5f68c51f8cd31d2c0762bec00dc6fcd25433ef1ab5b6/6566583
		totalPythMessagelegacyEmitter = 6566583
	}
	uTotalTxCount, err := strconv.ParseUint(totalTxCount, 10, 64)
	if err != nil {
		uTotalTxCount = 0
	}
	uTotalPyth, err := strconv.ParseUint(totalPythMessage, 10, 64)
	if err != nil {
		uTotalPyth = 0
	}
	totalMessage := totalPythMessagelegacyEmitter + uTotalTxCount + uTotalPyth
	return strconv.FormatUint(totalMessage, 10)
}

func (r *Repository) getTotalTxCount(ctx context.Context) (string, error) {

	trxCountQuery := buildTotalTrxCountQuery(r.bucketInfiniteRetention, r.bucket30DaysRetention, time.Now())
	result, err := r.queryAPI.Query(ctx, trxCountQuery)
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

	trxVolumeQuery := buildTotalTrxVolumeQuery(r.bucketInfiniteRetention, r.bucket30DaysRetention, time.Now())
	result, err := r.queryAPI.Query(ctx, trxVolumeQuery)
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
	msg24hrVolumeQuery := buildMessages24HrQuery(r.bucket24HoursRetention)
	result, err := r.queryAPI.Query(ctx, msg24hrVolumeQuery)
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

func buildMessages24HrQuery(bucket24Hr string) string {
	return fmt.Sprintf(queryTemplateMessages24h, bucket24Hr, bucket24Hr)
}

func (r *Repository) getVolume(ctx context.Context, from offset) (string, error) {

	// query volume
	queryVolume := buildVolumeQuery(r.bucketInfiniteRetention, from)
	result, err := r.queryAPI.Query(ctx, queryVolume)
	if err != nil {
		r.logger.Error("failed to query volume", zap.Any("from", from), zap.Error(err))
		return "", err
	}
	if result.Err() != nil {
		r.logger.Error("volume query result has errors", zap.Error(err), zap.Any("from", from))
		return "", result.Err()
	}
	if !result.Next() {
		return "", fmt.Errorf("expected at least one record in %s volume query result", from)
	}

	// deserialize the row returned
	row := struct {
		Value uint64 `mapstructure:"_value"`
	}{}
	if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return "", fmt.Errorf("failed to decode %s volume count query response: %w", from, err)
	}

	// convert the volume to a string and return
	volume := convertToDecimal(row.Value)
	return volume, nil
}

func buildVolumeQuery(bucketInfinite string, from offset) string {
	return fmt.Sprintf(queryTemplateVolume, bucketInfinite, from)
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

// ListTransactionsByAddress returns a sorted list of transactions for a given address.
func (r *Repository) ListTransactionsByAddress(
	ctx context.Context,
	address string,
	pagination *pagination.Pagination,
) ([]TransactionDto, error) {
	return r.mongoRepo.ListTransactionsByAddress(ctx, address, pagination)
}

func (r *Repository) FindApplicationActivity(ctx *fasthttp.RequestCtx, q ApplicationActivityQuery) ([]ApplicationActivityTotalsResult, []ApplicationActivityResult, error) {

	if q.AppId != "" && q.ExclusiveAppID {
		res, err := r.findAppsActivity(ctx, q)
		return nil, res, err
	}

	var totals []ApplicationActivityTotalsResult
	var totalsErr error
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		totals, totalsErr = r.findTotalsAppsActivity(ctx, q)
	}()

	appsActivity, err2 := r.findAppsActivity(ctx, q)
	if err2 != nil {
		return nil, nil, err2
	}

	wg.Wait()
	if totalsErr != nil {
		return nil, nil, totalsErr
	}

	return totals, appsActivity, nil
}

func (r *Repository) FindGlobalTransactionByID(ctx context.Context, q *GlobalTransactionQuery) (*GlobalTransactionDoc, error) {
	return r.mongoRepo.FindGlobalTransactionByID(ctx, q)
}

// FindTransactions returns transactions matching a specified search criteria.
func (r *Repository) FindTransactions(
	ctx context.Context,
	input *FindTransactionsInput,
) ([]TransactionDto, error) {
	return r.mongoRepo.FindTransactions(ctx, input)
}

func (r *Repository) findTotalsAppsActivity(ctx *fasthttp.RequestCtx, q ApplicationActivityQuery) ([]ApplicationActivityTotalsResult, error) {
	query := r.buildTotalsAppActivityQuery(q)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var response []ApplicationActivityTotalsResult
	for result.Next() {
		var row ApplicationActivityTotalsResult
		if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		response = append(response, row)
	}

	return response, nil
}

func (r *Repository) findAppsActivity(ctx *fasthttp.RequestCtx, q ApplicationActivityQuery) ([]ApplicationActivityResult, error) {
	query := r.buildAppActivityQuery(q)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var response []ApplicationActivityResult
	for result.Next() {
		var row ApplicationActivityResult
		if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		response = append(response, row)
	}

	return response, nil
}

func (r *Repository) FindChainActivityTops(ctx *fasthttp.RequestCtx, q ChainActivityTopsQuery) ([]ChainActivityTopResult, error) {
	query := r.buildChainActivityQueryTops(q)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var response []ChainActivityTopResult
	for result.Next() {
		var row ChainActivityTopResult
		if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		parsedTime, errTime := time.Parse(time.RFC3339Nano, row.To)
		if errTime == nil {
			row.To = parsedTime.Format(time.RFC3339)
		}
		response = append(response, row)
	}

	return response, nil
}

func (r *Repository) FindTokensVolume(ctx context.Context) ([]TokenVolume, error) {
	query := `
		import "date"

		from(bucket: "%s")
			|> range(start: -1d)
			|> filter(fn: (r) => r._measurement == "tokens_symbol_volume_all_time")
			|> group(columns:["symbol"])
			|> last()
			|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> group()
			|> sort(columns:["volume"],desc:true)
			|> limit(n:100)
			|> map(fn: (r) => ({r with volume: float(v:r.volume) / 100000000.0 }))
	`
	query = fmt.Sprintf(query, r.bucket24HoursRetention)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var response []TokenVolume
	for result.Next() {
		var row TokenVolume
		if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		response = append(response, row)
	}
	return response, nil
}

func (r *Repository) FindTokenSymbolActivity(ctx context.Context, payload TokenSymbolActivityQuery) ([]TokenSymbolActivityResult, error) {
	query := r.buildTokenSymbolActivityQuery(payload)

	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var response []TokenSymbolActivityResult
	for result.Next() {
		var row TokenSymbolActivityResult
		if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}

		emitterChainID, err := strconv.Atoi(row.EmitterChainStr)
		if err != nil {
			r.logger.Error("failed to convert emitter chain id to int", zap.Error(err))
		}
		destChainID, err := strconv.Atoi(row.DestinationChainStr)
		if err != nil {
			r.logger.Error("failed to convert destination chain id to int", zap.Error(err))
		}
		row.EmitterChain = sdk.ChainID(emitterChainID)
		row.DestinationChain = sdk.ChainID(destChainID)
		response = append(response, row)
	}

	return response, nil
}

func (r *Repository) buildTokenSymbolActivityQuery(q TokenSymbolActivityQuery) string {

	var start, stop string

	switch q.Timespan {
	case Hour:
		stop = q.To.Truncate(1 * time.Hour).UTC().Format(time.RFC3339)
		start = q.From.Truncate(1 * time.Hour).UTC().Format(time.RFC3339)
	case Day:
		start = q.From.Truncate(24 * time.Hour).UTC().Format(time.RFC3339)
		stop = q.To.Truncate(24 * time.Hour).UTC().Format(time.RFC3339)
	case Month:
		start = time.Date(q.From.Year(), q.From.Month(), 1, 0, 0, 0, 0, q.From.Location()).UTC().Format(time.RFC3339)
		stop = time.Date(q.To.Year(), q.To.Month(), 1, 0, 0, 0, 0, q.To.Location()).UTC().Format(time.RFC3339)
	default:
		start = time.Date(q.From.Year(), 1, 1, 0, 0, 0, 0, q.From.Location()).UTC().Format(time.RFC3339)
		stop = time.Date(q.To.Year(), 1, 1, 0, 0, 0, 0, q.To.Location()).UTC().Format(time.RFC3339)
	}

	filterTargetChain := ""
	if len(q.TargetChains) > 0 {
		val := fmt.Sprintf("r.destination_chain == \"%d\"", q.TargetChains[0])
		buff := ""
		for _, tc := range q.TargetChains[1:] {
			buff += fmt.Sprintf(" or r.destination_chain == \"%d\"", tc)
		}
		filterTargetChain = fmt.Sprintf("|> filter(fn: (r) => %s%s)", val, buff)
	}

	filterSourceChain := ""
	if len(q.SourceChains) > 0 {
		val := fmt.Sprintf("r.emitter_chain == \"%d\"", q.SourceChains[0])
		buff := ""
		for _, tc := range q.SourceChains[1:] {
			buff += fmt.Sprintf(" or r.emitter_chain == \"%d\"", tc)
		}
		filterSourceChain = fmt.Sprintf("|> filter(fn: (r) => %s%s)", val, buff)
	}

	filterTokenSymbol := ""
	if len(q.TokenSymbols) > 0 {
		val := fmt.Sprintf("r.symbol == \"%s\"", q.TokenSymbols[0])
		buff := ""
		for _, tc := range q.TokenSymbols[1:] {
			buff += fmt.Sprintf(" or r.symbol == \"%s\"", tc)
		}
		filterTokenSymbol = fmt.Sprintf("|> filter(fn: (r) => %s%s)", val, buff)
	}

	query := `
	import "date"

	sumAndCount = (tables=<-, column) => {
		return tables
				|> reduce(
					identity: {
						_value: uint(v:0),
						txs: uint(v:0)
					},
					fn: (r, accumulator) => ({
						_value: accumulator._value + r._value,
						txs: accumulator.txs + uint(v:1)
					})
				)
	}
	
	from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
		|> filter(fn: (r) => r._field == "volume" or r._field == "symbol")
		|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> keep(columns:["_start","_stop","_time","emitter_chain","destination_chain","symbol","volume"])
		|> filter(fn: (r) => r.volume > 0)
		%s //filter by symbol
		%s //filter by source_chain
		%s //filter by target_chain
		|> rename(columns: {volume: "_value"})
		|> set(key: "_field", value: "volume")
		|> group(columns:["symbol","emitter_chain","destination_chain","_field"])
		|> aggregateWindow(every: %s, fn: sumAndCount, createEmpty: true)
		|> map(fn: (r) => ({
				r with 
				volume: if exists r._value then float(v:r._value) / 100000000.0 else float(v:0),
				to: r._time,
				_time: date.sub(d: %s, from: r._time),
		}))
		|> drop(columns:["_value","_start","_stop","_field"])	
	`

	return fmt.Sprintf(query, r.bucketInfiniteRetention, start, stop, filterTokenSymbol, filterSourceChain, filterTargetChain, q.Timespan, q.Timespan)
}

func (r *Repository) buildChainActivityQueryTops(q ChainActivityTopsQuery) string {

	var start, stop string

	switch q.Timespan {
	case Hour:
		start = q.From.Truncate(1 * time.Hour).UTC().Format(time.RFC3339)
		stop = q.To.Truncate(1 * time.Hour).UTC().Format(time.RFC3339)
	case Day:
		start = q.From.Truncate(24 * time.Hour).UTC().Format(time.RFC3339)
		stop = q.To.Truncate(24 * time.Hour).UTC().Format(time.RFC3339)
	case Month:
		start = time.Date(q.From.Year(), q.From.Month(), 1, 0, 0, 0, 0, q.From.Location()).UTC().Format(time.RFC3339)
		stop = time.Date(q.To.Year(), q.To.Month(), 1, 0, 0, 0, 0, q.To.Location()).UTC().Format(time.RFC3339)
	default:
		start = time.Date(q.From.Year(), 1, 1, 0, 0, 0, 0, q.From.Location()).UTC().Format(time.RFC3339)
		stop = time.Date(q.To.Year(), 1, 1, 0, 0, 0, 0, q.To.Location()).UTC().Format(time.RFC3339)
	}

	filterTargetChain := ""
	if len(q.TargetChains) > 0 {
		val := fmt.Sprintf("r.destination_chain == \"%d\"", q.TargetChains[0])
		buff := ""
		for _, tc := range q.TargetChains[1:] {
			buff += fmt.Sprintf(" or r.destination_chain == \"%d\"", tc)
		}
		filterTargetChain = fmt.Sprintf("|> filter(fn: (r) => %s%s)", val, buff)
	}

	filterSourceChain := ""
	if len(q.SourceChains) > 0 {
		val := fmt.Sprintf("r.emitter_chain == \"%d\"", q.SourceChains[0])
		buff := ""
		for _, tc := range q.SourceChains[1:] {
			buff += fmt.Sprintf(" or r.emitter_chain == \"%d\"", tc)
		}
		filterSourceChain = fmt.Sprintf("|> filter(fn: (r) => %s%s)", val, buff)
	}

	filterAppId := ""
	if q.AppId != "" {
		filterAppId = "|> filter(fn: (r) => r.app_id == \"" + q.AppId + "\")"
	}

	if len(q.TargetChains) == 0 && q.AppId == "" {
		return r.buildQueryChainActivityTopsByEmitter(q, start, stop, filterSourceChain)
	}

	var query string
	switch q.Timespan {
	case Hour:
		query = r.buildQueryChainActivityHourly(start, stop, filterSourceChain, filterTargetChain, filterAppId)
	case Day:
		query = r.buildQueryChainActivityDaily(start, stop, filterSourceChain, filterTargetChain, filterAppId)
	case Month:
		query = r.buildQueryChainActivityMonthly(start, stop, filterSourceChain, filterTargetChain, filterAppId)
	default:
		query = r.buildQueryChainActivityYearly(start, stop, filterSourceChain, filterTargetChain, filterAppId)
	}
	return query
}

func (r *Repository) buildQueryChainActivityTopsByEmitter(q ChainActivityTopsQuery, start, stop, filterSourceChain string) string {

	measurement := ""
	switch q.Timespan {
	case Hour:
		measurement = "emitter_chain_activity_1h"
	default:
		measurement = "emitter_chain_activity_1d"
	}

	if q.Timespan == Hour || q.Timespan == Day {
		query := `
					import "date"

					from(bucket: "%s")
					|> range(start: %s,stop: %s)
					|> filter(fn: (r) => r._measurement == "%s")
					%s
					|> pivot(rowKey:["_time","emitter_chain"], columnKey: ["_field"], valueColumn: "_value")
					|> sort(columns:["emitter_chain","_time"],desc:false)`
		return fmt.Sprintf(query, r.bucketInfiniteRetention, start, stop, measurement, filterSourceChain)
	}

	if q.Timespan == Month {
		query := `
				import "date"
				import "join"

				data = from(bucket: "%s")
						|> range(start: %s,stop: %s)
						|> filter(fn: (r) => r._measurement == "%s")
						%s
						|> drop(columns:["to"])
						|> window(every: 1mo, period:1mo)
						|> drop(columns:["_time"])
						|> rename(columns: {_start: "_time"})
						|> map(fn: (r) => ({r with to: string(v: r._stop)}))

				vols = data
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

				counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

				join.inner(
						left: vols,
						right: counts,
						on: (l, r) => l._time == r._time and l.emitter_chain == r.emitter_chain,
						as: (l, r) => ({l with count: r.count}),
				)
				|> group()
				|> sort(columns:["emitter_chain","_time"],desc:false)`
		return fmt.Sprintf(query, r.bucketInfiniteRetention, start, stop, measurement, filterSourceChain)
	}

	query := `
				import "date"
				import "join"

				data = from(bucket: "%s")
						|> range(start: %s,stop: %s)
						|> filter(fn: (r) => r._measurement == "%s")
						%s
						|> drop(columns:["to"])
						|> window(every: 1y, period:1y)
						|> drop(columns:["_time"])
						|> rename(columns: {_start: "_time"})
						|> map(fn: (r) => ({r with to: string(v: r._stop)}))

				vols = data
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

				counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

				join.inner(
						left: vols,
						right: counts,
						on: (l, r) => l._time == r._time and l.emitter_chain == r.emitter_chain,
						as: (l, r) => ({l with count: r.count}),
				)
				|> group()
				|> sort(columns:["emitter_chain","_time"],desc:false)`
	return fmt.Sprintf(query, r.bucketInfiniteRetention, start, stop, measurement, filterSourceChain)

}

func (r *Repository) buildQueryChainActivityHourly(start, stop, filterSourceChain, filterTargetChain, filterAppId string) string {
	query := `
					import "date"
					import "join"

					data = from(bucket: "%s")
		  			|> range(start: %s,stop: %s)
		  			|> filter(fn: (r) => r._measurement == "chain_activity_1h")
					%s
					%s
					%s
					|> drop(columns:["destination_chain"])

					vols = data		
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

					counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

					join.inner(
					    left: vols,
					    right: counts,
					    on: (l, r) => l._time == r._time and l.to == r.to and l.emitter_chain == r.emitter_chain,
					    as: (l, r) => ({l with count: r.count}),
					)
					|> group()
					|> sort(columns:["emitter_chain","_time"],desc:false)`
	return fmt.Sprintf(query, r.bucketInfiniteRetention, start, stop, filterSourceChain, filterTargetChain, filterAppId)
}

func (r *Repository) buildQueryChainActivityDaily(start, stop, filterSourceChain, filterTargetChain, filterAppId string) string {

	query := `
					import "date"
					import "join"

					data = from(bucket: "%s")
		  			|> range(start: %s,stop: %s)
		  			|> filter(fn: (r) => r._measurement == "chain_activity_1d")
					%s
					%s
					%s
					|> drop(columns:["destination_chain"])

					vols = data		
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

					counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

					join.inner(
					    left: vols,
					    right: counts,
					    on: (l, r) => l._time == r._time and l.to == r.to and l.emitter_chain == r.emitter_chain,
					    as: (l, r) => ({l with count: r.count}),
					)
					|> group()
					|> sort(columns:["emitter_chain","_time"],desc:false)`
	return fmt.Sprintf(query, r.bucketInfiniteRetention, start, stop, filterSourceChain, filterTargetChain, filterAppId)
}

func (r *Repository) buildQueryChainActivityMonthly(start, stop, filterSourceChain, filterTargetChain, filterAppId string) string {
	query := `
				import "date"
				import "join"

				data = from(bucket: "%s")
						|> range(start: %s,stop: %s)
						|> filter(fn: (r) => r._measurement == "chain_activity_1d")
						%s
						%s
						%s
						|> drop(columns:["destination_chain","to","app_id"])
						|> window(every: 1mo, period:1mo)
						|> drop(columns:["_time"])
						|> rename(columns: {_start: "_time"})
						|> map(fn: (r) => ({r with to: string(v: r._stop)}))

				vols = data
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

				counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

				join.inner(
						left: vols,
						right: counts,
						on: (l, r) => l._time == r._time and l.emitter_chain == r.emitter_chain,
						as: (l, r) => ({l with count: r.count}),
				)
				|> group()
				|> sort(columns:["emitter_chain","_time"],desc:false)`
	return fmt.Sprintf(query, r.bucketInfiniteRetention, start, stop, filterSourceChain, filterTargetChain, filterAppId)
}

func (r *Repository) buildQueryChainActivityYearly(start, stop, filterSourceChain, filterTargetChain, filterAppId string) string {
	query := `
				import "date"
				import "join"

				data = from(bucket: "%s")
						|> range(start: %s,stop: %s)
						|> filter(fn: (r) => r._measurement == "chain_activity_1d")
						%s
						%s
						%s
						|> drop(columns:["destination_chain","to","app_id"])
						|> window(every: 1y, period:1y)
						|> drop(columns:["_time"])
						|> rename(columns: {_start: "_time"})
						|> map(fn: (r) => ({r with to: string(v: r._stop)}))

				vols = data
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

				counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

				join.inner(
						left: vols,
						right: counts,
						on: (l, r) => l._time == r._time and l.emitter_chain == r.emitter_chain,
						as: (l, r) => ({l with count: r.count}),
				)
				|> group()
				|> sort(columns:["emitter_chain","_time"],desc:false)`
	return fmt.Sprintf(query, r.bucketInfiniteRetention, start, stop, filterSourceChain, filterTargetChain, filterAppId)
}

func (r *Repository) buildTotalsAppActivityQuery(q ApplicationActivityQuery) string {

	var measurement string
	var bucket string
	var from, to time.Time

	switch q.Timespan {
	case "1h":
		measurement = "|> filter(fn: (r) => r._measurement == \"protocols_stats_totals_1h\")"
		bucket = r.bucket30DaysRetention
		from = q.From.Truncate(1 * time.Hour)
		to = q.To.Truncate(1 * time.Hour)
	default: // default is 1d
		measurement = "|> filter(fn: (r) => r._measurement == \"protocols_stats_totals_1d\" and r.version == \"v1\")"
		bucket = r.bucketInfiniteRetention
		from = q.From.Truncate(24 * time.Hour)
		to = q.To.Truncate(24 * time.Hour)
	}

	filterByAppId := ""
	if q.AppId != "" && !q.ExclusiveAppID {
		filterByAppId = fmt.Sprintf("|> filter(fn: (r) => r.app_id == \"TOTAL_%s\")", strings.ToUpper(q.AppId))
	}

	if q.Timespan == Month {
		return r.buildTotalsAppActivityQueryMonthly(q, measurement, bucket, filterByAppId)
	}

	query := `
			import "date"

			allData = from(bucket: "%s")
						|> range(start: %s,stop: %s)
						%s
						%s
						|> drop(columns:["emitter_chain","destination_chain"])
			
			totalMsgs = allData
						|> filter(fn: (r) => r._field == "total_messages")
						|> aggregateWindow(every: %s, fn: sum,createEmpty:true)
						|> map(fn: (r) => ({
								r with
								_value: if not exists r._value then uint(v:0) else uint(v:r._value)
     						}))
						|> group(columns:["_time","app_id","_field"])
						|> sum()
						
			tvt = allData
						|> filter(fn: (r) => r._field == "total_value_transferred")
						|> aggregateWindow(every: %s, fn: sum, createEmpty:true)
						|> map(fn: (r) => ({
								r with
								_value: if not exists r._value then uint(v:0) else r._value
     						}))
						|> group(columns:["_time","app_id","_field"])
						|> sum()

			union(tables: [totalMsgs, tvt])
				|> pivot(rowKey:["_time","app_id"], columnKey: ["_field"], valueColumn: "_value")
				|> map(fn: (r) => ({
						r with
						"total_value_transferred": float(v:r.total_value_transferred) / 100000000.0,
						"to": r._time,
						"_time": date.sub(d: %s, from: r._time)
     			}))
			`

	return fmt.Sprintf(query, bucket, from.Format(time.RFC3339), to.Format(time.RFC3339), measurement, filterByAppId, q.Timespan, q.Timespan, q.Timespan)
}

func (r *Repository) buildAppActivityQuery(q ApplicationActivityQuery) string {

	var measurement string
	var bucket string
	var from, to time.Time

	switch q.Timespan {
	case "1h":
		measurement = "protocols_stats_1h"
		bucket = r.bucket30DaysRetention
		from = q.From.Truncate(1 * time.Hour)
		to = q.To.Truncate(1 * time.Hour)
	default: // default is 1d
		measurement = "protocols_stats_1d"
		bucket = r.bucketInfiniteRetention
		from = q.From.Truncate(24 * time.Hour)
		to = q.To.Truncate(24 * time.Hour)
	}

	filterByAppId := ""
	if q.AppId != "" {
		if !q.ExclusiveAppID {
			filterByAppId = fmt.Sprintf("|> filter(fn: (r) => r.app_id_1 == \"%s\" or r.app_id_2 == \"%s\" or r.app_id_3 == \"%s\")", q.AppId, q.AppId, q.AppId)
		} else {
			filterByAppId = fmt.Sprintf("|> filter(fn: (r) => r.app_id_1 == \"%s\" and r.app_id_2 == \"none\" and r.app_id_3 == \"none\")", q.AppId)
		}
	}

	if q.Timespan == Month {
		return r.buildAppActivityQueryMonthly(q, measurement, bucket, filterByAppId)
	}

	query := `
			import "date"

				allData = from(bucket: "%s")
							|> range(start: %s,stop: %s)
							|> filter(fn: (r) => r._measurement == "%s")
							|> filter(fn: (r) => not exists r.protocol )
							%s
							|> drop(columns:["emitter_chain","destination_chain","_measurement"])

				totalMsgs = allData
							|> filter(fn: (r) => r._field == "total_messages")
							|> aggregateWindow(every: %s, fn: sum, createEmpty:true)
							|> map(fn: (r) => ({
										r with
										_value: if not exists r._value then uint(v:0) else r._value
								}))
							|> group(columns:["_time","_field","app_id_1","app_id_2","app_id_3"])
							|> sum()
						
				tvt = allData
						|> filter(fn: (r) => r._field == "total_value_transferred")
						|> aggregateWindow(every: %s, fn: sum, createEmpty:true)
						|> map(fn: (r) => ({
								r with
								_value: if not exists r._value then uint(v:0) else r._value
     						}))
						|> group(columns:["_time","_field","app_id_1","app_id_2","app_id_3"])
						|> sum()
						
				union(tables: [totalMsgs, tvt])
				|> pivot(rowKey:["_time","app_id_1","app_id_2","app_id_3"], columnKey: ["_field"], valueColumn: "_value")
				|> map(fn: (r) => ({
						r with
						"total_value_transferred": float(v:r.total_value_transferred) / 100000000.0,
						"to": r._time,
						"_time": date.sub(d: %s, from: r._time)
				}))`

	return fmt.Sprintf(query, bucket, from.Format(time.RFC3339), to.Format(time.RFC3339), measurement, filterByAppId, q.Timespan, q.Timespan, q.Timespan)
}

func (r *Repository) buildAppActivityQueryMonthly(q ApplicationActivityQuery, measurement string, bucket string, filterByAppId string) string {
	query := `
			import "date"
			import "join"

			allData = from(bucket: "%s")
						|> range(start: %s,stop: %s)
						|> filter(fn: (r) => r._measurement == "%s")
						|> filter(fn: (r) => not exists r.protocol )
						%s
						|> drop(columns:["emitter_chain","destination_chain","_measurement"])

			totalMsgs = allData
						|> filter(fn: (r) => r._field == "total_messages")
						|> aggregateWindow(every: 1mo, fn: sum)
						|> rename(columns: {_value: "total_messages"})
						|> map(fn: (r) => ({
								r with
								_time: date.sub(d: 1mo, from: r._time),
								total_messages: if not exists r.total_messages then uint(v:0) else r.total_messages
     						}))
						|> drop(columns:["_start","_stop"])
						|> group()
			
			
			tvt = allData
					|> filter(fn: (r) => r._field == "total_value_transferred")
					|> aggregateWindow(every: 1mo, fn: sum)
					|> rename(columns: {_value: "total_value_transferred"})		
					|> map(fn: (r) => ({
						r with
						_time: date.sub(d: 1mo, from: r._time),
						total_value_transferred: if not exists r.total_value_transferred then uint(v:0) else r.total_value_transferred
					}))
					|> drop(columns:["_start","_stop"])
					|> group()
						
			join.inner(
			    left: totalMsgs,
			    right: tvt,
			    on: (l, r) => l.app_id_1 == r.app_id_1 and l.app_id_2 == r.app_id_2 and l.app_id_3 == r.app_id_3 and l._time == r._time,
			    as: (l, r) => ({
					"_time":l._time,
					"to":date.add(d: 1mo, to: l._time),
					"app_id_1": l.app_id_1,
					"app_id_2": l.app_id_2,
					"app_id_3": l.app_id_3,
					"total_messages":l.total_messages,
					"total_value_transferred": float(v:r.total_value_transferred) / 100000000.0
					})
			)
		`

	from := time.Date(q.From.Year(), q.From.Month(), 1, 0, 0, 0, 0, q.From.Location())
	to := time.Date(q.To.Year(), q.To.Month(), 1, 0, 0, 0, 0, q.To.Location())
	return fmt.Sprintf(query, bucket, from.Format(time.RFC3339), to.Format(time.RFC3339), measurement, filterByAppId)
}

func (r *Repository) buildTotalsAppActivityQueryMonthly(q ApplicationActivityQuery, filterMeasurement string, bucket string, filterByAppID string) string {
	query := `
			import "date"
			import "join"

			allData = from(bucket: "%s")
						|> range(start: %s,stop: %s)
						%s
						%s
						|> drop(columns:["emitter_chain","destination_chain","version","_measurement"])
			
			totalMsgs = allData
						|> filter(fn: (r) => r._field == "total_messages")
						|> aggregateWindow(every: 1mo, fn: sum)
						|> rename(columns: {_value: "total_messages"})
						|> group()
						
			tvt = allData
						|> filter(fn: (r) => r._field == "total_value_transferred")
						|> aggregateWindow(every: 1mo, fn: sum)
						|> rename(columns: {_value: "total_value_transferred"})
						|> group()

			join.inner(
			    left: totalMsgs,
			    right: tvt,
			    on: (l, r) => l.app_id == r.app_id and l._time == r._time,
			    as: (l, r) => ({
					"to":l._time,
					"_time": date.sub(d: 1mo, from: l._time),
					"app_id": l.app_id,
					"total_messages":l.total_messages,
					"total_value_transferred": float(v:r.total_value_transferred) / 100000000.0
					}),
			)
	`
	from := time.Date(q.From.Year(), q.From.Month(), 1, 0, 0, 0, 0, q.From.Location())
	to := time.Date(q.To.Year(), q.To.Month(), 1, 0, 0, 0, 0, q.To.Location())
	return fmt.Sprintf(query, bucket, from.Format(time.RFC3339), to.Format(time.RFC3339), filterMeasurement, filterByAppID)
}
