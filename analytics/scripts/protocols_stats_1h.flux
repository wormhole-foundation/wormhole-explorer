import "date"

sourceBucket = "wormscan-mainnet-staging"

ts = date.truncate(t: now(), unit: 1h)

data = from(bucket: sourceBucket)
  |> range(start: -3h, stop: ts)
  |> filter(fn: (r) => r._measurement == "vaa_volume_v2")
  |> filter(fn: (r) => r.app_id == "CCTP_WORMHOLE_INTEGRATION" or r.app_id == "PORTAL_TOKEN_BRIDGE")
  |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
  |> filter(fn: (r) => exists(r.notional) and exists(r.amount) and exists(r.volume))
  |> keep(columns: ["app_id", "version","notional", "amount", "volume", "_time"])
  |> map(fn: (r) => ({ r with _time: ts }))


totalVolume = data |> group(columns: ["app_id", "version"])
                   |> sum(column: "volume")
				           |> map(fn: (r) => ({ r with _time: ts }))
									 |> rename(columns: {"volume": "total_volume"})

totalMessages = data |> group(columns: ["app_id", "version"])
                    |> count(column: "volume")
                    |> map(fn: (r) => ({ r with _time: ts }))
								    |> rename(columns: {"volume": "total_messages"})

totalAmount = data |> group(columns: ["app_id", "version"])
                   |> sum(column: "amount")
				           |> map(fn: (r) => ({ r with _time: ts }))
									 |> rename(columns: {"amount": "total_amount"})


partialJoin = join(tables:{totalVolume,totalAmount},
	on: ["_time","app_id","version"]
)

join(tables:{partialJoin,totalMessages},
    on: ["_time","app_id","version"]
)
