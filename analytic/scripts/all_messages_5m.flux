import "date"

option task = {
    name: "count of all messages with 5-minute granularity",
    every: 5m,
}

start = date.sub(from: now(), d: 5m)
stop = now()

from(bucket: "wormscan-24hours")
  |> range(start: start, stop: stop)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count_all_messages")
  |> filter(fn: (r) => r["_field"] == "count")
  |> group()
  |> count()
  |> set(key: "_measurement", value: "vaa_count_all_messages_5m")
  |> set(key: "_field", value: "volume")
  |> map(fn: (r) => ({r with _time: start}))
  |> to(bucket: "wormscan-24hours")