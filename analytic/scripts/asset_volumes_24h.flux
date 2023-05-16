import "date"

option task = {
    name: "asset volume with 24-hour granularity",
    every: 24h,
}

start = date.sub(from: now(), d: 24h)
stop = now()

from(bucket: "wormscan")
    |> range(start: start, stop: stop)
    |> filter(fn: (r) => r["_measurement"] == "vaa_volume")
    |> filter(fn: (r) => r["_field"] == "volume")
    |> drop(columns: ["app_id", "destination_address", "destination_chain", "symbol"])
    |> group(columns: ["emitter_chain", "token_address", "token_chain"])
    |> sum(column: "_value")
    |> set(key: "_measurement", value: "asset_volumes_24h")
    |> set(key: "_field", value: "volume")
    |> map(fn: (r) => ({r with _time: start}))
    |> to(bucket: "wormscan-30days")