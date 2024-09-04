import "date"

option task = {
    name: "cctp total volume all-time every hour",
    every: 1h,
}

bucketInfinite = "wormscan"
bucket24Hr = "wormscan-24hours"
destMeasurement = "cctp_status_total"
nowts = date.truncate(t: now(), unit: 1h)

lastData = from(bucket: bucket24Hr)
    	|> range(start: -1d)
    	|> filter(fn: (r) => r._measurement == destMeasurement)
    	|> last()
    	|> drop(columns:["_start","_stop"])

lastTxs = lastData
            |> filter(fn: (r) => r._field == "txs")

lastVolume = lastData
                |> filter(fn: (r) => r._field == "volume")

lastExecutionTime = lastData
                        |> keep(columns:["_time"])
                        |> tableFind(fn: (key) => true)
                        |> getRecord(idx: 0)

deltaData = from(bucket: bucketInfinite)
    		    |> range(start: lastExecutionTime._time, stop:now())
    		    |> filter(fn: (r) => r._measurement == "circle-message-sent")
    		    |> filter(fn: (r) => r._field == "amount")
                |> keep(columns:["_field","_value"])
                |> toUInt()
                |> reduce(
                    identity: {
                            volume: uint(v:0),
                            txs: uint(v:0)
                    },
                    fn: (r, accumulator) => ({
                            volume: accumulator.volume + r._value,
                            txs: accumulator.txs + uint(v:1)
                        })
                    )
deltaTxs = deltaData
				|> drop(columns:["volume"])
				|> rename(columns: {txs: "_value"})
				|> set(key:"_field",value:"txs")

deltaVolume = deltaData
				|> drop(columns:["txs"])
				|> rename(columns: {volume: "_value"})
				|> set(key:"_field",value:"volume")

txs = union(tables:[lastTxs, deltaTxs])
				|> sum()

volume = union(tables:[lastVolume, deltaVolume])
				|> sum()

union(tables:[txs, volume])
    |> set(key:"_measurement", value: destMeasurement)
    |> map(fn: (r) => ({ r with _time: nowts }))
    |> to(bucket: bucket24Hr)
