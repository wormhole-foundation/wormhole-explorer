package transactions

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueries_createRangeQuery(t *testing.T) {
	var tests = []struct {
		tm                                                              time.Time
		ts                                                              string
		wantStartLastVaa, wantStartAggregatesVaa, wantStopAggregatesVaa string
	}{
		{
			//2023-05-04T12:25:48.112233445Z
			tm:                     time.Date(2023, 5, 4, 12, 25, 48, 112233445, time.UTC),
			ts:                     "1d",
			wantStartLastVaa:       "2023-05-04T12:00:00Z",
			wantStartAggregatesVaa: "2023-05-03T12:00:00Z",
			wantStopAggregatesVaa:  "2023-05-04T12:00:00.000000001Z",
		},
		{
			//2023-05-04T20:59:17.992233445Z
			tm:                     time.Date(2023, 5, 4, 20, 59, 17, 992233445, time.UTC),
			ts:                     "1w",
			wantStartLastVaa:       "2023-05-04T20:00:00Z",
			wantStartAggregatesVaa: "2023-04-27T20:00:00Z",
			wantStopAggregatesVaa:  "2023-05-04T20:00:00.000000001Z",
		},
		{
			//2023-05-04T17:09:33.987654321Z
			tm:                     time.Date(2023, 5, 4, 17, 9, 33, 987654321, time.UTC),
			ts:                     "1mo",
			wantStartLastVaa:       "2023-05-04T17:00:00Z",
			wantStartAggregatesVaa: "2023-04-04T17:00:00Z",
			wantStopAggregatesVaa:  "2023-05-04T17:00:00.000000001Z",
		},
	}

	for _, tt := range tests {
		startLastVaa, startAggregatesVaa, stopAggregatesVaa := createRangeQuery(tt.tm, tt.ts)
		assert.Equal(t, tt.wantStartLastVaa, startLastVaa)
		assert.Equal(t, tt.wantStartAggregatesVaa, startAggregatesVaa)
		assert.Equal(t, tt.wantStopAggregatesVaa, stopAggregatesVaa)
	}
}

func TestQueries_buildLastTrxQuery(t *testing.T) {

	expected := `
lastVaaCount = from(bucket: "wormscan-1month")
  |> range(start: 2023-05-04T18:00:00Z)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count")
  |> group()
  |> aggregateWindow(every: 1h, fn: count, createEmpty: true)

aggregatesVaaCount = from(bucket: "wormscan-1month")
  |> range(start: 2023-05-03T18:00:00Z , stop: 2023-05-04T18:00:00.000000001Z)
  |> filter(fn: (r) => r["_measurement"] == "vaa_count_1h")
  |> aggregateWindow(every: 1h, fn: count, createEmpty: true)

union(tables: [aggregatesVaaCount, lastVaaCount])
  |> group()
  |> sort(columns: ["_time"], desc: true)
`
	//2023-05-04T18:39:10.985Z
	tm := time.Date(2023, 5, 4, 18, 39, 10, 985, time.UTC)
	actual := buildLastTrxQuery("wormscan-1month", tm, &TransactionCountQuery{TimeSpan: "1d", SampleRate: "1h"})
	fmt.Println(actual)
	fmt.Println(actual)
	assert.Equal(t, expected, actual)
}
