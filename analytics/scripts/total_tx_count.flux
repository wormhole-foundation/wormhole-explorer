import "date"

option task = {
    name: "total tx count by portal bridge",
    every: 24h,
}

stop = date.truncate(t: now(), unit: 24h)

from(bucket: "wormscan")
  |> range(start: 1970-01-01T00:00:00Z, stop: stop)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group()
  |> count()
  |> map(fn: (r) => ({ _time: r._stop, _value: r._value, _measurement: "total_tx_count", _field: "value" }))
  |> to(bucket: "wormscan-30days")
