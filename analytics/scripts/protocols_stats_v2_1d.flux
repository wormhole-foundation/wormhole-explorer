import "date"
import "array"
import "join"
import "regexp"
import "types"
import "influxdata/influxdb/schema"

option task = {
    name: "calculate total_value_transferred and total_messages for all protocols every day",
    every: 1d,
}

bucketInfinite = "wormscan"
srcBucket = bucketInfinite
destBucket = bucketInfinite
destMeasurementTotals = "protocols_stats_totals_1d"
destMeasurementDeAggregated = "protocols_stats_1d"
ts = date.truncate(t: now(), unit: 1d)
since = date.sub(d: 1d, from: ts)


allByAppId1 = from(bucket: srcBucket)
        |> range(start: since, stop: ts)
        |> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
        |> filter(fn: (r) => r._field == "volume")
		|> drop(columns:["app_id_2","app_id_3","token_chain","token_address","size","version"])
		|> filter(fn: (r)=> r.app_id_1 != "none")
		|> group(columns:["app_id_1","destination_chain","emitter_chain"])
		|> rename(columns:{"app_id_1":"app_id"})

allByAppId2 = from(bucket: srcBucket)
        |> range(start: -1h)
        |> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
        |> filter(fn: (r) => r._field == "volume")
		|> drop(columns:["app_id_1","app_id_3","token_chain","token_address","size","version"])
		|> filter(fn: (r) => r.app_id_2 != "none")
		|> group(columns:["app_id_1","destination_chain","emitter_chain"])
		|> rename(columns:{"app_id_2":"app_id"})


allByAppId3 = from(bucket: srcBucket)
        |> range(start: -1h)
        |> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
        |> filter(fn: (r) => r._field == "volume")
		|> drop(columns:["app_id_1","app_id_1","token_chain","token_address","size","version"])
		|> filter(fn: (r)=> r.app_id_3 != "none")
		|> group(columns:["app_id_1","destination_chain","emitter_chain"])
		|> rename(columns:{"app_id_3":"app_id"})

allTotals = union(tables: [allByAppId1,allByAppId2,allByAppId3])
        |> set(key:"_field",value:"volume")
        |> group(columns:["app_id","emitter_chain","destination_chain","_time"])
        |> map(fn: (r) => ({
            "app_id": "TOTAL_" + r.app_id,
            "_value": r._value,
            "emitter_chain": r.emitter_chain,
            "destination_chain": r.destination_chain,
            "_time": date.truncate(t: r._time, unit: 1h)
        }))

allTotals
        |> sum()
        |> set(key:"_field",value:"total_value_transferred")
        |> set(key: "_measurement", value: destMeasurementTotals)
        |> to(bucket: destBucket)

allTotals
        |> count()
        |> set(key:"_field",value:"total_messages")
        |> set(key: "_measurement", value: destMeasurementTotals)
        |> to(bucket: destBucket)


// Calculate deAggregated values

allData = from(bucket: srcBucket)
		|> range(start: since,stop: ts)
		|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
		|> filter(fn: (r) => r._field == "volume")
		|> drop(columns:["size","_time"])
		|> rename(columns: {_start: "_time"})
		|> group(columns:["_time","app_id_1","app_id_2","app_id_3","emitter_chain","destination_chain"])

allData
		|> sum()
		|> set(key: "_field", value: "total_value_transferred")
		|> set(key: "_measurement", value: destMeasurementDeAggregated)
		|> to(bucket: destBucket)

allData
		|> count()
		|> map(fn: (r) => ({r with _value: uint(v: r._value)}))
		|> set(key: "_field", value: "total_messages")
		|> set(key: "_measurement", value: destMeasurementDeAggregated)
		|> to(bucket: destBucket)