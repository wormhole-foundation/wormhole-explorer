import "date"
import "array"
import "join"
import "regexp"
import "types"
import "influxdata/influxdb/schema"

// vars
ts = date.truncate(t: now(), unit: 1h)
since = date.sub(d: 1h, from: ts)
bucketInfinite = "wormscan-mainnet-staging"
srcBucket = bucketInfinite
destBucket = bucketInfinite
destMeasurement = "test_protocols_stats_1h_v3"


option task = {
    name: "calculate every hour the volume and txs for every combination of appIds and its totals",
    every: 1h,
}

filter_vaa_volume_v2 = (r) => {
    return r._measurement == "vaa_volume_v2"
}

filter_by_app_id = (appId) => {
	 return (r) => r.appID_1 == appId or r.appID_2 == appId or r.appID_3 == appId
}

getColumnValue = (tables=<-, field, appId) => {
    value = tables
        |> findColumn(fn: (key) => key._field == field, column: "_value")

	time = tables
        |> findColumn(fn: (key) => key._field == field, column: "_start")

    return {
		    "app_id": "TOTAL_" + appId,
			"_field": field,
			"_value": value[0],
			"_time": time[0],
			"_measurement": destMeasurement
		  }
}


// Calculate TOTALS for each of the appID_1 values.

appIds1 = schema.tagValues(bucket: "wormscan-mainnet-staging", tag: "appID_1")
appIds2 = schema.tagValues(bucket: "wormscan-mainnet-staging", tag: "appID_2")
appIds3 = schema.tagValues(bucket: "wormscan-mainnet-staging", tag: "appID_3")

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