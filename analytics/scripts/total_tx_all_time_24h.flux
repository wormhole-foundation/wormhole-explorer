import "date"

option task = {
    name: "total tx for all time every 24-hours",
    every: 24h,
}

sourceBucket = "wormscan"
destinationBucket = "wormscan-30days"

stop = date.truncate(t: now(), unit: 24h)

from(bucket: sourceBucket)
  |> range(start: 1970-01-01T00:00:00Z, stop: stop)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume_v2")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group()
  |> count()
  |> map(fn: (r) => ({ _time: r._stop, _value: r._value, _measurement: "total_tx_count_v2", _field: "value" }))
  |> to(bucket: destinationBucket)


from(bucket: sourceBucket)
  |> range(start: 1970-01-01T00:00:00Z, stop: stop)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume_v2")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group()
  |> sum()
  |> map(fn: (r) => ({ _time: r._stop, _value: r._value, _measurement: "total_tx_volume_v2", _field: "value" }))
  |> to(bucket: destinationBucket)