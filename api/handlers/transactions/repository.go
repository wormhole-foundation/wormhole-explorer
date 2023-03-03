package transactions

import (
	"context"
	"fmt"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
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

type Repository struct {
	influxCli influxdb2.Client
	queryAPI  api.QueryAPI
	bucket    string
	logger *zap.Logger
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

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"go.uber.org/zap"
)

// GetLastTrx get the last transactions.
func (r *Repository) GetLastTrx(timeSpan string, sampleRate string) ([]string, error) {
	client := influxdb2.NewClient("http://localhost:8086", "NCJOkIcBbMD2dFPG5DOsuArzBPAjB3uIxrsvRK66jnxxgAEC0R8qxGICuH1VrgxoaSJ3a1eF5w_cpyzMdda28A==")
	queryAPI := client.QueryAPI("my-org")

	// query
	query := `from(bucket: "wormscan") |> range(start: -30d) |> filter(fn: (r) => r["_measurement"] == "vaa_count") |> group() |> aggregateWindow(every: 10m, fn: count, createEmpty: false)`

	// Get parser flux query result
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		r.logger.Error(err.Error())
		panic(err)
	}
	// Iterate over query result
	for result.Next() {
		// Notice when group key has changed
		if result.TableChanged() {
			r.logger.Info(fmt.Sprintf("table: %s\n", result.TableMetadata().String()))
		}
		// Access data
		r.logger.Info(fmt.Sprintf("value: %v\n", result.Record().String()))
		// Access data
		r.logger.Info(fmt.Sprintf("value: %v\n", result.Record().Value()))
	}
	if result.Err() != nil {
		r.logger.Error(result.Err().Error())
		panic(result.Err())
	}
	result.Close()
	client.Close()
	return []string{}, nil
}
