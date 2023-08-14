package transactions

import (
	"fmt"
	"time"
)

// queryTemplateVaaCount1d1h is the query used to get the last VAA count and the aggregated VAA count for the last 24 hours by hour.
const queryTemplateVaaCount1d1h = `
lastVaaCount = from(bucket: "%s")
  |> range(start: %s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count")
  |> group()
  |> aggregateWindow(every: %s, fn: count, createEmpty: true)
aggregatesVaaCount = from(bucket: "%s")
  |> range(start: %s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count_1h")
union(tables: [aggregatesVaaCount, lastVaaCount])
  |> group()
  |> sort(columns: ["_time"], desc: true)
`

// queryTemplateVaaCount1d1h is the query used to get the last VAA count and the aggregated VAA count for 1 week or month by day.
const queryTemplateVaaCount = `
lastVaaCount = from(bucket: "%s")
  |> range(start: %s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count")
  |> group()
aggregatesVaaCount = from(bucket: "%s")
  |> range(start: %s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count_1h")
  |> aggregateWindow(every: 1h, fn: sum, createEmpty: true)
union(tables: [aggregatesVaaCount, lastVaaCount])
  |> group()
  |> aggregateWindow(every: 1d, fn: sum, createEmpty: true)
  |> sort(columns: ["_time"], desc: true)
`

func buildLastTrxQuery(
	dataPointsBucket string,
	aggregationsBucket string,
	tm time.Time,
	q *TransactionCountQuery,
) string {

	startLastVaa, startAggregatesVaa := createRangeQuery(tm, q.TimeSpan)
	if q.TimeSpan == "1d" && q.SampleRate == "1h" {
		return fmt.Sprintf(queryTemplateVaaCount1d1h, dataPointsBucket, startLastVaa, q.SampleRate, aggregationsBucket, startAggregatesVaa)
	}
	return fmt.Sprintf(queryTemplateVaaCount, dataPointsBucket, startLastVaa, aggregationsBucket, startAggregatesVaa)
}

func createRangeQuery(t time.Time, timeSpan string) (string, string) {

	const format = time.RFC3339Nano

	var startLastVaa, startAggregatesVaa time.Time

	switch timeSpan {
	case "1w":
		startLastVaa = t.Truncate(time.Hour * 24)
		startAggregatesVaa = startLastVaa.Add(-time.Hour * 24 * 7)
	case "1mo":
		startLastVaa = t.Truncate(time.Hour * 24)
		startAggregatesVaa = startLastVaa.Add(-time.Hour * 24 * 30)
	case "3mo":
		startLastVaa = t.Truncate(time.Hour * 24)
		startAggregatesVaa = startLastVaa.Add(-time.Hour * 24 * 90)
	default:
		startLastVaa = t.Truncate(time.Hour * 1)
		startAggregatesVaa = startLastVaa.Add(-time.Hour * 24)
	}

	return startLastVaa.Format(format), startAggregatesVaa.Format(format)
}

const queryTemplateTotalTrxCount = `
current = from(bucket: "%s")
  |> range(start: %s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group()
  |> count()
last = from(bucket: "%s")
  |> range(start: -1mo)
  |> filter(fn: (r) => r["_measurement"] == "total_tx_count")
  |> last()
union(tables: [current, last])
  |> group()
  |> sum()
`

func buildTotalTrxCountQuery(bucketForever, bucket30Days string, t time.Time) string {
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	return fmt.Sprintf(queryTemplateTotalTrxCount, bucketForever, start, bucket30Days)
}

const queryTemplateTotalTrxVolume = `
current = from(bucket: "%s")
  |> range(start: %s)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group()
  |> sum()
last = from(bucket: "%s")
  |> range(start: -1mo)
  |> filter(fn: (r) => r["_measurement"] == "total_tx_volume")
  |> last()
union(tables: [current, last])
  |> group()
  |> sum()
`

func buildTotalTrxVolumeQuery(bucketForever, bucket30Days string, t time.Time) string {
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	return fmt.Sprintf(queryTemplateTotalTrxVolume, bucketForever, start, bucket30Days)
}
