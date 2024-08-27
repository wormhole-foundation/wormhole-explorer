package stats

import (
	"fmt"
	"time"
)

const queryTemplateNTTTopAddress = `
import "influxdata/influxdb/schema"
import "date"
import "strings"

bucket = "%s"
start = %s
symbol = "%s"

last = from(bucket: bucket)
  |> range(start: 1970-01-01T00:00:00Z, stop: start)
  |> filter(fn: (r) => r._measurement == "ntt_symbol_address_1d" and r._field == "%s" )
  |> filter(fn: (r) => r.symbol == symbol)
  |> group(columns:["from_address"])

current = from(bucket: bucket)
	|> range(start: start)
	|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
	|> filter(fn: (r) => r.app_id_1 == "NATIVE_TOKEN_TRANSFER" or r.app_id_2 == "NATIVE_TOKEN_TRANSFER" or r.app_id_3 == "NATIVE_TOKEN_TRANSFER")
	|> filter(fn: (r) => (r._field == "symbol" and r._value != "") or r._field == "volume" or r._field == "from_address")
	|> schema.fieldsAsCols()
	|> filter(fn: (r) => r.symbol != "" and r.from_address != "")
	|> map(fn: (r) => ({r with symbol: strings.toUpper(v: r.symbol), _value: r.volume}))
	|> filter(fn: (r) => r.symbol == symbol)
	|> group(columns:["from_address"])
	|> %s()

union(tables: [current, last])
	|> sum()
  	|> group()
	|> sort(columns:["_value"], desc:true)
	|> limit(n:10)
`

func buildNTTTopAddress(bucket string, symbol string, isNotional bool, t time.Time) string {
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	field := "total_transferred"
	aggregation := "count"
	if isNotional {
		aggregation = "sum"
		field = "total_volume_transferred"
	}

	return fmt.Sprintf(queryTemplateNTTTopAddress, bucket, start, symbol, field, aggregation)
}
