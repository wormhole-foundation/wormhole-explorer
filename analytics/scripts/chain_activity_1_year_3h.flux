import "date"

option task = {
    name: "chain activity for 1 year with 3-hour granularity",
    every: 3h,
}

sourceBucket = "wormscan"
destinationBucket = "wormscan-24hours"
execution = date.truncate(t: now(), unit: 1h)
start = date.truncate(t: -1y, unit: 24h)

from(bucket: sourceBucket)
  |> range(start: start)
  |> filter(fn: (r) => r._measurement == "vaa_volume" and r._field == "volume")
  |> group(columns: ["emitter_chain", "destination_chain", "app_id"])
  |> count(column: "_value")
  |> map(fn: (r) => ({r with _time: execution}))
  |> set(key: "_measurement", value: "chain_activity_1_year_3h")
  |> set(key: "_field", value: "count")
  |> to(bucket: destinationBucket)

from(bucket: sourceBucket)
  |> range(start: start)
  |> filter(fn: (r) => r._measurement == "vaa_volume" and r._field == "volume")
  |> group(columns: ["emitter_chain", "destination_chain", "app_id"])
  |> sum(column: "_value")
  |> map(fn: (r) => ({r with _time: execution}))
  |> set(key: "_measurement", value: "chain_activity_1_year_3h")
  |> set(key: "_field", value: "notional")
  |> to(bucket: destinationBucket)
