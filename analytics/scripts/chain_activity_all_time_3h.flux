import "date"

option task = {
    name: "chain activity for all time with 3-hour granularity",
    every: 3h,
}

sourceBucket = "wormscan"
destinationBucket = "wormscan-24hours"
execution = date.truncate(t: now(), unit: 1h)

from(bucket: sourceBucket)
  |> range(start: 1970-01-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "vaa_volume" and r._field == "volume")
  |> group(columns: ["emitter_chain", "destination_chain", "app_id"])
  |> count(column: "_value")
  |> map(fn: (r) => ({r with _time: execution}))
  |> set(key: "_measurement", value: "chain_activity_all_time_3h")
  |> set(key: "_field", value: "count")
  |> to(bucket: destinationBucket)

from(bucket: sourceBucket)
  |> range(start: 1970-01-01T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "vaa_volume" and r._field == "volume")
  |> group(columns: ["emitter_chain", "destination_chain", "app_id"])
  |> sum(column: "_value")
  |> map(fn: (r) => ({r with _time: execution}))
  |> set(key: "_measurement", value: "chain_activity_all_time_3h")
  |> set(key: "_field", value: "notional")
  |> to(bucket: destinationBucket)
