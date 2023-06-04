import "date"

option task = {
    name: "vaa_count grouped by hour",
    every: 1h,
}

start = date.truncate(t: -1h, unit: 1h)
stop = date.truncate(t: now(), unit: 1h)

from(bucket: "wormscan-30days")
    |> range(start: start, stop: stop)
    |> filter(fn: (r) => r["_measurement"] == "vaa_count")
    |> group()
    |> aggregateWindow(every: 1h, fn: count, createEmpty: true)
    |> set(key: "_measurement", value: "vaa_count_1h")
    |> to(bucket: "wormscan-30days", fieldFn: (r) => ({"count": r._value}))
