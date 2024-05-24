import "date"
import "array"
import "join"
import "regexp"
import "types"
import "influxdata/influxdb/schema"

ts = date.truncate(t: now(), unit: 1h)
since = date.sub(d: 1h, from: ts)
bucketInfinite = "wormscan-mainnet-staging"
srcBucket = bucketInfinite
destBucket = bucketInfinite
destMeasurement = "protocols_stats_1h"


filter_vaa_volume_v3 = (r) => {
    return r._measurement == "vaa_volume_v3" and r.version == "v5"
}

filter_by_app_id = (appId) => {
	 return (r) => r.app_id_1 == appId or r.app_id_2 == appId or r.app_id_3 == appId
}

filter_by_field = (field) => {
	 return (r) => r._field == field
}

getColumnValue = (tables=<-, field, appId) => {
  value = tables
        |> findColumn(fn: (key) => key._field == field, column: "_value")

	time = tables
        |> findColumn(fn: (key) => key._field == field, column: "_start")

	destinationChain = tables
        |> findColumn(fn: (key) => key._field == field, column: "destination_chain")

	emitterChain = tables
        |> findColumn(fn: (key) => key._field == field, column: "emitter_chain")

    return {
		  "app_id": "TOTAL_" + appId,
			"_field": field,
			"_value": value[0],
			"_time": time[0],
			"_measurement": destMeasurement,
			"destination_chain":destinationChain[0],
			"emitter_chain":emitterChain[0],
		  }
}

allVaas = from(bucket: srcBucket)
		|> range(start: 2024-05-24T13:00:00Z,stop:2024-05-24T14:00:00Z)
		|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
		|> filter(fn: (r) => r._field == "volume")
		|> rename(columns:{"_value":"volume"})
		|> keep(columns:["_start","_stop","_time","volume","_field","app_id_1","app_id_2","app_id_3","emitter_chain","destination_chain"])
		|> group()

appIds1 = schema.tagValues(bucket: "wormscan-mainnet-staging", tag: "app_id_1")
appIds2 = schema.tagValues(bucket: "wormscan-mainnet-staging", tag: "app_id_2")
appIds3 = schema.tagValues(bucket: "wormscan-mainnet-staging", tag: "app_id_3")

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
		"volume":l.volume
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
		"volume":l.volume
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
	                    	"volume":l.volume
	                    	}),
            )

allTotals = union(tables: [vaasAppID1,vaasAppID2,vaasAppID3])
                    |> rename(columns:{"volume":"_value"})
                    |> set(key:"_field",value:"volume")
                    |> group(columns:["app_id","emitter_chain","destination_chain","_start","_stop"])
                    |> map(fn: (r) => ({r with app_id : string(v: "TOTAL_"+r.app_id)}))

totalsVols = allTotals
		            |> sum()
		            |> set(key:"_field",value:"total_value_transferred")
		            |> set(key: "_measurement", value: destMeasurement)
		            |> to(bucket: destBucket)

totalsCounts = allTotals
		            |> count()
		            |> set(key:"_field",value:"total_messages")
		            |> set(key: "_measurement", value: destMeasurement)
		            |> to(bucket: destBucket)

// Calculate TOTALS for each of the appID_1 values.

appIds1 = schema.tagValues(bucket: "wormscan-mainnet-staging", tag: "app_id_1")
appIds2 = schema.tagValues(bucket: "wormscan-mainnet-staging", tag: "app_id_2")
appIds3 = schema.tagValues(bucket: "wormscan-mainnet-staging", tag: "app_id_3")

allAppIDs = union(tables: [appIds1,appIds2,appIds3])
	|> filter(fn: (r) => r._value != "none")
	|> distinct()

allAppIDs
		|> map(fn: (r) => {
			return from(bucket: srcBucket)
							|> range(start: since,stop: ts)
							|> filter(fn: filter_vaa_volume_v2)
							|> filter(fn: filter_by_app_id(appId: r._value))
							|> drop(columns:["app_id_1","app_id_2","app_id_3","size"])
							|> sum()
							|> first()
							|> set(key: "_field", value: "total_value_transferred")
							|> getColumnValue(field:"total_value_transferred",appId:r._value)
			})
		|> to(bucket: destBucket)

allAppIDs
		|> map(fn: (r) => {
			return from(bucket: srcBucket)
							|> range(start: since,stop: ts)
							|> filter(fn: filter_vaa_volume_v2)
							|> filter(fn: filter_by_app_id(appId: r._value))
							|> drop(columns:["app_id_1","app_id_2","app_id_3","size"])
							|> count()
							|> first()
							|> set(key: "_field", value: "total_messages")
							|> getColumnValue(field:"total_messages",appId:r._value)
			})
		|> to(bucket: destBucket)


// Calculate deAggregated values

allData = from(bucket: srcBucket)
		|> range(start: since,stop: ts)
		|> filter(fn: (r) => r._measurement == "vaa_volume_v2")
		|> drop(columns:["size"])
		|> group(columns:["appID_1","appID_2","appID_3","_start"])

allData
		|> sum()
		|> set(key: "_field", value: "total_value_transferred")
		|> set(key: "_measurement", value: destMeasurement)
		|> rename(columns: {_start: "_time"})
		|> to(bucket: destBucket)

allData
		|> count()
		|> set(key: "_field", value: "total_messages")
		|> set(key: "_measurement", value: destMeasurement)
		|> rename(columns: {_start: "_time"})
		|> to(bucket: destBucket)