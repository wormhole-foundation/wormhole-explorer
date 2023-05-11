import "date"

start = date.sub(from: now(), d: 24h)
stop = now()

from(bucket: "wormscan-mainnet-staging")
    |> range(start: start)
    |> filter(fn: (r) => r["_measurement"] == "vaa_volume")
    |> filter(fn: (r) => r["_field"] == "volume")
    |> drop(columns: ["app_id", "destination_address", "token_address", "token_chain", "_field"])
    |> group(columns: ["emitter_chain", "destination_chain"])
    |> count(column: "_value")
    |> set(key: "_measurement", value: "chain_pair_transfers_24h")
    |> map(fn: (r) => ({r with _time: start}))
    |> to(bucket: "wormscan-30days")