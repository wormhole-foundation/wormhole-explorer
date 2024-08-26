package stats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueries_buildNTTTransferChainActivityWithSymbol(t *testing.T) {

	expected := `
import "influxdata/influxdb/schema"
import "strings"

bucket = "wormscan"
today = 2024-08-23T00:00:00Z
field = "total_transferred"

last = from(bucket: bucket)
    |> range(start: 1970-01-01T00:00:00Z, stop: today)
    |> filter(fn: (r) => r._measurement == "ntt_symbol_chain_1d" and r._field == field)
    |> filter(fn: (r) => r.symbol == "W")
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
    |> filter(fn: (r) => r.symbol == "W")
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> map(fn: (r) => ({r with _value: r.volume}))
    |> count()
	
union(tables: [current, last])
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> sum()
`
	//2023-08-23T18:39:10.985Z
	tm := time.Date(2024, 8, 23, 18, 39, 10, 985, time.UTC)
	actual := buildNTTChainActivity("wormscan", tm, "W", false)
	assert.Equal(t, expected, actual)
}

func TestQueries_buildNTTTransferChainActivityWithoutSymbol(t *testing.T) {

	expected := `
import "influxdata/influxdb/schema"
import "strings"

bucket = "wormscan"
today = 2024-08-23T00:00:00Z
field = "total_transferred"

last = from(bucket: bucket)
    |> range(start: 1970-01-01T00:00:00Z, stop: today)
    |> filter(fn: (r) => r._measurement == "ntt_symbol_chain_1d" and r._field == field)
    |> filter(fn: (r) => true)
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
    |> filter(fn: (r) => true)
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> map(fn: (r) => ({r with _value: r.volume}))
    |> count()
	
union(tables: [current, last])
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> sum()
`
	//2023-08-23T18:39:10.985Z
	tm := time.Date(2024, 8, 23, 18, 39, 10, 985, time.UTC)
	actual := buildNTTChainActivity("wormscan", tm, "", false)
	assert.Equal(t, expected, actual)
}

func TestQueries_buildNTTVolumeChainActivityWithSymbol(t *testing.T) {

	expected := `
import "influxdata/influxdb/schema"
import "strings"

bucket = "wormscan"
today = 2024-08-23T00:00:00Z
field = "total_volume_transferred"

last = from(bucket: bucket)
    |> range(start: 1970-01-01T00:00:00Z, stop: today)
    |> filter(fn: (r) => r._measurement == "ntt_symbol_chain_1d" and r._field == field)
    |> filter(fn: (r) => r.symbol == "W")
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
    |> filter(fn: (r) => r.symbol == "W")
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> map(fn: (r) => ({r with _value: r.volume}))
    |> sum()
	
union(tables: [current, last])
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> sum()
`
	//2023-08-23T18:39:10.985Z
	tm := time.Date(2024, 8, 23, 18, 39, 10, 985, time.UTC)
	actual := buildNTTChainActivity("wormscan", tm, "W", true)
	assert.Equal(t, expected, actual)
}

func TestQueries_buildNTTVolumeChainActivityWithoutSymbol(t *testing.T) {

	expected := `
import "influxdata/influxdb/schema"
import "strings"

bucket = "wormscan"
today = 2024-08-23T00:00:00Z
field = "total_volume_transferred"

last = from(bucket: bucket)
    |> range(start: 1970-01-01T00:00:00Z, stop: today)
    |> filter(fn: (r) => r._measurement == "ntt_symbol_chain_1d" and r._field == field)
    |> filter(fn: (r) => true)
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
    |> filter(fn: (r) => true)
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> map(fn: (r) => ({r with _value: r.volume}))
    |> sum()
	
union(tables: [current, last])
    |> group(columns:["symbol","emitter_chain","destination_chain"])
    |> sum()
`
	//2023-08-23T18:39:10.985Z
	tm := time.Date(2024, 8, 23, 18, 39, 10, 985, time.UTC)
	actual := buildNTTChainActivity("wormscan", tm, "", true)
	assert.Equal(t, expected, actual)
}
