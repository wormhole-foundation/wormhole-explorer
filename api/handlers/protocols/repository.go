package protocols

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/mitchellh/mapstructure"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"go.uber.org/zap"
	"time"
)

// QueryCoreProtocolTotalStartOfDay Query template for core protocols (cctp and portal_token_bridge) to fetch total values till the start of current day
const QueryCoreProtocolTotalStartOfDay = `
		import "date"
	
		startOfCurrentDay = date.truncate(t: now(), unit: 1d)
		
	data =	from(bucket: "%s")
		|> range(start: 1970-01-01T00:00:00Z,stop:startOfCurrentDay)
		|> filter(fn: (r) => r._measurement == "%s" and r.version == "v1" and r.app_id == "TOTAL_%s")
		|> drop(columns: ["emitter_chain","destination_chain","version"])
		
tvt = data	
		|> filter(fn : (r) => r._field == "total_value_transferred")
		|> group()
		|> sum()
		|> set(key:"_field",value:"total_value_transferred")
		|> map(fn: (r) => ({r with _value: int(v: r._value)}))
		
totalMsgs =	data	
		|> filter(fn : (r) => r._field == "total_messages")
		|> group()
		|> sum()	
		|> set(key:"_field",value:"total_messages")
		|> map(fn: (r) => ({r with _value: int(v: r._value)}))
		
union(tables:[tvt,totalMsgs])		
		|> set(key:"_time",value:string(v:startOfCurrentDay))
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> set(key:"app_id",value:"%s")
`

// QueryCoreProtocolDeltaSinceStartOfDay calculate delta since the beginning of current day
const QueryCoreProtocolDeltaSinceStartOfDay = `
		import "date"
		import "types"

		ts = date.truncate(t: now(), unit: 1h)
		startOfDay = date.truncate(t: now(), unit: 1d)
		
	data =	from(bucket: "%s")
		|> range(start: startOfDay,stop:ts)
		|> filter(fn: (r) => r._measurement == "%s" and r.app_id == "TOTAL_%s")
		|> drop(columns: ["emitter_chain","destination_chain","version"])
		
tvt =	data	
		|> filter(fn : (r) => r._field == "total_value_transferred")
		|> group()
		|> sum()
		|> set(key:"_field",value:"total_value_transferred")
		|> map(fn: (r) => ({r with _value: int(v: r._value)}))
		
totalMsgs =	data	
		|> filter(fn : (r) => r._field == "total_messages")
		|> group()
		|> sum()	
		|> set(key:"_field",value:"total_messages")
		|> map(fn: (r) => ({r with _value: int(v: r._value)}))
		
union(tables:[tvt,totalMsgs])		
		|> set(key:"_time",value:string(v:startOfDay))
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> set(key:"app_id",value:"%s")
`

// QueryCoreProtocolDeltaLastDay calculate last day delta
const QueryCoreProtocolDeltaLastDay = `
		import "date"
		import "types"

		ts = date.truncate(t: now(), unit: 1h)
		yesterday = date.sub(d: 1d, from: ts)
		
	data =	from(bucket: "%s")
		|> range(start: yesterday,stop:ts)
		|> filter(fn: (r) => r._measurement == "%s" and r.app_id == "TOTAL_%s")
		|> drop(columns: ["emitter_chain","destination_chain"])
		
tvt =	data	
		|> filter(fn : (r) => r._field == "total_value_transferred")
		|> group()
		|> sum()
		|> set(key:"_field",value:"total_value_transferred")
		|> map(fn: (r) => ({r with _value: int(v: r._value)}))
		
totalMsgs =	data	
		|> filter(fn : (r) => r._field == "total_messages")
		|> group()
		|> sum()	
		|> set(key:"_field",value:"total_messages")
		|> map(fn: (r) => ({r with _value: int(v: r._value)}))
		
union(tables:[tvt,totalMsgs])		
		|> set(key:"_time",value:string(v:yesterday))
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> set(key:"app_id",value:"%s")
`

const QueryTemplateProtocolStatsLastDay = `
		from(bucket: "%s")
  			|> range(start: %s, stop: %s)
  			|> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s")
			|> first()
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
`

const QueryTemplateProtocolStats = `
		data = from(bucket: "%s")
  					|> range(start: -2d)
  					|> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s")

		totalMsg = data
					|> filter(fn: (r) => r._field == "total_messages")
					|> sort(columns:["_time"],desc:false)
					|> last()	

		tvl = data	
  					|> filter(fn: (r) => r._field == "total_value_locked")
					|> sort(columns:["_time"],desc:false)
					|> last()

		volume = data	
  					|> filter(fn: (r) => r._field == "volume")
					|> sort(columns:["_time"],desc:false)
					|> last()
	
		union(tables:[totalMsg,tvl,volume])
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
`

const QueryTemplateProtocolActivity = `	
		data = 
		from(bucket: "%s")
		  |> range(start: %s)
		  |> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s")
			
		tvs = data	
			|> filter(fn: (r) => r._field == "total_value_secure") 
			|> cumulativeSum()
			|> last()
		
		tvt = data	
			|> filter(fn: (r) => r._field == "total_value_transferred") 
			|> cumulativeSum()
			|> last()
			
		volume = data	
		  	|> filter(fn: (r) => r._field == "volume")
			|> sort(columns:["_time"],desc:false)
		  	|> cumulativeSum()
			|> last()	

		txs = data	
  			|> filter(fn: (r) => r._field == "txs")
			|> sort(columns:["_time"],desc:false)
  			|> cumulativeSum()
			|> last()

		union(tables:[tvs,tvt,volume,txs])
		 |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
`

type Repository struct {
	queryAPI                QueryDoer
	logger                  *zap.Logger
	bucket24Hrs             string
	bucketInfinite          string
	bucket30d               string
	coreProtocolMeasurement struct {
		Daily  string
		Hourly string
	}
}

type rowStat struct {
	Protocol         string    `mapstructure:"protocol"`
	TotalMessages    uint64    `mapstructure:"total_messages"`
	TotalValueLocked float64   `mapstructure:"total_value_locked"`
	Volume           float64   `mapstructure:"volume"`
	Time             time.Time `mapstructure:"_time"`
}

type intRowStat struct {
	Protocol              string  `mapstructure:"app_id"`
	TotalMessages         uint64  `mapstructure:"total_messages"`
	TotalValueTransferred float64 `mapstructure:"total_value_transferred"`
}

type intStats struct {
	Latest        intRowStat
	DeltaLast24hr intRowStat
}

type rowActivity struct {
	Protocol              string    `mapstructure:"protocol"`
	Time                  time.Time `mapstructure:"_time"`
	TotalUsd              float64   `mapstructure:"total_usd"`
	TotalValueTransferred float64   `mapstructure:"total_value_transferred"`
	TotalValueSecure      float64   `mapstructure:"total_value_secure"`
	Txs                   uint64    `mapstructure:"txs"`
}

type stats struct {
	Latest rowStat
	Last24 rowStat
}

type QueryDoer interface {
	Query(ctx context.Context, query string) (QueryResult, error)
}

type queryApiWrapper struct {
	qApi api.QueryAPI
}

type QueryResult interface {
	Next() bool
	Record() *query.FluxRecord
	Err() error
	Close() error
}

func WrapQueryAPI(qApi api.QueryAPI) QueryDoer {
	return &queryApiWrapper{qApi: qApi}
}

func NewRepository(qApi QueryDoer, bucketInfinite, bucket30d, bucket24Hrs string, logger *zap.Logger) *Repository {
	return &Repository{
		queryAPI:       qApi,
		bucketInfinite: bucketInfinite,
		bucket30d:      bucket30d,
		bucket24Hrs:    bucket24Hrs,
		logger:         logger,
		coreProtocolMeasurement: struct {
			Daily  string
			Hourly string
		}{
			Daily:  dbconsts.TotalProtocolsStatsDaily,
			Hourly: dbconsts.TotalProtocolsStatsHourly,
		},
	}
}

func (q *queryApiWrapper) Query(ctx context.Context, query string) (QueryResult, error) {
	return q.qApi.Query(ctx, query)
}

func (r *Repository) getProtocolStats(ctx context.Context, protocol string) (rowStat, error) {

	q := fmt.Sprintf(QueryTemplateProtocolStats, r.bucket30d, dbconsts.ProtocolsStatsMeasurementHourly, protocol)

	statsData, err := fetchSingleRecordData[rowStat](r.logger, r.queryAPI, ctx, q, protocol)
	if err != nil {
		r.logger.Error("error fetching latest daily stats", zap.Error(err))
		return rowStat{}, err
	}

	return rowStat{
		Protocol:         protocol,
		TotalMessages:    statsData.TotalMessages,
		TotalValueLocked: statsData.TotalValueLocked,
		Volume:           statsData.Volume,
	}, nil

}

func (r *Repository) getProtocolStatsLastDay(ctx context.Context, protocol string) (rowStat, error) {

	to := time.Now().UTC().Truncate(24 * time.Hour)
	from := to.Add(-24 * time.Hour)
	q := fmt.Sprintf(QueryTemplateProtocolStatsLastDay, r.bucket30d, from.Format(time.RFC3339), to.Format(time.RFC3339), dbconsts.ProtocolsStatsMeasurementHourly, protocol)

	lastDayData, err := fetchSingleRecordData[rowStat](r.logger, r.queryAPI, ctx, q, protocol)
	if err != nil {
		r.logger.Error("error fetching last day stats", zap.Error(err))
		return rowStat{}, err
	}

	return lastDayData, nil

}

func (r *Repository) getProtocolActivity(ctx context.Context, protocol string) (rowActivity, error) {

	q := fmt.Sprintf(QueryTemplateProtocolActivity, r.bucketInfinite, "1970-01-01T00:00:00Z", dbconsts.ProtocolsActivityMeasurementDaily, protocol)
	activityDaily, err := fetchSingleRecordData[rowActivity](r.logger, r.queryAPI, ctx, q, protocol)
	if err != nil {
		r.logger.Error("error fetching latest daily activity", zap.Error(err))
		return rowActivity{}, err
	}

	q = fmt.Sprintf(QueryTemplateProtocolActivity, r.bucket30d, activityDaily.Time.Format(time.RFC3339), dbconsts.ProtocolsActivityMeasurementHourly, protocol)
	activityHourly, err := fetchSingleRecordData[rowActivity](r.logger, r.queryAPI, ctx, q, protocol)

	return rowActivity{
		Protocol:              protocol,
		Txs:                   activityDaily.Txs + activityHourly.Txs,
		TotalUsd:              activityDaily.TotalUsd + activityHourly.TotalUsd,
		TotalValueTransferred: activityDaily.TotalValueTransferred + activityHourly.TotalValueTransferred,
		TotalValueSecure:      activityDaily.TotalValueSecure + activityHourly.TotalValueSecure,
	}, nil
}

// returns latest and last 24 hr for core protocols (portal_token_bridge and ntt)
func (r *Repository) getCoreProtocolStats(ctx context.Context, protocol string) (intStats, error) {

	// calculate total values till the start of current day
	totalTillCurrentDayQuery := fmt.Sprintf(QueryCoreProtocolTotalStartOfDay, r.bucketInfinite, r.coreProtocolMeasurement.Daily, protocol, protocol)
	totalsUntilToday, err := fetchSingleRecordData[intRowStat](r.logger, r.queryAPI, ctx, totalTillCurrentDayQuery, protocol)
	if err != nil {
		return intStats{}, err
	}

	// calculate delta since the beginning of current day
	q2 := fmt.Sprintf(QueryCoreProtocolDeltaSinceStartOfDay, r.bucket30d, r.coreProtocolMeasurement.Hourly, protocol, protocol)
	currentDayStats, errCD := fetchSingleRecordData[intRowStat](r.logger, r.queryAPI, ctx, q2, protocol)
	if errCD != nil {
		return intStats{}, errCD
	}

	latestTotal := intRowStat{
		Protocol:              protocol,
		TotalMessages:         totalsUntilToday.TotalMessages + currentDayStats.TotalMessages,
		TotalValueTransferred: (totalsUntilToday.TotalValueTransferred + currentDayStats.TotalValueTransferred) / getProtocolDecimals(protocol),
	}

	result := intStats{
		Latest: latestTotal,
	}

	// calculate last day delta
	q3 := fmt.Sprintf(QueryCoreProtocolDeltaLastDay, r.bucket30d, r.coreProtocolMeasurement.Hourly, protocol, protocol)
	deltaYesterdayStats, errQ3 := fetchSingleRecordData[intRowStat](r.logger, r.queryAPI, ctx, q3, protocol)
	if errQ3 != nil {
		return result, errQ3
	}
	deltaYesterdayStats.TotalValueTransferred = deltaYesterdayStats.TotalValueTransferred / getProtocolDecimals(protocol)

	result.DeltaLast24hr = deltaYesterdayStats
	return result, nil
}

func (r *Repository) getCCTPStats(ctx context.Context, protocol string) (intStats, error) {

	queryTemplate := `
		from(bucket: "%s")
			|> range(start: -1d)
			|> filter(fn: (r) => r._measurement == "cctp_status_total_v2")
			%s
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> rename(columns: {txs: "total_messages", volume: "total_value_transferred"})
	`
	q := fmt.Sprintf(queryTemplate, r.bucket24Hrs, "|> last()")
	statsData, err := fetchSingleRecordData[intRowStat](r.logger, r.queryAPI, ctx, q, protocol)
	if err != nil {
		r.logger.Error("error fetching cctp totals stats", zap.Error(err))
		return intStats{}, err
	}

	q = fmt.Sprintf(queryTemplate, r.bucket24Hrs, "|> first()")
	totals24HrAgo, err := fetchSingleRecordData[intRowStat](r.logger, r.queryAPI, ctx, q, protocol)
	if err != nil {
		r.logger.Error("error fetching cctp totals stats", zap.Error(err))
		return intStats{}, err
	}

	return intStats{
		Latest: intRowStat{
			Protocol:              protocol,
			TotalMessages:         statsData.TotalMessages,
			TotalValueTransferred: statsData.TotalValueTransferred / getProtocolDecimals(protocol),
		},
		DeltaLast24hr: intRowStat{
			Protocol:              protocol,
			TotalMessages:         statsData.TotalMessages - totals24HrAgo.TotalMessages,
			TotalValueTransferred: (statsData.TotalValueTransferred - totals24HrAgo.TotalValueTransferred) / getProtocolDecimals(protocol),
		},
	}, nil

}

func fetchSingleRecordData[T any](logger *zap.Logger, queryAPI QueryDoer, ctx context.Context, query, protocol string) (T, error) {
	var res T
	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		logger.Error("error executing query to fetch data", zap.Error(err), zap.String("protocol", protocol), zap.String("query", query))
		return res, err
	}
	defer result.Close()

	if !result.Next() {
		if result.Err() != nil {
			logger.Error("error reading query response", zap.Error(result.Err()), zap.String("protocol", protocol), zap.String("query", query))
			return res, result.Err()
		}
		logger.Info("empty query response", zap.String("protocol", protocol), zap.String("query", query))
		return res, err
	}

	err = mapstructure.Decode(result.Record().Values(), &res)
	return res, err
}

func getProtocolDecimals(protocol string) float64 {
	switch protocol {
	case CCTP:
		return 1e6
	default:
		return 1e8
	}
}
