package transactions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueries_createRangeQuery(t *testing.T) {
	var tests = []struct {
		tm                                       time.Time
		ts                                       string
		wantStartLastVaa, wantStartAggregatesVaa string
	}{
		{
			//2023-05-04T12:25:48.112233445Z
			tm:                     time.Date(2023, 5, 4, 12, 25, 48, 112233445, time.UTC),
			ts:                     "1d",
			wantStartLastVaa:       "2023-05-04T12:00:00Z",
			wantStartAggregatesVaa: "2023-05-03T12:00:00Z",
		},
		{
			//2023-05-04T20:59:17.992233445Z
			tm:                     time.Date(2023, 5, 4, 20, 59, 17, 992233445, time.UTC),
			ts:                     "1w",
			wantStartLastVaa:       "2023-05-04T00:00:00Z",
			wantStartAggregatesVaa: "2023-04-27T00:00:00Z",
		},
		{
			//2023-05-04T17:09:33.987654321Z
			tm:                     time.Date(2023, 5, 4, 17, 9, 33, 987654321, time.UTC),
			ts:                     "1mo",
			wantStartLastVaa:       "2023-05-04T00:00:00Z",
			wantStartAggregatesVaa: "2023-04-04T00:00:00Z",
		},
	}

	for _, tt := range tests {
		startLastVaa, startAggregatesVaa := createRangeQuery(tt.tm, tt.ts)
		assert.Equal(t, tt.wantStartLastVaa, startLastVaa)
		assert.Equal(t, tt.wantStartAggregatesVaa, startAggregatesVaa)
	}
}

func TestQueries_buildLastTrxQuery1d1h(t *testing.T) {

	expected := `
lastVaaCount = from(bucket: "wormscan-1month")
  |> range(start: 2023-05-04T18:00:00Z)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count")
  |> group()
  |> aggregateWindow(every: 1h, fn: count, createEmpty: true)
aggregatesVaaCount = from(bucket: "wormscan-1month")
  |> range(start: 2023-05-03T18:00:00Z)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count_1h")
union(tables: [aggregatesVaaCount, lastVaaCount])
  |> group()
  |> sort(columns: ["_time"], desc: true)
`
	//2023-05-04T18:39:10.985Z
	tm := time.Date(2023, 5, 4, 18, 39, 10, 985, time.UTC)
	actual := buildLastTrxQuery("wormscan-1month", tm, &TransactionCountQuery{TimeSpan: "1d", SampleRate: "1h"})
	assert.Equal(t, expected, actual)
}

func TestQueries_buildLastTrxQuery1w1d(t *testing.T) {

	expected := `
lastVaaCount = from(bucket: "wormscan-1month")
  |> range(start: 2023-05-04T00:00:00Z)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count")
  |> group()
aggregatesVaaCount = from(bucket: "wormscan-1month")
  |> range(start: 2023-04-27T00:00:00Z)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count_1h")
  |> aggregateWindow(every: 1h, fn: sum, createEmpty: true)
union(tables: [aggregatesVaaCount, lastVaaCount])
  |> group()
  |> aggregateWindow(every: 1d, fn: sum, createEmpty: true)
  |> sort(columns: ["_time"], desc: true)
`
	//2023-05-04T18:39:10.985Z
	tm := time.Date(2023, 5, 4, 18, 39, 10, 985, time.UTC)
	actual := buildLastTrxQuery("wormscan-1month", tm, &TransactionCountQuery{TimeSpan: "1w", SampleRate: "1d"})
	assert.Equal(t, expected, actual)
}

func TestQueries_buildTotalTrxCountQuery(t *testing.T) {

	expected := `
current = from(bucket: "bucket-forever")
  |> range(start: 2023-05-12T00:00:00Z)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume_v2")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group()
  |> count()
last = from(bucket: "bucket-30days")
  |> range(start: -1mo)
  |> filter(fn: (r) => r["_measurement"] == "total_tx_count_v2")
  |> last()
union(tables: [current, last])
  |> group()
  |> sum()
`
	//2023-05-04T18:39:10.985Z
	tm := time.Date(2023, 5, 12, 16, 53, 10, 985, time.UTC)
	actual := buildTotalTrxCountQuery("bucket-forever", "bucket-30days", tm)
	assert.Equal(t, expected, actual)
}

func TestQueries_buildTotalTrxVolumeQuery(t *testing.T) {

	expected := `
current = from(bucket: "bucket-forever")
  |> range(start: 2023-05-10T00:00:00Z)
  |> filter(fn: (r) => r["_measurement"] == "vaa_volume_v2")
  |> filter(fn: (r) => r["_field"] == "volume")
  |> group()
  |> sum()
last = from(bucket: "bucket-30days")
  |> range(start: -1mo)
  |> filter(fn: (r) => r["_measurement"] == "total_tx_volume_v2")
  |> last()
union(tables: [current, last])
  |> group()
  |> sum()
`
	//2023-05-04T18:39:10.985Z
	tm := time.Date(2023, 5, 10, 16, 53, 10, 985, time.UTC)
	actual := buildTotalTrxVolumeQuery("bucket-forever", "bucket-30days", tm)
	assert.Equal(t, expected, actual)
}
