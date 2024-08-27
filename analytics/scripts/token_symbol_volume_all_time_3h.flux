import "date"

option task = {
    name: "total volume all-time per token symbol",
    every: 3h,
}

bucketInfinite = "wormscan"
bucket24Hr = "wormscan-24hours"

srcBucket = bucketInfinite
destBucket = bucket24Hr
destMeasurement = "tokens_symbol_volume_all_time"
nowts = date.truncate(t: now(), unit: 1h)

from(bucket: srcBucket)
	|> range(start: 1970-01-01T00:00:00Z)
	|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
	|> filter(fn: (r) => r._field == "volume" or r._field == "symbol")
	|> keep(columns:["_time","_field","_value"])
	|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
	|> drop(columns:["_time"])
	|> filter(fn: (r) => r.volume > 0 and r.symbol != "")
	|> group(columns:["symbol"])
	|> sum(column:"volume")
	|> rename(columns: {volume: "_value"})
	|> set(key: "_field", value: "volume")
	|> set(key: "_measurement", value: destMeasurement)
	|> map(fn: (r) => ({ r with _time : nowts}))
	|> to(bucket: destBucket)