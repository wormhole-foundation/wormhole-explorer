import "date"
import "influxdata/influxdb/schema"
import "json"

option task = {
    name: "assets by symbol for 7 days with 3-hour granularity",
    every: 3h,
}

sourceBucket = "wormscan"
destinationBucket = "wormscan-24hours"
measurement = "assets_by_symbol_7_days_3h_v2"
start = date.truncate(t: -7d, unit: 24h)
execution = date.truncate(t: now(), unit: 1h)


from(bucket: sourceBucket)
    |> range(start: start)
    |> filter(fn: (r) => r._measurement == "vaa_volume_v2" and (r._field == "symbol" or r._field == "volume"))
    |> schema.fieldsAsCols()
    |> filter(fn: (r) => r.symbol != "")
    |> map(fn: (r) => ({r with _value: r.volume}))
    |> group(columns: ["symbol","emitter_chain", "token_address", "token_chain"])
    |> reduce(
        fn: (r, accumulator) => ({
            volume: accumulator.volume + r._value,
            count: accumulator.count + 1,
            }),
            identity: {volume: uint(v: 0), count: 0}
    )
    |> group()
    |> map(fn: (r) => ({r with _time: execution, _field: "txs_volume", _value: string(v: json.encode(v: {"txs": r.count, "volume": r.volume}))}))
    |> drop(columns: ["volume", "count"])
    |> set(key: "_measurement", value: measurement)
    |> to(bucket: destinationBucket)
