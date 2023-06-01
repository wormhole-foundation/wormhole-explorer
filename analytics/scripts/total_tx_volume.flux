import "date"

option task = {
    name: "total tx volume by portal bridge",
    every: 24h,
}

stop = date.truncate(t: now(), unit: 24h)

from(bucket: "wormscan")
  |> range(start: 1970-01-01T00:00:00Z, stop: stop)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group()
  |> sum()
  |> map(fn: (r) => ({ _time: r._stop, _value: r._value, _measurement: "total_tx_volume", _field: "value" }))
  |> to(bucket: "wormscan-30days")
