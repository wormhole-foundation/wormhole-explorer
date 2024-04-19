import "date"


runTask = (start,stop,srcBucket,destBucket,destMeasurement) => {

    data = from(bucket: srcBucket)
                 	|> range(start: start,stop: stop)
                 	|> filter(fn: (r) => r._measurement == "vaa_volume_v2" and r.version == "v2")
                    |> filter(fn: (r) => r._field == "volume" and r._value > 0)
               		|> drop(columns:["destination_chain","app_id","token_chain","token_address","version","_measurement","_time"])
    				|> rename(columns: {_start: "_time"})
                 	|> group(columns: ["emitter_chain","_time"])

    vols =	data
    			|> sum(column: "_value")
    			|> set(key: "_field", value: "volume")
    			|> set(key: "to", value: string(v:stop))
    			|> set(key: "_measurement", value: destMeasurement)
    			|> to(bucket: destBucket)

    return data
    		    |> count(column: "_value")
    		    |> set(key: "_field", value: "count")
    		    |> set(key: "to", value: string(v:stop))
    		    |> set(key: "_measurement", value: destMeasurement)
    		    |> to(bucket: destBucket)
}

bucketInfinite = "wormscan"
destMeasurement = "emitter_chain_activity_1d"

stop = date.truncate(t: now(),unit: 1d)
start = date.sub(d: 1d, from: stop)

option task = {
    name: "calculate chain activity per emitter every day",
    every: 1d,
}

runTask(start:start, stop: stop, srcBucket: bucketInfinite, destBucket: bucketInfinite, destMeasurement: destMeasurement)