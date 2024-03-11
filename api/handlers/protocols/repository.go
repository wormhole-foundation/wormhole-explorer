package protocols

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/mitchellh/mapstructure"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"go.uber.org/zap"
)

const QueryTemplateLatestPoint = `
from(bucket: "%s")
    |> range(start: -1d)
    |> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s" and r.version == "%s")
    |> last()
    |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
`

const QueryTemplateLast24Point = `
from(bucket: "%s")
    |> range(start: -1d)
    |> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s" and r.version == "%s")
    |> first()
	|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
`

const QueryTemplateActivityLatestPoint = `
from(bucket: "%s")
  |> range(start: 1970-01-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s" and r.version == "%s")
  |> keep(columns: ["_time","_field","protocol", "_value", "total_value_secure", "total_value_transferred"])
  |> last()
  |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
`

// QueryIntProtocolsTotalStartOfDay Query template for internal protocols (cctp and portal_token_bridge) to fetch total values till the start of current day
const QueryIntProtocolsTotalStartOfDay = `
		import "date"
		import "types"
		
		startOfCurrentDay = date.truncate(t: now(), unit: 1d)
		
	data =	from(bucket: "%s")
		|> range(start: 1970-01-01T00:00:00Z,stop:startOfCurrentDay)
		|> filter(fn: (r) => r._measurement == "%s" and r.app_id == "%s")
		
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
		
union(tables:[tvt,totalMsgs])		
		|> set(key:"_time",value:string(v:startOfCurrentDay))
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> set(key:"app_id",value:"%s")
`

// QueryIntProtocolsDeltaSinceStartOfDay calculate delta since the beginning of current day
const QueryIntProtocolsDeltaSinceStartOfDay = `
		import "date"
		import "types"

		ts = date.truncate(t: now(), unit: 1h)
		startOfDay = date.truncate(t: now(), unit: 1d)
		
	data =	from(bucket: "%s")
		|> range(start: startOfDay,stop:ts)
		|> filter(fn: (r) => r._measurement == "%s" and r.app_id == "%s")
		
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
		
union(tables:[tvt,totalMsgs])		
		|> set(key:"_time",value:string(v:startOfDay))
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> set(key:"app_id",value:"%s")
`

// QueryIntProtocolsDeltaLastDay calculate last day delta
const QueryIntProtocolsDeltaLastDay = `
		import "date"
		import "types"
		

		ts = date.truncate(t: now(), unit: 1h)
		yesterday = date.sub(d: 1d, from: ts)
		
	data =	from(bucket: "%s")
		|> range(start: yesterday,stop:ts)
		|> filter(fn: (r) => r._measurement == "%s" and r.app_id == "%s")
		
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
		
union(tables:[tvt,totalMsgs])		
		|> set(key:"_time",value:string(v:yesterday))
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> set(key:"app_id",value:"%s")
`

type Repository struct {
	queryAPI               QueryDoer
	logger                 *zap.Logger
	bucketInfinite         string
	bucket30d              string
	statsVersion           string
	activityVersion        string
	intProtocolMeasurement map[string]struct {
		Daily  string
		Hourly string
	}
}

type rowStat struct {
	Protocol         string  `mapstructure:"protocol"`
	TotalMessages    uint64  `mapstructure:"total_messages"`
	TotalValueLocked float64 `mapstructure:"total_value_locked"`
}

type intRowStat struct {
	Protocol              string `mapstructure:"app_id"`
	TotalMessages         uint64 `mapstructure:"total_messages"`
	TotalValueTransferred uint64 `mapstructure:"total_value_transferred"`
}

type intStats struct {
	Latest        intRowStat
	DeltaLast24hr intRowStat
}

type rowActivity struct {
	Protocol              string  `mapstructure:"protocol"`
	DestinationChainId    string  `mapstructure:"destination_chain_id"`
	EmitterChainId        string  `mapstructure:"emitter_chain_id"`
	From                  string  `mapstructure:"from"`
	TotalUsd              float64 `mapstructure:"total_usd"`
	TotalValueTransferred float64 `mapstructure:"total_value_transferred"`
	TotalVolumeSecure     float64 `mapstructure:"total_value_secure"`
	Txs                   uint64  `mapstructure:"txs"`
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

func NewRepository(qApi QueryDoer, bucketInfinite, bucket30d, statsVersion, activityVersion string, logger *zap.Logger) *Repository {
	return &Repository{
		queryAPI:        qApi,
		bucketInfinite:  bucketInfinite,
		bucket30d:       bucket30d,
		statsVersion:    statsVersion,
		activityVersion: activityVersion,
		logger:          logger,
		intProtocolMeasurement: map[string]struct {
			Daily  string
			Hourly string
		}{
			CCTP:              {Daily: dbconsts.CctpStatsMeasurementDaily, Hourly: dbconsts.CctpStatsMeasurementHourly},
			PortalTokenBridge: {Daily: dbconsts.TokenBridgeStatsMeasurementDaily, Hourly: dbconsts.TokenBridgeStatsMeasurementHourly},
		},
	}
}

func (q *queryApiWrapper) Query(ctx context.Context, query string) (QueryResult, error) {
	return q.qApi.Query(ctx, query)
}

// returns latest and last 24 hr stats for a given protocol
func (r *Repository) getProtocolStats(ctx context.Context, protocol string) (stats, error) {

	// fetch latest stat
	q := buildQuery(QueryTemplateLatestPoint, r.bucket30d, dbconsts.ProtocolsStatsMeasurement, protocol, r.statsVersion)
	latest, err := fetchSingleRecordData[rowStat](r.logger, r.queryAPI, ctx, q, protocol)
	if err != nil {
		return stats{}, err
	}
	// fetch last 24 hr stat
	q = buildQuery(QueryTemplateLast24Point, r.bucket30d, dbconsts.ProtocolsStatsMeasurement, protocol, r.statsVersion)
	last24hr, err := fetchSingleRecordData[rowStat](r.logger, r.queryAPI, ctx, q, protocol)
	return stats{
		Latest: latest,
		Last24: last24hr,
	}, err
}

func (r *Repository) getProtocolActivity(ctx context.Context, protocol string) (rowActivity, error) {
	q := buildQuery(QueryTemplateActivityLatestPoint, r.bucket30d, dbconsts.ProtocolsActivityMeasurement, protocol, r.activityVersion)
	return fetchSingleRecordData[rowActivity](r.logger, r.queryAPI, ctx, q, protocol)
}

// returns latest and last 24 hr for internal protocols (cctp and portal_token_bridge)
func (r *Repository) getInternalProtocolStats(ctx context.Context, protocol string) (intStats, error) {

	// calculate total values till the start of current day
	totalTillCurrentDayQuery := fmt.Sprintf(QueryIntProtocolsTotalStartOfDay, r.bucketInfinite, r.intProtocolMeasurement[protocol].Daily, protocol, protocol)
	totalsUntilToday, err := fetchSingleRecordData[intRowStat](r.logger, r.queryAPI, ctx, totalTillCurrentDayQuery, protocol)
	if err != nil {
		return intStats{}, err
	}

	// calculate delta since the beginning of current day
	q2 := fmt.Sprintf(QueryIntProtocolsDeltaSinceStartOfDay, r.bucket30d, r.intProtocolMeasurement[protocol].Hourly, protocol, protocol)
	currentDayStats, errCD := fetchSingleRecordData[intRowStat](r.logger, r.queryAPI, ctx, q2, protocol)
	if errCD != nil {
		return intStats{}, errCD
	}

	latestTotal := intRowStat{
		Protocol:              protocol,
		TotalMessages:         totalsUntilToday.TotalMessages + currentDayStats.TotalMessages,
		TotalValueTransferred: totalsUntilToday.TotalValueTransferred + currentDayStats.TotalValueTransferred,
	}

	result := intStats{
		Latest: latestTotal,
	}

	// calculate last day delta
	q3 := fmt.Sprintf(QueryIntProtocolsDeltaLastDay, r.bucket30d, r.intProtocolMeasurement[protocol].Hourly, protocol, protocol)
	deltaYesterdayStats, errQ3 := fetchSingleRecordData[intRowStat](r.logger, r.queryAPI, ctx, q3, protocol)
	if errQ3 != nil {
		return result, errQ3
	}

	result.DeltaLast24hr = deltaYesterdayStats
	return result, nil
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

func buildQuery(queryTemplate, bucket, measurement, contributorName, version string) string {
	return fmt.Sprintf(queryTemplate, bucket, measurement, contributorName, version)
}
