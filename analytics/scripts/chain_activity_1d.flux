import "date"

runTask = (start,stop,srcBucket,destBucket,destMeasurement) => {
    data = from(bucket: srcBucket)
  			|> range(start: start,stop: stop)
  			|> filter(fn: (r) => r._measurement == "vaa_volume_v2" and r._field == "volume")
			|> set(key: "_measurement", value: destMeasurement)
  			|> group(columns: ["emitter_chain", "destination_chain", "app_id"])
				
notional = data
		|> sum(column: "_value")
		|> rename(columns: {_field: "notional"})
							
txs = data
		|> count(column: "_value")
		|> rename(columns: {_field: "count"})

return join(tables: {t1: notional, t2: txs}, on: ["emitter_chain","destination_chain","app_id"])
	    |> set(key: "_time", value: string(v:start))
        |> to(bucket: destBucket)
}


bucketInfinite = "wormscan-mainnet-staging"
destMeasurement = "chain_activity_1d"
stop = date.truncate(t: now(),unit: 24h)
start = date.sub(d: 1d, from: stop)

option task = {
    name: "calculate chain activity every day",
    every: 1d,
}

runTask(start:start, stop: stop, srcBucket: bucketInfinite, destBucket: bucketInfinite, destMeasurement: destMeasurement)