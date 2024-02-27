import "date"

sourceBucket = "wormscan-mainnet-staging"
destinationBucket = "wormscan-24hours-mainnet-staging"
destinationMeasurement = "msosto_test_cctp_tb_stats"

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
				   |> rename(columns: {"volume": "total_value_transferred"})

totalMessages = data |> group(columns: ["app_id", "version"])
                     |> count(column: "volume")
                     |> map(fn: (r) => ({ r with _time: ts }))
				     |> rename(columns: {"volume": "total_messages"})

almost = join(tables:{totalVolume,totalMessages},
	on: ["_time","app_id","version"]
)
|> map(fn: (r) => ({ r with version: "v1" }))
|> map(fn: (r) => ({ r with _measurement: destinationMeasurement }))

totalMsgPoints = almost
						|> rename(columns: {"total_messages": "_value"})
			 			|> set(key: "_field", value: "total_messages")
						|> map(fn: (r) => ({ r with _value: float(v: r._value) }))
						|> drop(columns: ["total_value_transferred"])


tvtPoints = almost
					|> rename(columns: {"total_value_transferred": "_value"})
			 		|> set(key: "_field", value: "total_value_transferred")
					|> map(fn: (r) => ({ r with _value: float(v: r._value) }))
					|> drop(columns: ["total_messages"])

union(tables: [totalMsgPoints,tvtPoints])
		 |> to(bucket: destinationBucket)

