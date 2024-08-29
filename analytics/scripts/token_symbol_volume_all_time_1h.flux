import "date"

option task = {
    name: "total volume all-time per token symbol every hour",
    every: 1h,
}

bucketInfinite = "wormscan"
bucket24Hr = "wormscan-24hours"
destMeasurement = "tokens_symbol_volume_all_time"
nowts = date.truncate(t: now(), unit: 1h)

lastData = from(bucket: bucket24Hr)
    	|> range(start: -1d)
    	|> filter(fn: (r) => r._measurement == destMeasurement)
    	|> last()
    	|> drop(columns:["_start","_stop"])

lastExecutionTime = lastData
    				|> keep(columns:["_time"])
    				|> tableFind(fn: (key) => true)
    				|> getRecord(idx: 0)

deltaSinceLastExecution = from(bucket: bucketInfinite)
    		|> range(start: lastExecutionTime._time, stop:now())
    		|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
    		|> filter(fn: (r) => r._field == "volume" or r._field == "symbol")
    		|> keep(columns:["_time","_field","_value"])
    		|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
    		|> drop(columns:["_time"])
    		|> filter(fn: (r) => r.volume > 0 and r.symbol != "")
    		|> group(columns:["symbol"])
    		|> sum(column:"volume")
    		|> rename(columns: {volume: "_value"})

union(tables:[lastData,deltaSinceLastExecution])
    |> group(columns:["symbol"])
    |> sum()
    |> set(key: "_field", value: "volume")
    |> set(key: "_measurement", value: destMeasurement)
    |> map(fn: (r) => ({
    		r with
    		_time: nowts
    }))
    |> to(bucket: bucket24Hr)