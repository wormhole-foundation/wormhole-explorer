package stats

import (
	"fmt"
	"time"
)

const queryTemplateSymbolWithAssets = `
from(bucket: "%s")
    |> range(start: %s)
    |> filter(fn: (r) => r._measurement == "%s" and r._field == "txs_volume")
    |> last()
    |> group()
`

func buildSymbolWithAssets(bucket string, t time.Time, measurement string) string {
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	return fmt.Sprintf(queryTemplateSymbolWithAssets, bucket, start, measurement)
}

const queryTemplateTopCorridors = `
from(bucket: "%s")
    |> range(start: %s)
    |> filter(fn: (r) => r._measurement == "%s" and r._field == "count")
    |> last()
    |> group()
`

func buildTopCorridors(bucket string, t time.Time, measurement string) string {
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	return fmt.Sprintf(queryTemplateTopCorridors, bucket, start, measurement)
}

const queryTemplateNTTTotalValueTokenTransferred = `
import "influxdata/influxdb/schema"
import "date"

bucket = "%s"
today =  %s

last = from(bucket: bucket)
  |> range(start: 1970-01-01T00:00:00Z, stop: today)
  |> filter(fn: (r) => r._measurement == "ntt_symbol_chain_1d" and r._field == "total_volume_transferred" )
	|> filter(fn: (r) => r.symbol == "W" and r.emitter_chain != r.destination_chain)
	|> group()
	|> sum()
	|> toFloat()
	|> map(fn: (r) => ({r with _value: r._value / 100000000.0}))

current = from(bucket: bucket)
  |> range(start: today)
	|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
	|> filter(fn: (r) => r.app_id_1 == "NATIVE_TOKEN_TRANSFER" or r.app_id_2 == "NATIVE_TOKEN_TRANSFER" or r.app_id_3 == "NATIVE_TOKEN_TRANSFER")
	|> filter(fn: (r) => (r._field == "symbol" and r._value != "") or r._field == "volume")
	|> schema.fieldsAsCols()
	|> filter(fn: (r) => r.symbol == "%s")
	|> map(fn: (r) => ({r with _value: r.volume}))
	|> group()
	|> sum()
	|> toFloat()
	|> map(fn: (r) => ({r with _value: r._value / 100000000.0}))
	
	
union(tables: [current, last])
  |> group()
  |> sum()
`

func buildNTTTotalValueTokenTransferred(bucket string, t time.Time, symbol string) string {
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	return fmt.Sprintf(queryTemplateNTTTotalValueTokenTransferred, bucket, start, symbol)
}

const queryTemplateNTTTotalTokenTransferred = `
import "influxdata/influxdb/schema"
import "date"

bucket = "%s"
today =  %s

last = from(bucket: bucket)
  |> range(start: 1970-01-01T00:00:00Z, stop: today)
  |> filter(fn: (r) => r._measurement == "ntt_symbol_chain_1d" and r._field == "total_transferred" )
	|> filter(fn: (r) => r.symbol == "W" and r.emitter_chain != r.destination_chain)
	|> group()
	|> sum()

current = from(bucket: bucket)
  |> range(start: today)
	|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
	|> filter(fn: (r) => r.app_id_1 == "NATIVE_TOKEN_TRANSFER" or r.app_id_2 == "NATIVE_TOKEN_TRANSFER" or r.app_id_3 == "NATIVE_TOKEN_TRANSFER")
	|> filter(fn: (r) => (r._field == "symbol" and r._value != "") or r._field == "volume")
	|> schema.fieldsAsCols()
	|> filter(fn: (r) => r.symbol == "%s")
	|> map(fn: (r) => ({r with _value: r.volume}))
	|> group()
	|> count()
	
	
union(tables: [current, last])
  |> group()
  |> sum()
`

func buildNTTTotalTokenTransferred(bucket string, t time.Time, symbol string) string {
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	return fmt.Sprintf(queryTemplateNTTTotalTokenTransferred, bucket, start, symbol)
}

const queryTemplateNTTAverageTransferSize = `
import "influxdata/influxdb/schema"
import "date"

bucket = "%s"

from(bucket: bucket)
  |> range(start: 2021-01-01T00:00:00Z)
	|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
	|> filter(fn: (r) => r.app_id_1 == "NATIVE_TOKEN_TRANSFER" or r.app_id_2 == "NATIVE_TOKEN_TRANSFER" or r.app_id_3 == "NATIVE_TOKEN_TRANSFER")
	|> filter(fn: (r) => (r._field == "symbol" and r._value != "") or r._field == "volume")
	|> schema.fieldsAsCols()
	|> filter(fn: (r) => r.symbol == "%s")
	|> map(fn: (r) => ({r with _value: r.volume}))
	|> group()
	|> mean()
	|> toFloat()
	|> map(fn: (r) => ({r with _value: r._value / 100000000.0}))
`

func buildNTTAverageTransferSize(bucket string, symbol string) string {
	return fmt.Sprintf(queryTemplateNTTAverageTransferSize, bucket, symbol)
}

const queryTemplateNTTChainActivity = `
import "influxdata/influxdb/schema"
import "strings"

bucket = "%s"
today = %s
field = "%s"

last = from(bucket: bucket)
    |> range(start: 1970-01-01T00:00:00Z, stop: today)
    |> filter(fn: (r) => r._measurement == "ntt_symbol_chain_1d" and r._field == field)
    |> filter(fn: (r) => %s)
	|> group(columns:["symbol","emitter_chain","destination_chain"])
	|> sum()

current = from(bucket: bucket)
    |> range(start: today)
    |> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
    |> filter(fn: (r) => r.app_id_1 == "NATIVE_TOKEN_TRANSFER" or r.app_id_2 == "NATIVE_TOKEN_TRANSFER" or r.app_id_3 == "NATIVE_TOKEN_TRANSFER")
    |> filter(fn: (r) => (r._field == "symbol" and r._value != "") or r._field == "volume")
    |> schema.fieldsAsCols()
    |> filter(fn: (r) => r.symbol != "")
    |> map(fn: (r) => ({r with symbol: strings.toUpper(v: r.symbol)}))
    |> filter(fn: (r) => %s)
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> map(fn: (r) => ({r with _value: r.volume}))
    |> %s()
	
union(tables: [current, last])
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> sum()
`

func buildNTTChainActivity(bucket string, t time.Time, symbol string, isNotional bool) string {
	filterCondition := fmt.Sprintf(`r.symbol == "%s"`, symbol)
	if symbol == "" {
		filterCondition = "true"
	}
	field := "total_transferred"
	aggregation := "count"
	if isNotional {
		field = "total_volume_transferred"
		aggregation = "sum"
	}
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	return fmt.Sprintf(queryTemplateNTTChainActivity, bucket, start, field, filterCondition, filterCondition, aggregation)
}

const queryTemplateNTTChainActivityByTime = `
start = %s
stop =  %s
bucket = "%s"
symbol = "%s"

from(bucket: bucket)
		|> range(start: start, stop: stop)
		|> filter(fn: (r) => r._measurement == "ntt_symbol_chain_1d" and r._field == "%s" )
		|> filter(fn: (r) => r.symbol == symbol)
		|> group(columns: ["symbol"])
		|> aggregateWindow(every: %s, fn: %s, createEmpty: true)`

func buildNTTChainActivityByTime(bucket string, start, stop, symbol string, isNotional bool, every string) string {
	aggregation := "count"
	field := "total_transferred"
	if isNotional {
		aggregation = "sum"
		field = "total_volume_transferred"
	}
	return fmt.Sprintf(queryTemplateNTTChainActivityByTime, start, stop, bucket, symbol, field, every, aggregation)
}
