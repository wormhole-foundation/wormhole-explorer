import "date"

option task = {
    name: "cctp and portal_token_bridge metrics every hour",
    every: 1h,
}


calculateLastHourMetrics = (protocol,protocolVersion,ts) => {

since = date.sub(d: 1h, from: ts)
sourceBucket = "wormscan"
destMeasurement = "core_protocols_stats_1h"
bucket30d = "wormscan-30days"

totalValueTransferred = from(bucket: sourceBucket)
                |> range(start: since, stop:ts)
                |> filter(fn: (r) => r._measurement == "vaa_volume_v2" and r.app_id == protocol)
                |> filter(fn: (r) => r._field == "volume" and r._value > 0)
                |> drop(columns:["destination_chain","emitter_chain","token_address","token_chain","version"])
                |> group()
                |> sum()
                |> map(fn: (r) => ({r with _time: since}))
                |> set(key: "_field", value: "total_value_transferred")

totalMessages = from(bucket: sourceBucket)
                |> range(start: since, stop:ts)
                |> filter(fn: (r) => r._measurement == "vaa_volume_v2" and r.app_id == protocol)
                |> filter(fn: (r) => r._field == "volume")
                |> group()
                |> count()
                |> map(fn: (r) => ({r with _time: since}))
                |> set(key: "_field", value: "total_messages")

return union(tables:[totalMessages,totalValueTransferred]) // if nothing happened during the last hour then union will result in empty and no point will be added.
        |> set(key: "app_id", value: protocol)
        |> set(key: "version", value: protocolVersion)
        |> set(key: "_measurement", value: destMeasurement)
        |> map(fn: (r) => ({r with _time: since}))
        |>to(bucket: bucket30d)
}

ts = date.truncate(t: now(), unit: 1h)

// execute function for CCTP_WORMHOLE_INTEGRATION
calculateLastHourMetrics(protocol:"CCTP_WORMHOLE_INTEGRATION",protocolVersion:"v1",ts:ts)


// execute function for PORTAL_TOKEN_BRIDGE
calculateLastHourMetrics(protocol:"PORTAL_TOKEN_BRIDGE",protocolVersion:"v1",ts:ts)