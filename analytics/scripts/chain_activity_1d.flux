import "date"

runTask = (start,stop,srcBucket,destBucket,destMeasurement) => {
        data = from(bucket: srcBucket)
      		    |> range(start: start,stop: stop)
      			|> filter(fn: (r) => r._measurement == "vaa_volume_v2" and r._field == "volume")
      			|> group(columns: ["emitter_chain", "destination_chain", "app_id"])

        data
    		|> sum(column: "_value")
    		|> set(key: "_field", value: "volume")
    		|> map(fn: (r) => ({ r with _time: start }))
    		|> set(key: "to", value: string(v:date.add(d: 1h, to: start)))
    		|> set(key: "_measurement", value: destMeasurement)
    		|> to(bucket: destBucket)

        return data
    		        |> count(column: "_value")
    		        |> set(key: "_field", value: "count")
    		        |> map(fn: (r) => ({ r with _time: start }))
    		        |> set(key: "to", value: string(v:date.add(d: 1h, to: start)))
    		        |> set(key: "_measurement", value: destMeasurement)
    		        |> to(bucket: destBucket)
}


bucketInfinite = "wormscan"
destMeasurement = "chain_activity_1d"
stop = date.truncate(t: now(),unit: 24h)
start = date.sub(d: 1d, from: stop)

option task = {
    name: "calculate chain activity every day",
    every: 1d,
}

runTask(start:start, stop: stop, srcBucket: bucketInfinite, destBucket: bucketInfinite, destMeasurement: destMeasurement)