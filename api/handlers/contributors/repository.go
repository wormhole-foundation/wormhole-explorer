package contributors

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

const QueryTemplateLatestPoint = `
from(bucket: "%s")
    |> range(start: -24h)
    |> filter(fn: (r) => r._measurement == "%s" and r.contributor == "%s" and r.version == "%s")
    |> last()
`

const QueryTemplateLast24Stat = `
from(bucket: "%s")
    |> range(start: -24h)
    |> filter(fn: (r) => r._measurement == "%s" and r.contributor == "%s" and r.version == "%s")
    |> first()
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
	Contributor      string `mapstructure:"contributor"`
	TotalMessages    string `mapstructure:"total_messages"`
	TotalValueLocked string `mapstructure:"total_value_locked"`
}

type rowActivity struct {
	Contributor           string `mapstructure:"contributor"`
	DestinationChainId    string `mapstructure:"destination_chain_id"`
	EmitterChainId        string `mapstructure:"emitter_chain_id"`
	From                  string `mapstructure:"from"`
	TotalUsd              string `mapstructure:"total_usd"`
	TotalValueTransferred string `mapstructure:"total_value_transferred"`
	TotalVolumeSecure     string `mapstructure:"total_volume_secure"`
	Txs                   string `mapstructure:"txs"`
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

// returns latest and last 24 hr stats for a given contributor
func (r *Repository) getContributorStats(ctx context.Context, contributor string) (stats, error) {
	// fetch latest stat
	latest, err := fetchSingleRecordData[rowStat](r.logger, r.queryAPI, ctx, r.statsBucket, QueryTemplateLatestPoint, "contributors_stats", contributor, r.statsVersion)
	if err != nil {
		return stats{}, err
	}
	// fetch last 24 hr stat
	last24hr, err := fetchSingleRecordData[rowStat](r.logger, r.queryAPI, ctx, r.activityBucket, QueryTemplateLast24Stat, "contributors_stats", contributor, r.statsVersion)
	return stats{
		Latest: latest,
		Last24: last24hr,
	}, err
}

func (r *Repository) getContributorActivity(ctx context.Context, contributor string) (rowActivity, error) {
	return fetchSingleRecordData[rowActivity](r.logger, r.queryAPI, ctx, r.activityBucket, QueryTemplateLatestPoint, "contributors_activity", contributor, r.activityVersion)
}

func fetchSingleRecordData[T any](logger *zap.Logger, queryAPI QueryDoer, ctx context.Context, bucket, queryTemplate, measurement, contributor, version string) (T, error) {
	var res T
	q := buildQuery(queryTemplate, bucket, measurement, contributor, version)
	result, err := queryAPI.Query(ctx, q)
	if err != nil {
		logger.Error("error executing query to fetch data", zap.Error(err), zap.String("contributor", contributor), zap.String("query", q))
		return res, err
	}
	defer result.Close()

	if !result.Next() {
		if result.Err() != nil {
			logger.Error("error reading query response", zap.Error(result.Err()), zap.String("contributor", contributor), zap.String("query", q))
			return res, result.Err()
		}
		logger.Info("empty query response", zap.String("contributor", contributor), zap.String("query", q))
		return res, err
	}

	err = mapstructure.Decode(result.Record().Values(), &res)
	return res, err
}

func buildQuery(queryTemplate, bucket, measurement, contributorName, version string) string {
	return fmt.Sprintf(queryTemplate, bucket, measurement, contributorName, version)
}
