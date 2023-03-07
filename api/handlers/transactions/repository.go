package transactions

import (
	"context"
	"fmt"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

const queryTemplate = `
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r._measurement == "vaa_volume" and r._field == "volume")
  |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
  |> %s(column: "volume")
`

const queryTemplateWithApps = `
from(bucket: "%s")
  |> range(start: %s, stop: %s)
  |> filter(fn: (r) => r._measurement == "vaa_volume")
  |> filter(fn: (r) => r._field == "volume" or  r._field == "app_id")
  |> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
  |> filter(fn: (r) => contains(value: r.app_id, set: %s))
  |> %s(column: "volume")
`

const queryTemplateVaaCount = `
from(bucket: "%s")
  |> range(start: -%s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count")
  |> group()
  |> aggregateWindow(every: %s, fn: count, createEmpty: true)
  |> map(fn:(r) => ( {_time: r._time, count: r._value}))
`

const queryTemplateVaaCountWithCumulativeSum = `
from(bucket: "%s")
  |> range(start: -%s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count")
  |> group()
  |> aggregateWindow(every: %s, fn: count, createEmpty: true)
  |> cumulativeSum()
  |> map(fn:(r) => ( {_time: r._time, count: r._value}))
`

type Repository struct {
	influxCli influxdb2.Client
	queryAPI  api.QueryAPI
	bucket    string
	logger    *zap.Logger
}

func NewRepository(client influxdb2.Client, org, bucket string, logger *zap.Logger) *Repository {
	queryAPI := client.QueryAPI(org)
	return &Repository{influxCli: client, queryAPI: queryAPI, bucket: bucket, logger: logger}
}

func (r *Repository) FindChainActivity(ctx context.Context, q *ChainActivityQuery) ([]ChainActivityResult, error) {
	query := r.buildFindVolumeQuery(q)
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
	return response, nil
}

func (r *Repository) buildFindVolumeQuery(q *ChainActivityQuery) string {
	start := q.GetStart().UTC().Format(time.RFC3339)
	stop := q.GetEnd().UTC().Format(time.RFC3339)
	var operation string
	if q.IsNotional {
		operation = "sum"
	} else {
		operation = "count"
	}
	if q.HasAppIDS() {
		apps := `["` + strings.Join(q.GetAppIDs(), `","`) + `"]`
		return fmt.Sprintf(queryTemplateWithApps, r.bucket, start, stop, apps, operation)
	}
	return fmt.Sprintf(queryTemplate, r.bucket, start, stop, operation)
}

// GetTransactionCount get the last transactions.
func (r *Repository) GetTransactionCount(ctx context.Context, q *TransactionCountQuery) ([]TransactionCountResult, error) {
	query := r.buildLastTrxQuery(q)
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
	return response, nil
}

func (r *Repository) buildLastTrxQuery(q *TransactionCountQuery) string {
	if q.CumulativeSum {
		return fmt.Sprintf(queryTemplateVaaCountWithCumulativeSum, r.bucket, q.TimeSpan, q.SampleRate)
	}
	return fmt.Sprintf(queryTemplateVaaCount, r.bucket, q.TimeSpan, q.SampleRate)
}
