package stats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/coingecko"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/stats"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Repository struct {
	nttRepo                 *stats.NTTRepository
	influxCli               influxdb2.Client
	queryAPI                api.QueryAPI
	bucket24HoursRetention  string
	bucketInfiniteRetention string
	coingeckoAPI            *coingecko.CoinGeckoAPI
	tokenProvider           *domain.TokenProvider
	logger                  *zap.Logger
}

func NewRepository(
	nttRepo *stats.NTTRepository,
	client influxdb2.Client,
	org string,
	bucket24HoursRetention string,
	bucketInfiniteRetention string,
	coingeckoAPI *coingecko.CoinGeckoAPI,
	tokenProvider *domain.TokenProvider,
	logger *zap.Logger,
) *Repository {

	r := Repository{
		nttRepo:                 nttRepo,
		influxCli:               client,
		queryAPI:                client.QueryAPI(org),
		bucket24HoursRetention:  bucket24HoursRetention,
		bucketInfiniteRetention: bucketInfiniteRetention,
		coingeckoAPI:            coingeckoAPI,
		tokenProvider:           tokenProvider,
		logger:                  logger,
	}
	return &r
}

func (r *Repository) GetSymbolWithAssets(ctx context.Context, timeSpan SymbolWithAssetsTimeSpan) ([]SymbolWithAssetDTO, error) {
	var measurement string
	switch timeSpan {
	case TimeSpan7Days:
		measurement = "assets_by_symbol_7_days_3h_v2"
	case TimeSpan15Days:
		measurement = "assets_by_symbol_15_days_3h_v2"
	case TimeSpan30Days:
		measurement = "assets_by_symbol_30_days_3h_v2"
	default:
		measurement = "assets_by_symbol_7_days_3h_v2"
	}

	query := buildSymbolWithAssets(r.bucket24HoursRetention, time.Now(), measurement)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	// Scan query results
	type Row struct {
		Symbol       string `mapstructure:"symbol"`
		EmitterChain string `mapstructure:"emitter_chain"`
		TokenChain   string `mapstructure:"token_chain"`
		TokenAddress string `mapstructure:"token_address"`
		JsonValue    string `mapstructure:"_value"`
	}

	type TxsVolume struct {
		Txs    decimal.Decimal
		Volume decimal.Decimal
	}

	var rows []Row
	for result.Next() {
		var row Row
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	divisor := decimal.NewFromInt(1_0000_0000)

	// Convert the rows into the response model
	var values []SymbolWithAssetDTO
	for _, row := range rows {

		// parse emitter chain
		emitterChain, err := strconv.ParseUint(row.EmitterChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert emitter chain field to uint16. %v", err)
		}

		// parse token chain
		tokenChain, err := strconv.ParseUint(row.TokenChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert token chain field to uint16. %v", err)
		}

		// parse the json value
		var txsVolume TxsVolume
		if err := json.Unmarshal([]byte(row.JsonValue), &txsVolume); err != nil {
			return nil, fmt.Errorf("failed to convert _value to struct. %v", err)
		}

		// append the new item to the response
		value := SymbolWithAssetDTO{
			Symbol:         row.Symbol,
			EmitterChainID: sdk.ChainID(emitterChain),
			TokenChainID:   sdk.ChainID(tokenChain),
			TokenAddress:   row.TokenAddress,
			Volume:         txsVolume.Volume.Div(divisor),
			Txs:            txsVolume.Txs,
		}

		// do not include invalid chain IDs in the response
		if !domain.ChainIdIsValid(value.EmitterChainID) {
			continue
		}

		values = append(values, value)
	}

	return values, nil
}

func (r *Repository) GetTopCorridores(ctx context.Context, timeSpan TopCorridorsTimeSpan) ([]TopCorridorsDTO, error) {
	var measurement string
	switch timeSpan {
	case TimeSpan7DaysTopCorridors:
		measurement = "top_100_corridors_7_days_3h_v2"
	case TimeSpan2DaysTopCorridors:
		measurement = "top_100_corridors_2_days_3h_v2"
	default:
		measurement = "top_100_corridors_2_days_3h_v2"
	}

	query := buildTopCorridors(r.bucket24HoursRetention, time.Now(), measurement)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	// Scan query results
	type Row struct {
		EmitterChain     string `mapstructure:"emitter_chain"`
		DestinationChain string `mapstructure:"destination_chain"`
		TokenChain       string `mapstructure:"token_chain"`
		TokenAddress     string `mapstructure:"token_address"`
		Txs              uint64 `mapstructure:"_value"`
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
	var values []TopCorridorsDTO
	for _, row := range rows {

		// parse emitter chain
		emitterChain, err := strconv.ParseUint(row.EmitterChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert emitter chain field to uint16. %v", err)
		}

		// parse emitter chain
		destinationChain, err := strconv.ParseUint(row.DestinationChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert destination chain field to uint16. %v", err)
		}

		// parse token chain
		tokenChain, err := strconv.ParseUint(row.TokenChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert token chain field to uint16. %v", err)
		}

		// append the new item to the response
		value := TopCorridorsDTO{
			EmitterChainID:     sdk.ChainID(emitterChain),
			DestinationChainID: sdk.ChainID(destinationChain),
			TokenChainID:       sdk.ChainID(tokenChain),
			TokenAddress:       row.TokenAddress,
			Txs:                row.Txs,
		}

		// do not include invalid chain IDs in the response
		if !domain.ChainIdIsValid(value.EmitterChainID) || !domain.ChainIdIsValid(value.DestinationChainID) {
			r.logger.Warn("Invalid chain ID in top corridors",
				zap.Uint16("emitter_chain", uint16(value.EmitterChainID)),
				zap.Uint16("destination_chain", uint16(value.DestinationChainID)),
				zap.Uint16("token_chain", uint16(value.TokenChainID)),
				zap.String("token_address", value.TokenAddress),
				zap.Uint64("txs", value.Txs),
			)
			continue
		}

		values = append(values, value)
	}

	return values, nil
}

func (r *Repository) GetNativeTokenTransferSummary(ctx context.Context, symbol string) (*NativeTokenTransferSummary, error) {
	var wg sync.WaitGroup

	var marketcap, circulatingSupply *decimal.Decimal
	var totalValueTokenTransferred, totalTokenTransferred *decimal.Decimal
	var medianTransferSize, averageTransferSize *decimal.Decimal

	// get symbol market cap and circulating supply
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		coingeckoID := r.tokenProvider.GetCoingeckoIDBySymbol(symbol)
		marketDataResponse, err := r.coingeckoAPI.GetMarketData(coingeckoID)
		if err != nil {
			r.logger.Error("failed to get market cap", zap.Error(err))
			return
		}
		marketcap = marketDataResponse.MarketData.MarketCap.Usd
		circulatingSupply = marketDataResponse.MarketData.CirculatingSupply
	}()

	// get total value token transferred
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		totalValueTokenTransferred, err = r.getNTTTotalValueTokenTransferred(ctx, symbol)
		if err != nil {
			r.logger.Error("failed to get total value token transferred", zap.Error(err))
		}
	}()

	// get total token transferred
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		totalTokenTransferred, err = r.getNTTTotalTokenTransferred(ctx, symbol)
		if err != nil {
			r.logger.Error("failed to get total token transferred", zap.Error(err))
		}
	}()

	// get median transfer size
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		nttMedian, err := r.nttRepo.GetNativeTokenTransferMedian(ctx, symbol)
		if err != nil {
			r.logger.Error("failed to get median transfer size", zap.Error(err))
		} else {
			medianTransferSize = &nttMedian
		}
	}()

	// get average transfer size
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		averageTransferSize, err = r.getNTTAverageTransferSize(ctx, symbol)
		if err != nil {
			r.logger.Error("failed to get average transfer size", zap.Error(err))
		}
	}()

	wg.Wait()

	summary := NativeTokenTransferSummary{
		MarketCap:                  marketcap,
		CirculatingSupply:          circulatingSupply,
		TotalValueTokenTransferred: totalValueTokenTransferred,
		TotalTokenTransferred:      totalTokenTransferred,
		MedianTransferSize:         medianTransferSize,
		AverageTransferSize:        averageTransferSize,
	}

	return &summary, nil

}

func (r *Repository) getNTTTotalValueTokenTransferred(ctx context.Context, symbol string) (*decimal.Decimal, error) {
	query := buildNTTTotalValueTokenTransferred(r.bucketInfiniteRetention, time.Now(), symbol)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query ntt total value tokend transferred",
			zap.String("symbol", symbol), zap.Error(err))
		return nil, err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query ntt total value tokend transferred has errors",
			zap.String("symbol", symbol), zap.Error(err))
		return nil, result.Err()
	}
	if !result.Next() {
		r.logger.Error("ntt total value tokend transferred query result has no next",
			zap.String("symbol", symbol))
		return nil, errors.New("no result")
	}
	row := struct {
		Value uint64 `mapstructure:"_value"`
	}{}
	if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return nil, fmt.Errorf("failed to decode total value transferred for symbol(%s): %w", symbol, err)
	}

	// convert the value to decimal
	value := decimal.NewFromInt(int64(row.Value))
	return &value, nil
}

func (r *Repository) getNTTTotalTokenTransferred(ctx context.Context, symbol string) (*decimal.Decimal, error) {
	query := buildNTTTotalTokenTransferred(r.bucketInfiniteRetention, time.Now(), symbol)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query ntt total token transferred",
			zap.String("symbol", symbol), zap.Error(err))
		return nil, err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query ntt total token transferred has errors",
			zap.String("symbol", symbol), zap.Error(err))
		return nil, result.Err()
	}
	if !result.Next() {
		r.logger.Error("ntt total token transferred query result has no next",
			zap.String("symbol", symbol))
		return nil, errors.New("no result")
	}
	row := struct {
		Value uint64 `mapstructure:"_value"`
	}{}

	if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return nil, fmt.Errorf("failed to decode total token transferred for symbol(%s): %w", symbol, err)
	}

	// convert the value to decimal
	value := decimal.NewFromInt(int64(row.Value))
	return &value, nil
}

func (r *Repository) getNTTAverageTransferSize(ctx context.Context, symbol string) (*decimal.Decimal, error) {
	query := buildNTTAverageTransferSize(r.bucketInfiniteRetention, symbol)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query ntt average transfer size",
			zap.String("symbol", symbol), zap.Error(err))
		return nil, err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query ntt average transfer size has errors",
			zap.String("symbol", symbol), zap.Error(err))
		return nil, result.Err()
	}
	if !result.Next() {
		r.logger.Error("ntt average transfer size query result has no next",
			zap.String("symbol", symbol))
		return nil, errors.New("no result")
	}
	row := struct {
		Value float64 `mapstructure:"_value"`
	}{}
	if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return nil, fmt.Errorf("failed to decode average transfer size for symbol(%s): %w", symbol, err)
	}

	// convert the value to decimal
	value := decimal.NewFromFloat(row.Value)
	return &value, nil
}

func (r *Repository) GetNativeTokenTransferActivity(ctx context.Context, isNotional bool, symbol string) ([]NativeTokenTransferActivity, error) {
	query := buildNTTChainActivity(r.bucketInfiniteRetention, time.Now(), symbol, isNotional)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query native token transfer activity", zap.Error(err))
		return nil, err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query native token transfer activity has errors", zap.Error(err))
		return nil, result.Err()
	}

	type Row struct {
		Value              uint64 `mapstructure:"_value"`
		DestinationChainID string `mapstructure:"destination_chain"`
		EmitterChainID     string `mapstructure:"emitter_chain"`
		Symbol             string `mapstructure:"symbol"`
	}

	var rows []Row
	for result.Next() {
		var row Row
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	var values []NativeTokenTransferActivity

	for _, row := range rows {

		// parse emitter chain
		emitterChain, err := strconv.ParseUint(row.EmitterChainID, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert emitter chain field to uint16. %v", err)
		}

		// parse emitter chain
		destinationChain, err := strconv.ParseUint(row.DestinationChainID, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert destination chain field to uint16. %v", err)
		}

		rowValue := decimal.NewFromUint64(row.Value)

		if isNotional {
			rowValue = rowValue.Div(decimal.NewFromInt(1_0000_0000))
		}

		// append the new item to the response
		value := NativeTokenTransferActivity{
			EmitterChainID:     sdk.ChainID(emitterChain),
			DestinationChainID: sdk.ChainID(destinationChain),
			Symbol:             row.Symbol,
			Value:              rowValue,
		}

		// do not include invalid chain IDs in the response
		if !domain.ChainIdIsValid(value.EmitterChainID) || !domain.ChainIdIsValid(value.DestinationChainID) {
			r.logger.Warn("Invalid chain ID in native token transfer activity",
				zap.Uint16("emitter_chain", uint16(value.EmitterChainID)),
				zap.Uint16("destination_chain", uint16(value.DestinationChainID)),
				zap.String("value", value.Value.String()),
			)
			continue
		}

		values = append(values, value)
	}

	return values, nil
}

func (r *Repository) GetNativeTokenTransferByTime(ctx context.Context, timespan NttTimespan, symbol string, isNotional bool, from, to time.Time) ([]NativeTokenTransferByTime, error) {
	var query string
	switch timespan {
	case HourNttTimespan:
		start := from.Truncate(1 * time.Hour).UTC().Format(time.RFC3339)
		stop := to.Truncate(1 * time.Hour).UTC().Format(time.RFC3339)
		query = r.buildQueryGetNativeTokenTransferByTimeHourly(start, stop, symbol, isNotional)
	case DayNttTimespan:
		start := from.Truncate(24 * time.Hour).UTC().Format(time.RFC3339)
		stop := to.Truncate(24 * time.Hour).UTC().Format(time.RFC3339)
		query = r.buildQueryGetNativeTokenTransferByTimeDaily(start, stop, symbol, isNotional)
	case MonthNttTimespan:
		start := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location()).UTC().Format(time.RFC3339)
		stop := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, to.Location()).UTC().Format(time.RFC3339)
		query = r.buildQueryGetNativeTokenTransferByTimeMonthly(start, stop, symbol, isNotional)
	default:
		start := time.Date(from.Year(), 1, 1, 0, 0, 0, 0, from.Location()).UTC().Format(time.RFC3339)
		stop := time.Date(to.Year(), 1, 1, 0, 0, 0, 0, to.Location()).UTC().Format(time.RFC3339)
		query = r.buildQueryGetNativeTokenTransferByTimeYearly(start, stop, symbol, isNotional)
	}

	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query native token transfer activity", zap.Error(err))
		return nil, err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query native token transfer activity has errors", zap.Error(err))
		return nil, result.Err()
	}

	type row struct {
		Value  uint64    `mapstructure:"_value"`
		Symbol string    `mapstructure:"symbol"`
		Time   time.Time `mapstructure:"_time"`
	}

	var rows []row
	for result.Next() {
		var row row
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	var values []NativeTokenTransferByTime

	for _, row := range rows {

		rowValue := decimal.NewFromUint64(row.Value)

		if isNotional {
			rowValue = rowValue.Div(decimal.NewFromInt(1_0000_0000))
		}

		time := buildTimeForNativeTokenTransferByTime(row.Time, timespan)

		// append the new item to the response
		value := NativeTokenTransferByTime{
			Symbol: row.Symbol,
			Value:  rowValue,
			Time:   time,
		}

		values = append(values, value)
	}

	return values, nil
}

func (r *Repository) buildQueryGetNativeTokenTransferByTimeHourly(start, stop, symbol string, isNotional bool) string {
	function := "count"
	if isNotional {
		function = "sum"
	}
	query := `
	import "influxdata/influxdb/schema"
	import "strings"

	start = %s
	stop =  %s
	bucket = "%s"
	symbol = "%s"

	from(bucket: bucket)
	|> range(start: start, stop: stop)
	|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
	|> filter(fn: (r) => r.app_id_1 == "NATIVE_TOKEN_TRANSFER" or r.app_id_2 == "NATIVE_TOKEN_TRANSFER" or r.app_id_3 == "NATIVE_TOKEN_TRANSFER")
	|> filter(fn: (r) => (r._field == "symbol" and r._value != "") or r._field == "volume")
	|> schema.fieldsAsCols()
	|> filter(fn: (r) => r.symbol == symbol)
	|> group()
	|> map(fn: (r) => ({r with _value: r.volume}))
	|> aggregateWindow(every: 1h, fn: %s, createEmpty: true)`
	return fmt.Sprintf(query, start, stop, r.bucketInfiniteRetention, strings.ToUpper(symbol), function)
}

func (r *Repository) buildQueryGetNativeTokenTransferByTimeDaily(start, stop, symbol string, isNotional bool) string {
	return buildNTTChainActivityByTime(r.bucketInfiniteRetention, start, stop, strings.ToUpper(symbol), isNotional, "1d")
}

func (r *Repository) buildQueryGetNativeTokenTransferByTimeMonthly(start, stop, symbol string, isNotional bool) string {
	return buildNTTChainActivityByTime(r.bucketInfiniteRetention, start, stop, strings.ToUpper(symbol), isNotional, "1mo")
}

func (r *Repository) buildQueryGetNativeTokenTransferByTimeYearly(start, stop, symbol string, isNotional bool) string {
	return buildNTTChainActivityByTime(r.bucketInfiniteRetention, start, stop, strings.ToUpper(symbol), isNotional, "1y")
}

func buildTimeForNativeTokenTransferByTime(timestamp time.Time, timespan NttTimespan) time.Time {
	switch timespan {
	case HourNttTimespan:
		return timestamp.Add(-1 * time.Hour)
	case DayNttTimespan:
		return timestamp.AddDate(0, 0, -1)
	case MonthNttTimespan:
		return timestamp.AddDate(0, -1, 0)
	case YearNttTimespan:
		return timestamp.AddDate(-1, 0, 0)
	default:
		return timestamp
	}
}
