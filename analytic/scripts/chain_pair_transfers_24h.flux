import "date"

option task = {
    name: "chain pair transfers with 24-hour granularity",
    every: 24h,
}

start = date.truncate(t: -24h, unit: 24h)
stop = date.truncate(t: now(), unit: 24h)

from(bucket: "wormscan")
    |> range(start: start, stop: stop)
    |> filter(fn: (r) => r["_measurement"] == "vaa_volume")
    |> filter(fn: (r) => r["_field"] == "volume")
    |> group(columns: ["emitter_chain", "destination_chain"])
    |> count(column: "_value")
    |> set(key: "_measurement", value: "chain_pair_transfers_24h")
    |> set(key: "_field", value: "num_transfers")
    |> map(fn: (r) => ({r with _time: start}))
    |> to(bucket: "wormscan-30days")