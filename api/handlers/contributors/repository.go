package contributors

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type repository struct {
	queryAPI       api.QueryAPI
	logger         *zap.Logger
	statsBucket    string
	activityBucket string
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

// returns latest and last 24 hr stats for a given contributor
func (r *repository) getContributorStats(ctx context.Context, contributor string) (stats, error) {
	// fetch latest stat
	//latest, err := r.fetchStats(ctx, queryTemplateLatestPoint, "contributors_stats", contributor)
	latest, err := fetchData[rowStat](r.queryAPI, ctx, r.statsBucket, queryTemplateLatestPoint, "contributors_stats", contributor)
	if err != nil {
		return stats{}, err
	}
	// fetch last 24 hr stat
	last24hr, err := fetchData[rowStat](r.queryAPI, ctx, r.activityBucket, queryTemplateLast24Stat, "contributors_stats", contributor)
	return stats{
		Latest: latest,
		Last24: last24hr,
	}, err
}

func fetchData[T any](queryAPI api.QueryAPI, ctx context.Context, bucket, queryTemplate, measurement, contributor string) (T, error) {
	var res T
	query := buildQuery(queryTemplate, bucket, measurement, contributor)
	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return res, err
	}
	defer result.Close()

	result.Next()
	if result.Err() != nil {
		return res, result.Err()
	}
	err = mapstructure.Decode(result.Record().Values(), &res)
	return res, err
}

const queryTemplateLatestPoint = `
from(bucket: "%s")
    |> range(start: -24h)
    |> filter(fn: (r) => r._measurement == "%s" and r.contributor == "%s")
    |> last()
	|> findRecord(fn: (key) => true, idx: 0)
`

const queryTemplateLast24Stat = `
from(bucket: "%s")
    |> range(start: -24h)
    |> filter(fn: (r) => r._measurement == "%s" and r.contributor == "%s")
    |> first()
	|> findRecord(fn: (key) => true, idx: 0)
`

func buildQuery(queryTemplate, bucket, measurement, contributorName string) string {
	return fmt.Sprintf(queryTemplate, bucket, measurement, contributorName)
}

func (r *repository) getContributorActivity(ctx context.Context, contributor string) (rowActivity, error) {
	return fetchData[rowActivity](r.queryAPI, ctx, r.activityBucket, queryTemplateLatestPoint, "contributors_activity", contributor)
}
