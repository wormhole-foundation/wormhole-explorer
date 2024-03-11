import "date"


calculateProtocolStats = (protocol,protocolVersion,taskCfg) => {

totalValueTransferred = from(bucket: taskCfg.sourceBucket)
                |> range(start: taskCfg.since, stop:taskCfg.ts)
                |> filter(fn: (r) => r._measurement == "vaa_volume_v2" and r.app_id == protocol)
                |> filter(fn: (r) => r._field == "volume" and r._value > 0)
                |> drop(columns:["destination_chain","emitter_chain","token_address","token_chain","version"])
                |> group()
                |> sum()
                |> map(fn: (r) => ({r with _time: time(v:taskCfg.since)}))
                |> set(key: "_field", value: "total_value_transferred")

totalMessages = from(bucket: taskCfg.sourceBucket)
                |> range(start: taskCfg.since, stop:taskCfg.ts)
                |> filter(fn: (r) => r._measurement == "vaa_volume_v2" and r.app_id == protocol)
                |> filter(fn: (r) => r._field == "volume")
                |> group()
                |> count()
                |> map(fn: (r) => ({r with _time: time(v:taskCfg.since)}))
                |> set(key: "_field", value: "total_messages")

return union(tables:[totalMessages,totalValueTransferred])
        |> set(key: "app_id", value: protocol)
        |> set(key: "version", value: protocolVersion)
        |> set(key: "_measurement", value: taskCfg.destMeasurement)
        |> map(fn: (r) => ({r with _time: time(v:taskCfg.since)}))
        |>to(bucket: taskCfg.destBucket)
}



ts = date.truncate(t: now(), unit: 1d)
bucketInfinite = "wormscan"
bucket30d = "wormscan-30days"

cfg = {
        sourceBucket:bucketInfinite,
        destBucket:bucketInfinite,
        destMeasurement:"core_protocols_stats_1d",
        since: date.sub(d: 1d, from: ts),
        ts:ts,
}

// Set this variable with the cfg of the desired task

option task = {
    name: "cctp and portal_token_bridge metrics every day",
    every: 1d,
}

calculateProtocolStats(protocol:"CCTP_WORMHOLE_INTEGRATION",protocolVersion:"v1",taskCfg:cfg)

calculateProtocolStats(protocol:"PORTAL_TOKEN_BRIDGE",protocolVersion:"v1",taskCfg:cfg)