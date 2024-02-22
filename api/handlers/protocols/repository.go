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
  |> range(start: -1d)
  |> filter(fn: (r) => r._measurement == "%s" and r.protocol == "%s" and r.version == "%s")
  |> keep(columns: ["_time","_field","protocol", "_value", "total_value_secure", "total_value_transferred"])
  |> last()
  |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
`

type Repository struct {
	queryAPI        QueryDoer
	logger          *zap.Logger
	statsBucket     string
	activityBucket  string
	statsVersion    string
	activityVersion string
}

type rowStat struct {
	Protocol         string  `mapstructure:"protocol"`
	TotalMessages    uint64  `mapstructure:"total_messages"`
	TotalValueLocked float64 `mapstructure:"total_value_locked"`
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

func NewRepository(qApi QueryDoer, statsBucket, activityBucket, statsVersion, activityVersion string, logger *zap.Logger) *Repository {
	return &Repository{
		queryAPI:        qApi,
		statsBucket:     statsBucket,
		activityBucket:  activityBucket,
		statsVersion:    statsVersion,
		activityVersion: activityVersion,
		logger:          logger,
	}
}

func (q *queryApiWrapper) Query(ctx context.Context, query string) (QueryResult, error) {
	return q.qApi.Query(ctx, query)
}

// returns latest and last 24 hr stats for a given protocol
func (r *Repository) getProtocolStats(ctx context.Context, contributor string) (stats, error) {
	// fetch latest stat
	latest, err := fetchSingleRecordData[rowStat](r.logger, r.queryAPI, ctx, r.statsBucket, QueryTemplateLatestPoint, dbconsts.ProtocolsStatsMeasurement, contributor, r.statsVersion)
	if err != nil {
		return stats{}, err
	}
	// fetch last 24 hr stat
	last24hr, err := fetchSingleRecordData[rowStat](r.logger, r.queryAPI, ctx, r.statsBucket, QueryTemplateLast24Point, dbconsts.ProtocolsStatsMeasurement, contributor, r.statsVersion)
	return stats{
		Latest: latest,
		Last24: last24hr,
	}, err
}

func (r *Repository) getProtocolActivity(ctx context.Context, contributor string) (rowActivity, error) {
	return fetchSingleRecordData[rowActivity](r.logger, r.queryAPI, ctx, r.activityBucket, QueryTemplateActivityLatestPoint, dbconsts.ProtocolsActivityMeasurement, contributor, r.activityVersion)
}

func fetchSingleRecordData[T any](logger *zap.Logger, queryAPI QueryDoer, ctx context.Context, bucket, queryTemplate, measurement, contributor, version string) (T, error) {
	var res T
	q := buildQuery(queryTemplate, bucket, measurement, contributor, version)
	result, err := queryAPI.Query(ctx, q)
	if err != nil {
		logger.Error("error executing query to fetch data", zap.Error(err), zap.String("protocol", contributor), zap.String("query", q))
		return res, err
	}
	defer result.Close()

	if !result.Next() {
		if result.Err() != nil {
			logger.Error("error reading query response", zap.Error(result.Err()), zap.String("protocol", contributor), zap.String("query", q))
			return res, result.Err()
		}
		logger.Info("empty query response", zap.String("protocol", contributor), zap.String("query", q))
		return res, err
	}

	err = mapstructure.Decode(result.Record().Values(), &res)
	return res, err
}

func buildQuery(queryTemplate, bucket, measurement, contributorName, version string) string {
	return fmt.Sprintf(queryTemplate, bucket, measurement, contributorName, version)
}
