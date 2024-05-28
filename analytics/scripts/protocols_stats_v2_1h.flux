import "date"
import "array"
import "join"
import "regexp"
import "types"
import "influxdata/influxdb/schema"

option task = {
    name: "calculate total_value_transferred and total_messages for all protocols every hour",
    every: 1h,
}

ts = date.truncate(t: now(), unit: 1h)
since = date.sub(d: 1h, from: ts)
bucketInfinite = "wormscan-24hours"
srcBucket = bucketInfinite
destBucket = bucketInfinite
destMeasurementTotals = "protocols_stats_totals_1h"
destMeasurementDeAggregated = "protocols_stats_1h"

allVaas = from(bucket: srcBucket)
		|> range(start: since,stop:ts)
		|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
		|> filter(fn: (r) => r._field == "volume")
		|> rename(columns:{"_value":"volume"})
		|> keep(columns:["_start","_stop","_time","volume","_field","app_id_1","app_id_2","app_id_3","emitter_chain","destination_chain"])
		|> group()

appIds1 = schema.tagValues(bucket: srcBucket, tag: "app_id_1")
appIds2 = schema.tagValues(bucket: srcBucket, tag: "app_id_2")
appIds3 = schema.tagValues(bucket: srcBucket, tag: "app_id_3")

allAppIDs = union(tables: [appIds1,appIds2,appIds3])
	|> filter(fn: (r) => r._value != "none")
	|> distinct()
	|> rename(columns:{"_value":"app_id"})

vaasAppID1 = join.inner(
    left: allVaas,
    right: allAppIDs,
    on: (l, r) => l.app_id_1 == r.app_id,
    as: (l, r) => ({
		"app_id":l.app_id_1,
		"emitter_chain":l.emitter_chain,
		"destination_chain":l.destination_chain,
		"volume":l.volume,
		"_time":l._start,
		}),
)

vaasAppID2 = join.inner(
    left: allVaas,
    right: allAppIDs,
    on: (l, r) => l.app_id_2 == r.app_id,
    as: (l, r) => ({
		"app_id":l.app_id_2,
		"emitter_chain":l.emitter_chain,
		"destination_chain":l.destination_chain,
		"volume":l.volume,
		"_time":l._start,
		}),
)

vaasAppID3 = join.inner(
    left: allVaas,
    right: allAppIDs,
    on: (l, r) => l.app_id_3 == r.app_id,
    as: (l, r) => ({
		"app_id":l.app_id_3,
		"emitter_chain":l.emitter_chain,
		"destination_chain":l.destination_chain,
		"volume":l.volume,
		"_time":l._start,
		}),
)

allTotals = union(tables: [vaasAppID1,vaasAppID2,vaasAppID3])
        |> rename(columns:{"volume":"_value"})
        |> set(key:"_field",value:"volume")
        |> group(columns:["app_id","emitter_chain","destination_chain","_time"])
        |> map(fn: (r) => ({r with app_id : string(v: "TOTAL_"+r.app_id)}))

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
		|> set(key: "_field", value: "total_messages")
		|> set(key: "_measurement", value: destMeasurementDeAggregated)
		|> to(bucket: destBucket)