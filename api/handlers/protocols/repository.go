package protocols

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/mitchellh/mapstructure"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbconsts"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"time"
)

const AllProtocolsDeltaLastDay = `
		import "date"
		import "types"
		import "strings"

		ts = date.truncate(t: now(), unit: 1h)
		yesterday = date.sub(d: 1d, from: ts)
		
		data = from(bucket: "%s")
				|> range(start: yesterday,stop:ts)
				|> filter(fn: (r) => r._measurement == "protocols_stats_totals_1h")
				|> drop(columns: ["emitter_chain","destination_chain"])
		
		tvt = data
			|> filter(fn : (r) => r._field == "total_value_transferred")
			|> group(columns:["app_id"])
			|> sum()
			|> set(key:"_field",value:"total_value_transferred")
			|> map(fn: (r) => ({r with _value: int(v: r._value)}))
		
		totalMsgs =	data
					|> filter(fn : (r) => r._field == "total_messages")
					|> group(columns:["app_id"])
					|> sum()
					|> set(key:"_field",value:"total_messages")
					|> map(fn: (r) => ({r with _value: int(v: r._value)}))
		
		union(tables:[tvt,totalMsgs])
			|> set(key:"_time",value:string(v:yesterday))
			|> pivot(rowKey:["_time","app_id"], columnKey: ["_field"], valueColumn: "_value")
			|> map(fn: (r) => ({r with app_id: strings.trimPrefix(v: r.app_id, prefix: "TOTAL_")}))
`

const AllProtocolStats24HrAgo = `
	import "date"
	import "strings"

	startOfCurrentDay = date.truncate(t: now(), unit: 1d)
	
	data =	from(bucket: "%s")
				|> range(start: 1970-01-01T00:00:00Z,stop:startOfCurrentDay)
				|> filter(fn: (r) => r._measurement == "protocols_stats_totals_1d" and r.version == "v1")
				|> drop(columns: ["emitter_chain","destination_chain","version"])
	
	tvt = data	
			|> filter(fn : (r) => r._field == "total_value_transferred")
			|> group(columns:["app_id"])
			|> sum()
			|> set(key:"_field",value:"total_value_transferred")
			|> map(fn: (r) => ({r with _value: int(v: r._value)}))
	
	totalMsgs =	data	
				|> filter(fn : (r) => r._field == "total_messages")
				|> group(columns:["app_id"])
				|> sum()
				|> set(key:"_field",value:"total_messages")
				|> map(fn: (r) => ({r with _value: int(v: r._value)}))
	
	union(tables:[tvt,totalMsgs])		
		|> set(key:"_time",value:string(v:startOfCurrentDay))
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> map(fn: (r) => ({r with app_id: strings.trimPrefix(v: r.app_id, prefix: "TOTAL_")}))
`

const AllProtocolsDeltaSinceStartOfDay = `
	import "date"
	import "strings"

	ts = date.truncate(t: now(), unit: 1h)
	startOfDay = date.truncate(t: now(), unit: 1d)
	
	data = from(bucket: "%s")
			|> range(start: startOfDay,stop:ts)
			|> filter(fn: (r) => r._measurement == "protocols_stats_totals_1h")
			|> drop(columns: ["emitter_chain","destination_chain","version"])
	
	tvt = data	
			|> filter(fn : (r) => r._field == "total_value_transferred")
			|> group(columns:["app_id"])
			|> sum()
			|> set(key:"_field",value:"total_value_transferred")
			|> map(fn: (r) => ({r with _value: int(v: r._value)}))
	
	totalMsgs =	data	
					|> filter(fn : (r) => r._field == "total_messages")
					|> group(columns:["app_id"])
					|> sum()	
					|> set(key:"_field",value:"total_messages")
					|> map(fn: (r) => ({r with _value: int(v: r._value)}))
	
	union(tables:[tvt,totalMsgs])
		|> set(key:"_time",value:string(v:startOfDay))
		|> pivot(rowKey:["_time","app_id"], columnKey: ["_field"], valueColumn: "_value")
		|> map(fn: (r) => ({r with app_id: strings.trimPrefix(v: r.app_id, prefix: "TOTAL_")}))
`

const QueryTemplateProtocolStats24HrAgo = `
		from(bucket: "%s")
  			|> range(start: -1d)
  			|> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s")
			|> first()
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
`

const QueryTemplateProtocolStatsNow = `
		from(bucket: "%s")
  			|> range(start: -2d)
  			|> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s")
			|> last()
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
`

const QueryTemplateProtocolActivity = `	
	data = 
		from(bucket: "%s")
		  |> range(start: %s)
		  |> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s")
	
		tvt = data	
			|> filter(fn: (r) => r._field == "total_value_transferred") 
			|> sum()

		txs = data	
  			|> filter(fn: (r) => r._field == "txs")
  			|> sum()

		union(tables:[tvt, txs])
			|> pivot(rowKey:["_start","_stop"], columnKey: ["_field"], valueColumn: "_value")
			|> rename(columns: {_stop: "_time"})
`

const QueryLast24HrActivity = `
	import "date"
	
	from(bucket: "%s")
		|> range(start: -5d)
		|> filter(fn: (r) => r._measurement == "%s")
		|> filter(fn: (r) => r.protocol == "%s")
		|> last()
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
	Protocol                      string    `mapstructure:"protocol"`
	Time                          time.Time `mapstructure:"_time"`
	TotalValueTransferred         float64   `mapstructure:"total_value_transferred"`
	Txs                           uint64    `mapstructure:"txs"`
	Last24HrTotalValueTransferred float64
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

func (r *Repository) getProtocolStatsNow(ctx context.Context, protocol string) (rowStat, error) {
	q := fmt.Sprintf(QueryTemplateProtocolStatsNow, r.bucket30d, dbconsts.ProtocolsStatsMeasurementHourly, protocol)
	return fetchSingleRecord[rowStat](ctx, r.logger, r.queryAPI, q, protocol)
}

func (r *Repository) getProtocolStats24hrAgo(ctx context.Context, protocol string) (rowStat, error) {
	q := fmt.Sprintf(QueryTemplateProtocolStats24HrAgo, r.bucket30d, dbconsts.ProtocolsStatsMeasurementHourly, protocol)
	return fetchSingleRecord[rowStat](ctx, r.logger, r.queryAPI, q, protocol)
}

func (r *Repository) getAllbridgeActivity(ctx context.Context) (rowActivity, error) {

	const allbridge = "allbridge"

	q := fmt.Sprintf(QueryTemplateProtocolActivity, r.bucketInfinite, "1970-01-01T00:00:00Z", dbconsts.ProtocolsActivityMeasurementDaily, allbridge)
	activityDaily, err := fetchSingleRecord[rowActivity](ctx, r.logger, r.queryAPI, q, allbridge)
	if err != nil {
		r.logger.Error("error fetching latest daily activity", zap.Error(err))
		return rowActivity{}, err
	}
	startOfDay := time.Now().UTC().Truncate(24 * time.Hour).Format(time.RFC3339)
	q = fmt.Sprintf(QueryTemplateProtocolActivity, r.bucket30d, startOfDay, dbconsts.ProtocolsActivityMeasurementHourly, allbridge)
	activityHourly, err := fetchSingleRecord[rowActivity](ctx, r.logger, r.queryAPI, q, allbridge)
	if err != nil {
		r.logger.Error("error fetching latest hourly activity", zap.Error(err))
		return rowActivity{}, err
	}

	q = fmt.Sprintf(QueryLast24HrActivity, r.bucketInfinite, dbconsts.ProtocolsActivityMeasurementDaily, allbridge)
	last24HrActivity, err := fetchSingleRecord[rowActivity](ctx, r.logger, r.queryAPI, q, allbridge)
	if err != nil {
		r.logger.Error("error fetching last 24 hr activity", zap.Error(err))
		return rowActivity{}, err
	}

	return rowActivity{
		Protocol:                      allbridge,
		Txs:                           activityDaily.Txs + activityHourly.Txs,
		TotalValueTransferred:         activityDaily.TotalValueTransferred + activityHourly.TotalValueTransferred,
		Last24HrTotalValueTransferred: last24HrActivity.TotalValueTransferred,
	}, nil
}

func (r *Repository) getAllProtocolStats(ctx context.Context) ([]intStats, error) {
	// calculate total values till the start of current day
	totalTillCurrentDayQuery := fmt.Sprintf(AllProtocolStats24HrAgo, r.bucketInfinite)
	recordsTillCurrentDay, err := fetchMultipleRecords[intRowStat](ctx, r.logger, r.queryAPI, totalTillCurrentDayQuery)
	if err != nil {
		return nil, err
	}

	// calculate delta since the beginning of current day
	currentDayStatsQuery := fmt.Sprintf(AllProtocolsDeltaSinceStartOfDay, r.bucket30d)
	recordsCurrentDay, err := fetchMultipleRecords[intRowStat](ctx, r.logger, r.queryAPI, currentDayStatsQuery)
	if err != nil {
		return nil, err
	}

	latestTotal := mergeStats(recordsTillCurrentDay, recordsCurrentDay)

	q3 := fmt.Sprintf(AllProtocolsDeltaLastDay, r.bucket30d)
	deltaYesterdayStats, errQ3 := fetchMultipleRecords[intRowStat](ctx, r.logger, r.queryAPI, q3)
	if errQ3 != nil {
		return nil, errQ3
	}
	for i := 0; i < len(deltaYesterdayStats); i++ {
		deltaYesterdayStats[i].TotalValueTransferred = deltaYesterdayStats[i].TotalValueTransferred / 1e8
	}

	result := make(map[string]intStats, len(latestTotal))
	for _, v := range latestTotal {
		result[v.Protocol] = intStats{
			Latest: v,
		}
	}
	for _, v := range deltaYesterdayStats {
		if total, ok := result[v.Protocol]; ok {
			total.DeltaLast24hr = v
			result[v.Protocol] = total
		}
	}

	return maps.Values(result), nil
}

func mergeStats(totalTillStartOfToday []intRowStat, currentDay []intRowStat) []intRowStat {
	protocolToStats := make(map[string]intRowStat, len(totalTillStartOfToday))
	for _, total := range totalTillStartOfToday {
		protocolToStats[total.Protocol] = total
	}
	for _, dayStats := range currentDay {
		if total, ok := protocolToStats[dayStats.Protocol]; ok {
			total.TotalMessages += dayStats.TotalMessages
			total.TotalValueTransferred += dayStats.TotalValueTransferred
			protocolToStats[dayStats.Protocol] = total
		} else {
			protocolToStats[dayStats.Protocol] = dayStats
		}
	}
	result := make([]intRowStat, 0, len(protocolToStats))
	for _, v := range protocolToStats {
		v.TotalValueTransferred = v.TotalValueTransferred / 1e8
		result = append(result, v)
	}
	return result
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
	statsData, err := fetchSingleRecord[intRowStat](ctx, r.logger, r.queryAPI, q, protocol)
	if err != nil {
		r.logger.Error("error fetching cctp totals stats", zap.Error(err))
		return intStats{}, err
	}

	q = fmt.Sprintf(queryTemplate, r.bucket24Hrs, "|> first()")
	totals24HrAgo, err := fetchSingleRecord[intRowStat](ctx, r.logger, r.queryAPI, q, protocol)
	if err != nil {
		r.logger.Error("error fetching cctp totals stats", zap.Error(err))
		return intStats{}, err
	}

	return intStats{
		Latest: intRowStat{
			Protocol:              protocol,
			TotalMessages:         statsData.TotalMessages,
			TotalValueTransferred: statsData.TotalValueTransferred / 1e6,
		},
		DeltaLast24hr: intRowStat{
			Protocol:              protocol,
			TotalMessages:         statsData.TotalMessages - totals24HrAgo.TotalMessages,
			TotalValueTransferred: (statsData.TotalValueTransferred - totals24HrAgo.TotalValueTransferred) / 1e6,
		},
	}, nil

}

func fetchMultipleRecords[T any](ctx context.Context, logger *zap.Logger, queryAPI QueryDoer, query string) ([]T, error) {

	result := make([]T, 0)
	resp, err := queryAPI.Query(ctx, query)
	if err != nil {
		logger.Error("error executing query to fetch data", zap.Error(err), zap.String("query", query))
		return result, err
	}
	defer resp.Close()

	for resp.Next() {
		if resp.Err() != nil {
			logger.Error("error reading query response", zap.Error(resp.Err()), zap.String("query", query))
			return result, resp.Err()
		}
		var res T
		err = mapstructure.Decode(resp.Record().Values(), &res)
		if err != nil {
			logger.Error("error decoding query response", zap.Error(err), zap.String("query", query))
			return result, err
		}
		result = append(result, res)
	}
	return result, nil
}

func fetchSingleRecord[T any](ctx context.Context, logger *zap.Logger, queryAPI QueryDoer, query, protocol string) (T, error) {
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
