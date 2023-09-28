import "date"

option task = {
    name: "asset volume with 24-hour granularity",
    every: 24h,
}

sourceBucket = "wormscan"
destinationBucket = "wormscan-30days"

start = date.truncate(t: -24h, unit: 24h)
stop = date.truncate(t: now(), unit: 24h)

from(bucket: sourceBucket)
    |> range(start: start, stop: stop)
    |> filter(fn: (r) => r["_measurement"] == "vaa_volume_v2")
    |> filter(fn: (r) => r["_field"] == "volume")
    |> group(columns: ["emitter_chain", "token_address", "token_chain"])
    |> sum(column: "_value")
    |> set(key: "_measurement", value: "asset_volumes_24h_v2")
    |> set(key: "_field", value: "volume")
    |> map(fn: (r) => ({r with _time: start}))
    |> to(bucket: destinationBucket)