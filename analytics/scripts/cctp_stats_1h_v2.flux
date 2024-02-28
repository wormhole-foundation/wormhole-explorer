import "date"
import "types"
import "array"

ts = date.truncate(t: now(), unit: 1h)

latestValue = from(bucket: "wormscan-24hours-mainnet-staging")
                |> range(start: 1970-01-01T00:00:00Z)
                |> filter(fn: (r) => r._measurement == "msosto_test_cctp_tb_stats")
	            	|> filter(fn: (r) => r.app_id == "CCTP_WORMHOLE_INTEGRATION" and r.version == "v2")
	            	|> last()
								|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
								|> map(fn: (r) => ({
 											r with total_messages: uint(v: r.total_messages),
								}))

latestValueRecord = latestValue
                						|> findRecord(
                						      fn: (key) => key._measurement == "msosto_test_cctp_tb_stats",
																	idx: 0,
                						)

deltaData = from(bucket: "wormscan-mainnet-staging")
  					|> range(start: latestValueRecord._time,stop:ts)
  					|> filter(fn: (r) => r._measurement == "vaa_volume_v2")
						|> filter(fn: (r) => r.app_id == "CCTP_WORMHOLE_INTEGRATION")
  					|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
  					|> filter(fn: (r) => exists(r.notional) and exists(r.amount) and exists(r.volume))
  					|> keep(columns: ["app_id", "version", "amount", "volume", "_time"])
  					|> map(fn: (r) => ({ r with _time: ts }))



defaultData = array.from(rows: [
{
app_id:"CCTP_WORMHOLE_INTEGRATION",
total_messages: 0,
total_value_transferred:0.0,
_time: ts
}
])


dd2 = deltaData |> group(columns: ["app_id", "version","_time"])
				         	|> reduce(
        		        		   fn: (r, accumulator) => ({
            	        			volume: accumulator.volume + float(v:r.volume),
				                    count: accumulator.count + 1,
            	    }),
		                			identity: {volume: float(v: 0), count: int(v: 0)}
    			        			)
				        	|> rename(columns: {"volume": "total_value_transferred"})
				        	|> rename(columns: {"count": "total_messages"})
									|> map(fn: (r) => ({
 											r with total_messages: uint(v: r.total_messages),
									}))
									|> drop(columns: ["version"])


lst = latestValue |> map(fn: (r) => ({ r with _time: ts }))
                  |> keep(columns: ["app_id","total_messages","total_value_transferred","_time"])

partialResult = union(tables: [dd2,lst])
|> group(columns: ["app_id","_time"])
|> reduce(
  		   fn: (r, accumulator) => ({
    			tvt_acc: accumulator.tvt_acc + float(v:r.total_value_transferred),
	        total_msg_acc: accumulator.total_msg_acc + uint(v:r.total_messages),
     	    }),
     			identity: {tvt_acc: float(v: 0), total_msg_acc: uint(v: 0)}
)
|> rename(columns: {"tvt_acc": "total_value_transferred"})
|> rename(columns: {"total_msg_acc": "total_messages"})


vaaWithoutPrice = from(bucket: "wormscan-mainnet-staging")
    |> range(start: latestValueRecord._time, stop: ts)
    |> filter(fn: (r) => r._measurement == "vaa_volume_v2")
    |> filter(fn: (r) => r.app_id == "CCTP_WORMHOLE_INTEGRATION")
    |> filter(fn: (r) => r._field == "volume" and r._value == 0)
    |> drop(columns: ["destination_chain","emitter_chain","token_address","token_chain","_time"])
    |> group(columns: ["app_id","version"])
    |> count()
    |> rename(columns: {"_value": "total_messages_withoutvaa"})
	  |> map(fn: (r) => ({ r with _time: ts }))
		|> keep(columns : ["_time","total_messages_withoutvaa","app_id"])




union(tables: [partialResult,vaaWithoutPrice])
	|> map(fn: (r) => ({
      r with
			 total_messages: uint(v: r.total_messages) + (if exists r.total_messages_withoutvaa then uint(v: r.total_messages_withoutvaa) else uint(v: 0))
 }))
 |> keep(columns: ["app_id","total_value_transferred","total_messages","version"])
 |> map(fn: (r) => ({ r with _time: ts }))
 |> map(fn: (r) => ({ r with _measurement: "msosto_test_cctp_tb_stats" }))

