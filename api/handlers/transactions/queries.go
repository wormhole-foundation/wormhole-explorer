package transactions

import (
	"fmt"
	"time"
)

const queryTemplateVaaCount = `
lastVaaCount = from(bucket: "%s")
  |> range(start: %s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count")
  |> group()
  |> aggregateWindow(every: %s, fn: count, createEmpty: true)

aggregatesVaaCount = from(bucket: "%s")
  |> range(start: %s , stop: %s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count_1h")
  |> aggregateWindow(every: %s, fn: count, createEmpty: true)

union(tables: [aggregatesVaaCount, lastVaaCount])
  |> group()
  |> sort(columns: ["_time"], desc: true)
`

func buildLastTrxQuery(bucket string, tm time.Time, q *TransactionCountQuery) string {
	startLastVaa, startAggregatesVaa, stopAggregatesVaa := createRangeQuery(tm, q.TimeSpan)
	return fmt.Sprintf(queryTemplateVaaCount, bucket, startLastVaa, q.SampleRate, bucket, startAggregatesVaa, stopAggregatesVaa, q.SampleRate)
}

func createRangeQuery(t time.Time, timeSpan string) (string, string, string) {

	const format = time.RFC3339Nano

	startLastVaa := t.Truncate(time.Hour * 1)
	stopAggregatesVaa := startLastVaa.Add(time.Nanosecond * 1)
	var startAggregatesVaa time.Time

	switch timeSpan {
	case "1w":
		startAggregatesVaa = startLastVaa.Add(-time.Hour * 24 * 7)
	case "1mo":
		startAggregatesVaa = startLastVaa.Add(-time.Hour * 24 * 30)
	default:
		startAggregatesVaa = startLastVaa.Add(-time.Hour * 24)
	}

	return startLastVaa.Format(format), startAggregatesVaa.Format(format), stopAggregatesVaa.Format(format)
}
