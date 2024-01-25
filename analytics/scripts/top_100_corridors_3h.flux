import "date"

option task = {
    name: "top 100 corridors with 3-hour granularity",
    every: 3h,
}

sourceBucket = "wormscan"
destinationBucket = "wormscan-24hours"
execution = date.truncate(t: now(), unit: 1h)
start7d = date.truncate(t: -7d, unit: 1h)
start2d = date.truncate(t: -2d, unit: 1h)

from(bucket: sourceBucket)
  |> range(start: start7d)
  |> filter(fn: (r) => r._measurement == "vaa_volume_v2" and r._field == "volume")
  |> group(columns: ["emitter_chain", "destination_chain", "token_chain", "token_address"])
  |> count(column: "_value")
  |> group()
  |> sort(desc:true)
  |> limit(n:100)
  |> map(fn: (r) => ({r with _time: execution}))
  |> set(key: "_measurement", value: "top_100_corridors_7_days_3h_v2")
  |> set(key: "_field", value: "count")
  |> to(bucket: destinationBucket)

from(bucket: sourceBucket)
  |> range(start: start2d)
  |> filter(fn: (r) => r._measurement == "vaa_volume_v2" and r._field == "volume")
  |> group(columns: ["emitter_chain", "destination_chain", "token_chain", "token_address"])
  |> count(column: "_value")
  |> group()
  |> sort(desc:true)
  |> limit(n:100)
  |> map(fn: (r) => ({r with _time: execution}))
  |> set(key: "_measurement", value: "top_100_corridors_2_days_3h_v2")
  |> set(key: "_field", value: "count")
  |> to(bucket: destinationBucket)

