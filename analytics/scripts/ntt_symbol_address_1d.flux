import "influxdata/influxdb/schema"
import "date"
import "strings"


option task = {
    name: "calculate the total quantity and total volume by symbol and from address for the ntt protocol every day",
    every: 1d,
}

stop = date.truncate(t: now(), unit: 1d)
start = date.sub(d: 1d, from: stop)

bucketInfinite = "wormscan"
sourceBucket = bucketInfinite
toBucket = bucketInfinite
measurement = "ntt_symbol_address_1d"

ntt = from(bucket: sourceBucket)
        |> range(start: start, stop: stop)
        |> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
        |> filter(fn: (r) => r.app_id_1 == "NATIVE_TOKEN_TRANSFER" or r.app_id_2 == "NATIVE_TOKEN_TRANSFER" or r.app_id_3 == "NATIVE_TOKEN_TRANSFER")
        |> filter(fn: (r) => (r._field == "symbol" and r._value != "") or r._field == "volume" or r._field == "from_address")
        |> schema.fieldsAsCols()
        |> filter(fn: (r) => r.symbol != "" and r.from_address != "")
        |> map(fn: (r) => ({r with symbol: strings.toUpper(v: r.symbol)}))
        |> group(columns:["symbol","from_address"])
	
ntt 
        |> sum(column: "volume")
        |> set(key: "_field", value: "total_volume_transferred")
        |> map(fn: (r) => ({r with _time: start, _value: r.volume}))
        |> drop(columns: ["volume"])
        |> set(key: "_measurement", value: measurement)
        |> to(bucket: toBucket)
	
ntt 
        |> count(column: "volume")
        |> set(key: "_field", value: "total_transferred")
        |> map(fn: (r) => ({r with _time: start, _value: r.volume}))
        |> drop(columns: ["volume"])
        |> set(key: "_measurement", value: measurement)
        |> to(bucket: toBucket)