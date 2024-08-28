package transactions

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/common"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
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

type repositoryCollections struct {
	vaas               *mongo.Collection
	vaasPythnet        *mongo.Collection
	parsedVaa          *mongo.Collection
	globalTransactions *mongo.Collection
}

type Repository struct {
	tvl                     *tvl.Tvl
	p2pNetwork              string
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
	db *mongo.Database,
	logger *zap.Logger,
) *Repository {

	r := Repository{
		tvl:                     tvl,
		p2pNetwork:              p2pNetwork,
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

func (r *Repository) GetScorecards(ctx context.Context) (*Scorecards, error) {

	// This function launches one goroutine for each scorecard.
	//
	// We use a `sync.WaitGroup` to block until all goroutines are done.
	var wg sync.WaitGroup

	var messages24h, tvl, totalTxCount, totalTxVolume, volume24h, volume7d, volume30d, totalPythMessage string

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
		volume24h, err = r.getVolume(ctx, _24h)
		if err != nil {
			r.logger.Error("failed to get 24h volume", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		volume7d, err = r.getVolume(ctx, _7d)
		if err != nil {
			r.logger.Error("failed to get 7d volume", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		volume30d, err = r.getVolume(ctx, _30d)
		if err != nil {
			r.logger.Error("failed to get 30d volume", zap.Error(err))
		}
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
		Tvl:           tvl,
		Volume24h:     volume24h,
		Volume7d:      volume7d,
		Volume30d:     volume30d,
	}
	return &scorecards, nil
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

func (r *Repository) getVolume(ctx context.Context, from offset) (string, error) {

	// query volume
	query := fmt.Sprintf(queryTemplateVolume, r.bucketInfiniteRetention, from)
	result, err := r.queryAPI.Query(ctx, query)
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
	if r.p2pNetwork != config.P2pMainNet {
		return "0", nil

	}
	pythEmitterAddr := "e101faedac5851e32b9b23b5f9411a8c2bac4aae3ed4dd7b811dd1a72ea4aa71"
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
			|> range(start: 1970-01-01T00:00:00Z)
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

func (r *Repository) FindTokenSymbolActivity(ctx context.Context, payload TokenSymbolActivityQuery) ([]tokenSymbolActivityResult, error) {
	query := r.buildTokenSymbolActivityQuery(payload)

	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var response []tokenSymbolActivityResult
	for result.Next() {
		var row tokenSymbolActivityResult
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

	switch q.Timespan {
	case "1h":
		measurement = "|> filter(fn: (r) => r._measurement == \"protocols_stats_totals_1h\")"
		bucket = r.bucket30DaysRetention
	default: // default is 1d
		measurement = "|> filter(fn: (r) => r._measurement == \"protocols_stats_totals_1d\" and r.version == \"v1\")"
		bucket = r.bucketInfiniteRetention
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
			import "join"

			allData = from(bucket: "%s")
						|> range(start: %s,stop: %s)
						%s
						%s
						|> drop(columns:["emitter_chain","destination_chain"])
			
			totalMsgs = allData
						|> filter(fn: (r) => r._field == "total_messages")
						|> group(columns:["_time","app_id"])
						|> sum()
						|> rename(columns: {_value: "total_messages"})
						
			tvt = allData
						|> filter(fn: (r) => r._field == "total_value_transferred")
						|> group(columns:["_time","app_id"])
						|> sum()
						|> rename(columns: {_value: "total_value_transferred"})

			join.inner(
			    left: totalMsgs,
			    right: tvt,
			    on: (l, r) => l.app_id == r.app_id and l._time == r._time,
			    as: (l, r) => ({
					"_time":l._time,
					"to":date.add(d: %s, to: l._time),
					"app_id": l.app_id,
					"total_messages":l.total_messages,
					"total_value_transferred": float(v:r.total_value_transferred) / 100000000.0
					}),
			)
			`

	return fmt.Sprintf(query, bucket, q.From.Format(time.RFC3339), q.To.Format(time.RFC3339), measurement, filterByAppId, q.Timespan)
}

func (r *Repository) buildAppActivityQuery(q ApplicationActivityQuery) string {

	var measurement string
	var bucket string

	switch q.Timespan {
	case "1h":
		measurement = "protocols_stats_1h"
		bucket = r.bucket30DaysRetention
	default: // default is 1d
		measurement = "protocols_stats_1d"
		bucket = r.bucketInfiniteRetention
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
			import "join"

			allData =	from(bucket: "%s")
						|> range(start: %s,stop: %s)
						|> filter(fn: (r) => r._measurement == "%s")
						|> filter(fn: (r) => not exists r.protocol )
						%s
						|> drop(columns:["emitter_chain","destination_chain"])

			totalMsgs = allData
						|> filter(fn: (r) => r._field == "total_messages")
						|> group(columns:["_time","app_id_1","app_id_2","app_id_3"])
						|> sum()
						|> rename(columns: {_value: "total_messages"})
						
			tvt = allData
						|> filter(fn: (r) => r._field == "total_value_transferred")
						|> group(columns:["_time","app_id_1","app_id_2","app_id_3"])
						|> sum()
						|> rename(columns: {_value: "total_value_transferred"})
			
			join.inner(
			    left: totalMsgs,
			    right: tvt,
			    on: (l, r) => l.app_id_1 == r.app_id_1 and l.app_id_2 == r.app_id_2 and l.app_id_3 == r.app_id_3 and l._time == r._time,
			    as: (l, r) => ({
					"_time":l._time,
					"to":date.add(d: %s, to: l._time),
					"app_id_1": l.app_id_1,
					"app_id_2": l.app_id_2,
					"app_id_3": l.app_id_3,
					"total_messages":l.total_messages,
					"total_value_transferred": float(v:r.total_value_transferred) / 100000000.0
					})
			)`

	return fmt.Sprintf(query, bucket, q.From.Format(time.RFC3339), q.To.Format(time.RFC3339), measurement, filterByAppId, q.Timespan)
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
	return fmt.Sprintf(query, bucket, q.From.Format(time.RFC3339), q.To.Format(time.RFC3339), measurement, filterByAppId)

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
	return fmt.Sprintf(query, bucket, q.From.Format(time.RFC3339), q.To.Format(time.RFC3339), filterMeasurement, filterByAppID)
}
