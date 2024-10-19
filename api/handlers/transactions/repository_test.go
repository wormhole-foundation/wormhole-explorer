package transactions

import (
	"bytes"
	"context"
	"errors"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_convertToDecimal(t *testing.T) {

	tcs := []struct {
		input  uint64
		output string
	}{
		{
			input:  1,
			output: "0.00000001",
		},
		{
			input:  1000_0000,
			output: "0.10000000",
		},
		{
			input:  1_0000_0000,
			output: "1.00000000",
		},
		{
			input:  1234_5678_1234,
			output: "1234.56781234",
		},
	}

	for i := range tcs {
		tc := tcs[i]

		result := convertToDecimal(tc.input)
		if result != tc.output {
			t.Errorf("expected %s, got %s", tc.output, result)
		}
	}

}

func Test_buildChainActivityQueryTops(t *testing.T) {

	repository := &Repository{
		bucketInfiniteRetention: "wormscan-testenv",
	}

	tcs := []struct {
		name     string
		input    ChainActivityTopsQuery
		expected string
	}{
		{
			name: "Search only by time range hourly",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{},
				TargetChains: []sdk.ChainID{},
				From:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Hour,
			},
			expected: `
					import "date"

					from(bucket: "wormscan-testenv")
					|> range(start: 2024-01-01T00:00:00Z,stop: 2024-01-03T05:00:00Z)
					|> filter(fn: (r) => r._measurement == "emitter_chain_activity_1h")
					
					|> pivot(rowKey:["_time","emitter_chain"], columnKey: ["_field"], valueColumn: "_value")
					|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search only by time range daily",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{},
				TargetChains: []sdk.ChainID{},
				From:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Day,
			},
			expected: `
					import "date"

					from(bucket: "wormscan-testenv")
					|> range(start: 2024-01-01T00:00:00Z,stop: 2024-01-03T00:00:00Z)
					|> filter(fn: (r) => r._measurement == "emitter_chain_activity_1d")
					
					|> pivot(rowKey:["_time","emitter_chain"], columnKey: ["_field"], valueColumn: "_value")
					|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search only by time range monthly",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{},
				TargetChains: []sdk.ChainID{},
				From:         time.Date(2023, 10, 7, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Month,
			},
			expected: `
				import "date"
				import "join"

				data = from(bucket: "wormscan-testenv")
						|> range(start: 2023-10-01T00:00:00Z,stop: 2024-01-01T00:00:00Z)
						|> filter(fn: (r) => r._measurement == "emitter_chain_activity_1d")
						
						|> drop(columns:["to"])
						|> window(every: 1mo, period:1mo)
						|> drop(columns:["_time"])
						|> rename(columns: {_start: "_time"})
						|> map(fn: (r) => ({r with to: string(v: r._stop)}))

				vols = data
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

				counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

				join.inner(
						left: vols,
						right: counts,
						on: (l, r) => l._time == r._time and l.emitter_chain == r.emitter_chain,
						as: (l, r) => ({l with count: r.count}),
				)
				|> group()
				|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search only by time range yearly",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{},
				TargetChains: []sdk.ChainID{},
				From:         time.Date(2020, 10, 7, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Year,
			},
			expected: `
				import "date"
				import "join"

				data = from(bucket: "wormscan-testenv")
						|> range(start: 2020-01-01T00:00:00Z,stop: 2024-01-01T00:00:00Z)
						|> filter(fn: (r) => r._measurement == "emitter_chain_activity_1d")
						
						|> drop(columns:["to"])
						|> window(every: 1y, period:1y)
						|> drop(columns:["_time"])
						|> rename(columns: {_start: "_time"})
						|> map(fn: (r) => ({r with to: string(v: r._stop)}))

				vols = data
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

				counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

				join.inner(
						left: vols,
						right: counts,
						on: (l, r) => l._time == r._time and l.emitter_chain == r.emitter_chain,
						as: (l, r) => ({l with count: r.count}),
				)
				|> group()
				|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search by emitter_chain daily",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{1},
				TargetChains: []sdk.ChainID{},
				From:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Day,
			},
			expected: `
					import "date"

					from(bucket: "wormscan-testenv")
					|> range(start: 2024-01-01T00:00:00Z,stop: 2024-01-03T00:00:00Z)
					|> filter(fn: (r) => r._measurement == "emitter_chain_activity_1d")
					|> filter(fn: (r) => r.emitter_chain == "1")
					|> pivot(rowKey:["_time","emitter_chain"], columnKey: ["_field"], valueColumn: "_value")
					|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search by multiple emitter_chain daily",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{1, 2},
				TargetChains: []sdk.ChainID{},
				From:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Day,
			},
			expected: `
					import "date"

					from(bucket: "wormscan-testenv")
					|> range(start: 2024-01-01T00:00:00Z,stop: 2024-01-03T00:00:00Z)
					|> filter(fn: (r) => r._measurement == "emitter_chain_activity_1d")
					|> filter(fn: (r) => r.emitter_chain == "1" or r.emitter_chain == "2")
					|> pivot(rowKey:["_time","emitter_chain"], columnKey: ["_field"], valueColumn: "_value")
					|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search by emitter_chain and target_chain hourly",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{1},
				TargetChains: []sdk.ChainID{2},
				From:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Hour,
			},
			expected: `
					import "date"
					import "join"

					data = from(bucket: "wormscan-testenv")
		  			|> range(start: 2024-01-01T00:00:00Z,stop: 2024-01-03T05:00:00Z)
		  			|> filter(fn: (r) => r._measurement == "chain_activity_1h")
					|> filter(fn: (r) => r.emitter_chain == "1")
					|> filter(fn: (r) => r.destination_chain == "2")
					
					|> drop(columns:["destination_chain"])

					vols = data		
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

					counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

					join.inner(
					    left: vols,
					    right: counts,
					    on: (l, r) => l._time == r._time and l.to == r.to and l.emitter_chain == r.emitter_chain,
					    as: (l, r) => ({l with count: r.count}),
					)
					|> group()
					|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search by multiple emitter_chain and multiple target_chain daily",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{1, 3},
				TargetChains: []sdk.ChainID{2, 4},
				From:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Day,
			},
			expected: `
					import "date"
					import "join"

					data = from(bucket: "wormscan-testenv")
		  			|> range(start: 2024-01-01T00:00:00Z,stop: 2024-01-03T00:00:00Z)
		  			|> filter(fn: (r) => r._measurement == "chain_activity_1d")
					|> filter(fn: (r) => r.emitter_chain == "1" or r.emitter_chain == "3")
					|> filter(fn: (r) => r.destination_chain == "2" or r.destination_chain == "4")
					
					|> drop(columns:["destination_chain"])

					vols = data		
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

					counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

					join.inner(
					    left: vols,
					    right: counts,
					    on: (l, r) => l._time == r._time and l.to == r.to and l.emitter_chain == r.emitter_chain,
					    as: (l, r) => ({l with count: r.count}),
					)
					|> group()
					|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search by app_id daily",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{},
				TargetChains: []sdk.ChainID{},
				From:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Day,
				AppId:        "CCTP_WORMHOLE_INTEGRATION",
			},
			expected: `
					import "date"
					import "join"

					data = from(bucket: "wormscan-testenv")
		  			|> range(start: 2024-01-01T00:00:00Z,stop: 2024-01-03T00:00:00Z)
		  			|> filter(fn: (r) => r._measurement == "chain_activity_1d")
					
					
					|> filter(fn: (r) => r.app_id == "CCTP_WORMHOLE_INTEGRATION")
					|> drop(columns:["destination_chain"])

					vols = data		
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

					counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

					join.inner(
					    left: vols,
					    right: counts,
					    on: (l, r) => l._time == r._time and l.to == r.to and l.emitter_chain == r.emitter_chain,
					    as: (l, r) => ({l with count: r.count}),
					)
					|> group()
					|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search by multiple emitter_chain, destination_chain and app_id daily",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{1, 2},
				TargetChains: []sdk.ChainID{3, 4},
				From:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 1, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Day,
				AppId:        "CCTP_WORMHOLE_INTEGRATION",
			},
			expected: `
					import "date"
					import "join"

					data = from(bucket: "wormscan-testenv")
		  			|> range(start: 2024-01-01T00:00:00Z,stop: 2024-01-03T00:00:00Z)
		  			|> filter(fn: (r) => r._measurement == "chain_activity_1d")
					|> filter(fn: (r) => r.emitter_chain == "1" or r.emitter_chain == "2")
					|> filter(fn: (r) => r.destination_chain == "3" or r.destination_chain == "4")
					|> filter(fn: (r) => r.app_id == "CCTP_WORMHOLE_INTEGRATION")
					|> drop(columns:["destination_chain"])

					vols = data		
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

					counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

					join.inner(
					    left: vols,
					    right: counts,
					    on: (l, r) => l._time == r._time and l.to == r.to and l.emitter_chain == r.emitter_chain,
					    as: (l, r) => ({l with count: r.count}),
					)
					|> group()
					|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search by multiple emitter_chain, destination_chain and app_id monthly",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{1, 2},
				TargetChains: []sdk.ChainID{3, 4},
				From:         time.Date(2023, 9, 7, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 3, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Month,
				AppId:        "CCTP_WORMHOLE_INTEGRATION",
			},
			expected: `
				import "date"
				import "join"

				data = from(bucket: "wormscan-testenv")
						|> range(start: 2023-09-01T00:00:00Z,stop: 2024-03-01T00:00:00Z)
						|> filter(fn: (r) => r._measurement == "chain_activity_1d")
						|> filter(fn: (r) => r.emitter_chain == "1" or r.emitter_chain == "2")
						|> filter(fn: (r) => r.destination_chain == "3" or r.destination_chain == "4")
						|> filter(fn: (r) => r.app_id == "CCTP_WORMHOLE_INTEGRATION")
						|> drop(columns:["destination_chain","to","app_id"])
						|> window(every: 1mo, period:1mo)
						|> drop(columns:["_time"])
						|> rename(columns: {_start: "_time"})
						|> map(fn: (r) => ({r with to: string(v: r._stop)}))

				vols = data
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

				counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

				join.inner(
						left: vols,
						right: counts,
						on: (l, r) => l._time == r._time and l.emitter_chain == r.emitter_chain,
						as: (l, r) => ({l with count: r.count}),
				)
				|> group()
				|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
		{
			name: "Search by multiple emitter_chain, destination_chain and app_id yearly",
			input: ChainActivityTopsQuery{
				SourceChains: []sdk.ChainID{1, 2},
				TargetChains: []sdk.ChainID{3, 4},
				From:         time.Date(2020, 9, 7, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2024, 3, 3, 5, 30, 5, 0, time.UTC),
				Timespan:     Year,
				AppId:        "CCTP_WORMHOLE_INTEGRATION",
			},
			expected: `
				import "date"
				import "join"

				data = from(bucket: "wormscan-testenv")
						|> range(start: 2020-01-01T00:00:00Z,stop: 2024-01-01T00:00:00Z)
						|> filter(fn: (r) => r._measurement == "chain_activity_1d")
						|> filter(fn: (r) => r.emitter_chain == "1" or r.emitter_chain == "2")
						|> filter(fn: (r) => r.destination_chain == "3" or r.destination_chain == "4")
						|> filter(fn: (r) => r.app_id == "CCTP_WORMHOLE_INTEGRATION")
						|> drop(columns:["destination_chain","to","app_id"])
						|> window(every: 1y, period:1y)
						|> drop(columns:["_time"])
						|> rename(columns: {_start: "_time"})
						|> map(fn: (r) => ({r with to: string(v: r._stop)}))

				vols = data
						|> filter(fn: (r) => (r._field == "volume" and r._value > 0))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "volume"})

				counts = data
						|> filter(fn: (r) => (r._field == "count"))
						|> group(columns:["_time","to","emitter_chain"])
						|> toUInt()
						|> sum()
						|> rename(columns: {_value: "count"})

				join.inner(
						left: vols,
						right: counts,
						on: (l, r) => l._time == r._time and l.emitter_chain == r.emitter_chain,
						as: (l, r) => ({l with count: r.count}),
				)
				|> group()
				|> sort(columns:["emitter_chain","_time"],desc:false)`,
		},
	}

	for _, testCase := range tcs {
		t.Run(testCase.name, func(t *testing.T) {
			got := repository.buildChainActivityQueryTops(testCase.input)
			assert.Equal(t, testCase.expected, got, "Expected query did not match actual one.")
		})
	}
}

func Test_buildAppActivityQuery(t *testing.T) {
	repository := &Repository{
		bucketInfiniteRetention: "wormscan-testenv",
		bucket30DaysRetention:   "wormscan-30days-testenv",
	}

	tcs := []struct {
		name                string
		input               ApplicationActivityQuery
		expectedAppQuery    string
		expectedTotalsQuery string
	}{
		{
			name: "Search by timespan monthly",
			input: ApplicationActivityQuery{
				AppId:    "CCTP_WORMHOLE_INTEGRATION",
				From:     time.Date(2023, 10, 7, 0, 0, 0, 0, time.UTC),
				To:       time.Date(2024, 3, 3, 5, 30, 5, 0, time.UTC),
				Timespan: Month,
			},
			expectedAppQuery:    "\n\t\t\timport \"date\"\n\t\t\timport \"join\"\n\n\t\t\tallData = from(bucket: \"wormscan-testenv\")\n\t\t\t\t\t\t|> range(start: 2023-10-01T00:00:00Z,stop: 2024-03-01T00:00:00Z)\n\t\t\t\t\t\t|> filter(fn: (r) => r._measurement == \"protocols_stats_1d\")\n\t\t\t\t\t\t|> filter(fn: (r) => not exists r.protocol )\n\t\t\t\t\t\t|> filter(fn: (r) => r.app_id_1 == \"CCTP_WORMHOLE_INTEGRATION\" or r.app_id_2 == \"CCTP_WORMHOLE_INTEGRATION\" or r.app_id_3 == \"CCTP_WORMHOLE_INTEGRATION\")\n\t\t\t\t\t\t|> drop(columns:[\"emitter_chain\",\"destination_chain\",\"_measurement\"])\n\n\t\t\ttotalMsgs = allData\n\t\t\t\t\t\t|> filter(fn: (r) => r._field == \"total_messages\")\n\t\t\t\t\t\t|> aggregateWindow(every: 1mo, fn: sum)\n\t\t\t\t\t\t|> rename(columns: {_value: \"total_messages\"})\n\t\t\t\t\t\t|> map(fn: (r) => ({\n\t\t\t\t\t\t\t\tr with\n\t\t\t\t\t\t\t\t_time: date.sub(d: 1mo, from: r._time),\n\t\t\t\t\t\t\t\ttotal_messages: if not exists r.total_messages then uint(v:0) else r.total_messages\n     \t\t\t\t\t\t}))\n\t\t\t\t\t\t|> drop(columns:[\"_start\",\"_stop\"])\n\t\t\t\t\t\t|> group()\n\t\t\t\n\t\t\t\n\t\t\ttvt = allData\n\t\t\t\t\t|> filter(fn: (r) => r._field == \"total_value_transferred\")\n\t\t\t\t\t|> aggregateWindow(every: 1mo, fn: sum)\n\t\t\t\t\t|> rename(columns: {_value: \"total_value_transferred\"})\t\t\n\t\t\t\t\t|> map(fn: (r) => ({\n\t\t\t\t\t\tr with\n\t\t\t\t\t\t_time: date.sub(d: 1mo, from: r._time),\n\t\t\t\t\t\ttotal_value_transferred: if not exists r.total_value_transferred then uint(v:0) else r.total_value_transferred\n\t\t\t\t\t}))\n\t\t\t\t\t|> drop(columns:[\"_start\",\"_stop\"])\n\t\t\t\t\t|> group()\n\t\t\t\t\t\t\n\t\t\tjoin.inner(\n\t\t\t    left: totalMsgs,\n\t\t\t    right: tvt,\n\t\t\t    on: (l, r) => l.app_id_1 == r.app_id_1 and l.app_id_2 == r.app_id_2 and l.app_id_3 == r.app_id_3 and l._time == r._time,\n\t\t\t    as: (l, r) => ({\n\t\t\t\t\t\"_time\":l._time,\n\t\t\t\t\t\"to\":date.add(d: 1mo, to: l._time),\n\t\t\t\t\t\"app_id_1\": l.app_id_1,\n\t\t\t\t\t\"app_id_2\": l.app_id_2,\n\t\t\t\t\t\"app_id_3\": l.app_id_3,\n\t\t\t\t\t\"total_messages\":l.total_messages,\n\t\t\t\t\t\"total_value_transferred\": float(v:r.total_value_transferred) / 100000000.0\n\t\t\t\t\t})\n\t\t\t)\n\t\t",
			expectedTotalsQuery: "\n\t\t\timport \"date\"\n\t\t\timport \"join\"\n\n\t\t\tallData = from(bucket: \"wormscan-testenv\")\n\t\t\t\t\t\t|> range(start: 2023-10-01T00:00:00Z,stop: 2024-03-01T00:00:00Z)\n\t\t\t\t\t\t|> filter(fn: (r) => r._measurement == \"protocols_stats_totals_1d\" and r.version == \"v1\")\n\t\t\t\t\t\t|> filter(fn: (r) => r.app_id == \"TOTAL_CCTP_WORMHOLE_INTEGRATION\")\n\t\t\t\t\t\t|> drop(columns:[\"emitter_chain\",\"destination_chain\",\"version\",\"_measurement\"])\n\t\t\t\n\t\t\ttotalMsgs = allData\n\t\t\t\t\t\t|> filter(fn: (r) => r._field == \"total_messages\")\n\t\t\t\t\t\t|> aggregateWindow(every: 1mo, fn: sum)\n\t\t\t\t\t\t|> rename(columns: {_value: \"total_messages\"})\n\t\t\t\t\t\t|> group()\n\t\t\t\t\t\t\n\t\t\ttvt = allData\n\t\t\t\t\t\t|> filter(fn: (r) => r._field == \"total_value_transferred\")\n\t\t\t\t\t\t|> aggregateWindow(every: 1mo, fn: sum)\n\t\t\t\t\t\t|> rename(columns: {_value: \"total_value_transferred\"})\n\t\t\t\t\t\t|> group()\n\n\t\t\tjoin.inner(\n\t\t\t    left: totalMsgs,\n\t\t\t    right: tvt,\n\t\t\t    on: (l, r) => l.app_id == r.app_id and l._time == r._time,\n\t\t\t    as: (l, r) => ({\n\t\t\t\t\t\"to\":l._time,\n\t\t\t\t\t\"_time\": date.sub(d: 1mo, from: l._time),\n\t\t\t\t\t\"app_id\": l.app_id,\n\t\t\t\t\t\"total_messages\":l.total_messages,\n\t\t\t\t\t\"total_value_transferred\": float(v:r.total_value_transferred) / 100000000.0\n\t\t\t\t\t}),\n\t\t\t)\n\t",
		},
		{
			name: "Search by timespan hourly",
			input: ApplicationActivityQuery{
				AppId:    "CCTP_WORMHOLE_INTEGRATION",
				From:     time.Date(2023, 10, 7, 11, 13, 55, 0, time.UTC),
				To:       time.Date(2024, 3, 3, 5, 30, 5, 0, time.UTC),
				Timespan: Hour,
			},
			expectedAppQuery:    "\n\t\t\timport \"date\"\n\n\t\t\t\tallData = from(bucket: \"wormscan-30days-testenv\")\n\t\t\t\t\t\t\t|> range(start: 2023-10-07T11:00:00Z,stop: 2024-03-03T05:00:00Z)\n\t\t\t\t\t\t\t|> filter(fn: (r) => r._measurement == \"protocols_stats_1h\")\n\t\t\t\t\t\t\t|> filter(fn: (r) => not exists r.protocol )\n\t\t\t\t\t\t\t|> filter(fn: (r) => r.app_id_1 == \"CCTP_WORMHOLE_INTEGRATION\" or r.app_id_2 == \"CCTP_WORMHOLE_INTEGRATION\" or r.app_id_3 == \"CCTP_WORMHOLE_INTEGRATION\")\n\t\t\t\t\t\t\t|> drop(columns:[\"emitter_chain\",\"destination_chain\",\"_measurement\"])\n\n\t\t\t\ttotalMsgs = allData\n\t\t\t\t\t\t\t|> filter(fn: (r) => r._field == \"total_messages\")\n\t\t\t\t\t\t\t|> aggregateWindow(every: 1h, fn: sum, createEmpty:true)\n\t\t\t\t\t\t\t|> map(fn: (r) => ({\n\t\t\t\t\t\t\t\t\t\tr with\n\t\t\t\t\t\t\t\t\t\t_value: if not exists r._value then uint(v:0) else r._value\n\t\t\t\t\t\t\t\t}))\n\t\t\t\t\t\t\t|> group(columns:[\"_time\",\"_field\",\"app_id_1\",\"app_id_2\",\"app_id_3\"])\n\t\t\t\t\t\t\t|> sum()\n\t\t\t\t\t\t\n\t\t\t\ttvt = allData\n\t\t\t\t\t\t|> filter(fn: (r) => r._field == \"total_value_transferred\")\n\t\t\t\t\t\t|> aggregateWindow(every: 1h, fn: sum, createEmpty:true)\n\t\t\t\t\t\t|> map(fn: (r) => ({\n\t\t\t\t\t\t\t\tr with\n\t\t\t\t\t\t\t\t_value: if not exists r._value then uint(v:0) else r._value\n     \t\t\t\t\t\t}))\n\t\t\t\t\t\t|> group(columns:[\"_time\",\"_field\",\"app_id_1\",\"app_id_2\",\"app_id_3\"])\n\t\t\t\t\t\t|> sum()\n\t\t\t\t\t\t\n\t\t\t\tunion(tables: [totalMsgs, tvt])\n\t\t\t\t|> pivot(rowKey:[\"_time\",\"app_id_1\",\"app_id_2\",\"app_id_3\"], columnKey: [\"_field\"], valueColumn: \"_value\")\n\t\t\t\t|> map(fn: (r) => ({\n\t\t\t\t\t\tr with\n\t\t\t\t\t\t\"total_value_transferred\": float(v:r.total_value_transferred) / 100000000.0,\n\t\t\t\t\t\t\"to\": r._time,\n\t\t\t\t\t\t\"_time\": date.sub(d: 1h, from: r._time)\n\t\t\t\t}))",
			expectedTotalsQuery: "\n\t\t\timport \"date\"\n\n\t\t\tallData = from(bucket: \"wormscan-30days-testenv\")\n\t\t\t\t\t\t|> range(start: 2023-10-07T11:00:00Z,stop: 2024-03-03T05:00:00Z)\n\t\t\t\t\t\t|> filter(fn: (r) => r._measurement == \"protocols_stats_totals_1h\")\n\t\t\t\t\t\t|> filter(fn: (r) => r.app_id == \"TOTAL_CCTP_WORMHOLE_INTEGRATION\")\n\t\t\t\t\t\t|> drop(columns:[\"emitter_chain\",\"destination_chain\"])\n\t\t\t\n\t\t\ttotalMsgs = allData\n\t\t\t\t\t\t|> filter(fn: (r) => r._field == \"total_messages\")\n\t\t\t\t\t\t|> aggregateWindow(every: 1h, fn: sum,createEmpty:true)\n\t\t\t\t\t\t|> map(fn: (r) => ({\n\t\t\t\t\t\t\t\tr with\n\t\t\t\t\t\t\t\t_value: if not exists r._value then uint(v:0) else uint(v:r._value)\n     \t\t\t\t\t\t}))\n\t\t\t\t\t\t|> group(columns:[\"_time\",\"app_id\",\"_field\"])\n\t\t\t\t\t\t|> sum()\n\t\t\t\t\t\t\n\t\t\ttvt = allData\n\t\t\t\t\t\t|> filter(fn: (r) => r._field == \"total_value_transferred\")\n\t\t\t\t\t\t|> aggregateWindow(every: 1h, fn: sum, createEmpty:true)\n\t\t\t\t\t\t|> map(fn: (r) => ({\n\t\t\t\t\t\t\t\tr with\n\t\t\t\t\t\t\t\t_value: if not exists r._value then uint(v:0) else r._value\n     \t\t\t\t\t\t}))\n\t\t\t\t\t\t|> group(columns:[\"_time\",\"app_id\",\"_field\"])\n\t\t\t\t\t\t|> sum()\n\n\t\t\tunion(tables: [totalMsgs, tvt])\n\t\t\t\t|> pivot(rowKey:[\"_time\",\"app_id\"], columnKey: [\"_field\"], valueColumn: \"_value\")\n\t\t\t\t|> map(fn: (r) => ({\n\t\t\t\t\t\tr with\n\t\t\t\t\t\t\"total_value_transferred\": float(v:r.total_value_transferred) / 100000000.0,\n\t\t\t\t\t\t\"to\": r._time,\n\t\t\t\t\t\t\"_time\": date.sub(d: 1h, from: r._time)\n     \t\t\t}))\n\t\t\t",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			query := repository.buildAppActivityQuery(tc.input)
			assert.Equal(t, tc.expectedAppQuery, query)
			totalsQuery := repository.buildTotalsAppActivityQuery(tc.input)
			assert.Equal(t, tc.expectedTotalsQuery, totalsQuery)
		})
	}

}

func Test_buildTokenSymbolActivityQuery(t *testing.T) {
	repository := &Repository{
		bucketInfiniteRetention: "wormscan-testenv",
		bucket30DaysRetention:   "wormscan-30days-testenv",
	}

	tcs := []struct {
		name          string
		input         TokenSymbolActivityQuery
		expectedQuery string
	}{
		{
			name: "Hourly timespan with single token symbol and single source/target chain",
			input: TokenSymbolActivityQuery{
				From:         time.Date(2023, 8, 1, 12, 0, 0, 0, time.UTC),
				To:           time.Date(2023, 8, 1, 13, 0, 0, 0, time.UTC),
				TokenSymbols: []string{"BTC"},
				SourceChains: []sdk.ChainID{1},
				TargetChains: []sdk.ChainID{2},
				Timespan:     Hour,
			},
			expectedQuery: `
	import "date"

	sumAndCount = (tables=<-, column) => {
		return tables
				|> reduce(
					identity: {
						_value: uint(v:0),
						txs: uint(v:0)
					},
					fn: (r, accumulator) => ({
						_value: accumulator._value + r._value,
						txs: accumulator.txs + uint(v:1)
					})
				)
	}
	
	from(bucket: "wormscan-testenv")
		|> range(start: 2023-08-01T12:00:00Z, stop: 2023-08-01T13:00:00Z)
		|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
		|> filter(fn: (r) => r._field == "volume" or r._field == "symbol")
		|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> keep(columns:["_start","_stop","_time","emitter_chain","destination_chain","symbol","volume"])
		|> filter(fn: (r) => r.volume > 0)
		|> filter(fn: (r) => r.symbol == "BTC") //filter by symbol
		|> filter(fn: (r) => r.emitter_chain == "1") //filter by source_chain
		|> filter(fn: (r) => r.destination_chain == "2") //filter by target_chain
		|> rename(columns: {volume: "_value"})
		|> set(key: "_field", value: "volume")
		|> group(columns:["symbol","emitter_chain","destination_chain","_field"])
		|> aggregateWindow(every: 1h, fn: sumAndCount, createEmpty: true)
		|> map(fn: (r) => ({
				r with 
				volume: if exists r._value then float(v:r._value) / 100000000.0 else float(v:0),
				to: r._time,
				_time: date.sub(d: 1h, from: r._time),
		}))
		|> drop(columns:["_value","_start","_stop","_field"])	
	`,
		},
		{
			name: "Daily timespan with multiple token symbols and multiple source/target chains",
			input: TokenSymbolActivityQuery{
				From:         time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC),
				To:           time.Date(2023, 8, 2, 0, 0, 0, 0, time.UTC),
				TokenSymbols: []string{"BTC", "ETH"},
				SourceChains: []sdk.ChainID{1, 2},
				TargetChains: []sdk.ChainID{3, 4},
				Timespan:     Day,
			},
			expectedQuery: `
	import "date"

	sumAndCount = (tables=<-, column) => {
		return tables
				|> reduce(
					identity: {
						_value: uint(v:0),
						txs: uint(v:0)
					},
					fn: (r, accumulator) => ({
						_value: accumulator._value + r._value,
						txs: accumulator.txs + uint(v:1)
					})
				)
	}
	
	from(bucket: "wormscan-testenv")
		|> range(start: 2023-08-01T00:00:00Z, stop: 2023-08-02T00:00:00Z)
		|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
		|> filter(fn: (r) => r._field == "volume" or r._field == "symbol")
		|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> keep(columns:["_start","_stop","_time","emitter_chain","destination_chain","symbol","volume"])
		|> filter(fn: (r) => r.volume > 0)
		|> filter(fn: (r) => r.symbol == "BTC" or r.symbol == "ETH") //filter by symbol
		|> filter(fn: (r) => r.emitter_chain == "1" or r.emitter_chain == "2") //filter by source_chain
		|> filter(fn: (r) => r.destination_chain == "3" or r.destination_chain == "4") //filter by target_chain
		|> rename(columns: {volume: "_value"})
		|> set(key: "_field", value: "volume")
		|> group(columns:["symbol","emitter_chain","destination_chain","_field"])
		|> aggregateWindow(every: 1d, fn: sumAndCount, createEmpty: true)
		|> map(fn: (r) => ({
				r with 
				volume: if exists r._value then float(v:r._value) / 100000000.0 else float(v:0),
				to: r._time,
				_time: date.sub(d: 1d, from: r._time),
		}))
		|> drop(columns:["_value","_start","_stop","_field"])	
	`,
		},
		{
			name: "Monthly timespan with no token symbols and no chains",
			input: TokenSymbolActivityQuery{
				From:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				To:       time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
				Timespan: Month,
			},
			expectedQuery: `
	import "date"

	sumAndCount = (tables=<-, column) => {
		return tables
				|> reduce(
					identity: {
						_value: uint(v:0),
						txs: uint(v:0)
					},
					fn: (r, accumulator) => ({
						_value: accumulator._value + r._value,
						txs: accumulator.txs + uint(v:1)
					})
				)
	}
	
	from(bucket: "wormscan-testenv")
		|> range(start: 2023-01-01T00:00:00Z, stop: 2023-06-01T00:00:00Z)
		|> filter(fn: (r) => r._measurement == "vaa_volume_v3" and r.version == "v5")
		|> filter(fn: (r) => r._field == "volume" or r._field == "symbol")
		|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> keep(columns:["_start","_stop","_time","emitter_chain","destination_chain","symbol","volume"])
		|> filter(fn: (r) => r.volume > 0)
		 //filter by symbol
		 //filter by source_chain
		 //filter by target_chain
		|> rename(columns: {volume: "_value"})
		|> set(key: "_field", value: "volume")
		|> group(columns:["symbol","emitter_chain","destination_chain","_field"])
		|> aggregateWindow(every: 1mo, fn: sumAndCount, createEmpty: true)
		|> map(fn: (r) => ({
				r with 
				volume: if exists r._value then float(v:r._value) / 100000000.0 else float(v:0),
				to: r._time,
				_time: date.sub(d: 1mo, from: r._time),
		}))
		|> drop(columns:["_value","_start","_stop","_field"])	
	`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			query := repository.buildTokenSymbolActivityQuery(tc.input)
			assert.Equal(t, strings.TrimSpace(tc.expectedQuery), strings.TrimSpace(query))
		})
	}
}

func TestGetScorecards(t *testing.T) {

	m := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name             string
		mockTvlReturn    string
		mockTvlErr       error
		mockQueryResults map[string]struct {
			res           *mockInfluxQueryResult
			expectedQuery string
		}
		mockPythResponse   bson.D
		expectedErr        bool
		expectedScorecards *Scorecards
	}{
		{
			name:          "All queries succeed",
			mockTvlReturn: "1000",
			mockTvlErr:    nil,
			mockQueryResults: map[string]struct {
				res           *mockInfluxQueryResult
				expectedQuery string
			}{
				"messages24h":   {mockInfluxResult(100), buildMessages24HrQuery("wormscan-24hours")},
				"totalTxCount":  {mockInfluxResult(100), buildTotalTrxCountQuery("wormscan", "wormscan-30days", time.Now())},
				"totalTxVolume": {mockInfluxResult(1000e8), buildTotalTrxVolumeQuery("wormscan", "wormscan-30days", time.Now())},
				"volume24h":     {mockInfluxResult(200e8), buildVolumeQuery("wormscan", _24h, []string{"MAYAN"})},
				"volume7d":      {mockInfluxResult(1500e8), buildVolumeQuery("wormscan", _7d, []string{"MAYAN"})},
				"mayan7d":       {mockInfluxResult(5000e8), buildMayanQuery("wormscan", _7d)},
				"volume30d":     {mockInfluxResult(5000e8), buildVolumeQuery("wormscan", _30d, []string{"MAYAN"})},
				"mayan30d":      {mockInfluxResult(5000e8), buildMayanQuery("wormscan", _30d)},
			},
			mockPythResponse: bson.D{{"_id", "some-id"}, {"sequence", "123456"}},
			expectedErr:      false,
			expectedScorecards: &Scorecards{
				Messages24h:   "100",
				TotalMessages: "965587054", // 965463498 + 100 + 123456
				TotalTxCount:  "100",
				TotalTxVolume: "1000.00000000",
				Tvl:           "1000",
				Volume24h:     "13294486.00000000",
				Volume7d:      "500000001500.00000000",
				Volume30d:     "500000005000.00000000",
			},
		},

		{
			name:          "Tvl query fails",
			mockTvlReturn: "",
			mockTvlErr:    errors.New("mock_tvl_error"),
			mockQueryResults: map[string]struct {
				res           *mockInfluxQueryResult
				expectedQuery string
			}{
				"messages24h":   {mockInfluxResult(100), buildMessages24HrQuery("wormscan-24hours")},
				"totalTxCount":  {mockInfluxResult(100), buildTotalTrxCountQuery("wormscan", "wormscan-30days", time.Now())},
				"totalTxVolume": {mockInfluxResult(1000e8), buildTotalTrxVolumeQuery("wormscan", "wormscan-30days", time.Now())},
				"volume24h":     {mockInfluxResult(200e8), buildVolumeQuery("wormscan", _24h, []string{"MAYAN"})},
				"volume7d":      {mockInfluxResult(1500e8), buildVolumeQuery("wormscan", _7d, []string{"MAYAN"})},
				"volume30d":     {mockInfluxResult(5000e8), buildVolumeQuery("wormscan", _30d, []string{"MAYAN"})},
				"mayan7d":       {mockInfluxResult(5000e8), buildMayanQuery("wormscan", _7d)},
				"mayan30d":      {mockInfluxResult(5000e8), buildMayanQuery("wormscan", _30d)},
			},
			mockPythResponse:   bson.D{{"_id", "some-id"}, {"sequence", "123456"}},
			expectedErr:        true,
			expectedScorecards: nil,
		},

		{
			name:          "Multiple queries fail",
			mockTvlReturn: "1000",
			mockTvlErr:    nil,
			mockQueryResults: map[string]struct {
				res           *mockInfluxQueryResult
				expectedQuery string
			}{
				"messages24h":   {mockInfluxError("failed_query"), buildMessages24HrQuery("wormscan-24hours")},
				"totalTxCount":  {mockInfluxResult(100), buildTotalTrxCountQuery("wormscan", "wormscan-30days", time.Now())},
				"totalTxVolume": {mockInfluxError("failed_query"), buildTotalTrxVolumeQuery("wormscan", "wormscan-30days", time.Now())},
				"volume24h":     {mockInfluxResult(200e8), buildVolumeQuery("wormscan", _24h, []string{"MAYAN"})},
				"volume7d":      {mockInfluxResult(1500e8), buildVolumeQuery("wormscan", _7d, []string{"MAYAN"})},
				"volume30d":     {mockInfluxResult(5000e8), buildVolumeQuery("wormscan", _30d, []string{"MAYAN"})},
				"mayan7d":       {mockInfluxResult(5000e8), buildMayanQuery("wormscan", _7d)},
				"mayan30d":      {mockInfluxResult(5000e8), buildMayanQuery("wormscan", _30d)},
			},
			mockPythResponse:   nil, // Simulating no documents found
			expectedErr:        true,
			expectedScorecards: nil,
		},
	}

	for _, tt := range tests {
		m.Run(tt.name, func(mt *mtest.T) {
			tvlMock := new(mockTvl)
			queryAPIMock := new(mockQueryAPI)
			mt.AddMockResponses(mtest.CreateCursorResponse(1, "wormhole.vaasPythnet", "firstBatch", tt.mockPythResponse))

			repo := &Repository{
				tvl:        tvlMock,
				queryAPI:   queryAPIMock,
				logger:     logger,
				p2pNetwork: config.P2pMainNet,
				collections: repositoryCollections{
					vaasPythnet: mt.Coll,
				},
				bucket24HoursRetention:  "wormscan-24hours",
				bucket30DaysRetention:   "wormscan-30days",
				bucketInfiniteRetention: "wormscan",
				mayanHttpClient:         mockMayanHttpClient,
			}

			tvlMock.On("Get", mock.Anything).Return(tt.mockTvlReturn, tt.mockTvlErr)

			for _, result := range tt.mockQueryResults {
				queryAPIMock.On("Query", mock.Anything, result.expectedQuery).Return(result.res, nil)
			}

			scorecards, err := repo.GetScorecards(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, scorecards)
				assert.Equal(t, tt.expectedScorecards, scorecards)
			}

			tvlMock.AssertExpectations(t)
			queryAPIMock.AssertExpectations(t)
		})
	}
}

func mockInfluxResult(value uint64) *mockInfluxQueryResult {
	result := new(mockInfluxQueryResult)
	result.On("Next").Return(true)
	result.On("Record").Return(query.NewFluxRecord(0, map[string]interface{}{"_value": value}))
	result.On("Err").Return(nil)
	return result
}

func mockInfluxError(errMsg string) *mockInfluxQueryResult {
	result := new(mockInfluxQueryResult)
	result.On("Next").Return(false)
	result.On("Err").Return(errors.New(errMsg))
	return result
}

// Mocking the Tvl interface
type mockTvl struct {
	mock.Mock
}

func (m *mockTvl) Get(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

type mockQueryAPI struct {
	mock.Mock
}

func (m *mockQueryAPI) Query(ctx context.Context, query string) (influxQueryResult, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(influxQueryResult), args.Error(1)
}

type mockInfluxQueryResult struct {
	mock.Mock
}

func (m *mockInfluxQueryResult) Err() error {
	return m.Called().Error(0)
}

func (m *mockInfluxQueryResult) Next() bool {
	return m.Called().Bool(0)
}

func (m *mockInfluxQueryResult) Record() *query.FluxRecord {
	args := m.Called()
	return args.Get(0).(*query.FluxRecord)
}

func mockMayanHttpClient(req *http.Request) (*http.Response, error) {
	jsonData := `{
		"last24h": {
			"volume": 13294286,
			"toSolCount": 3440,
			"fromSolCount": 1471,
			"swaps": 6092,
			"activeTraders": 5045
		},
		"allTime": {
			"volume": 1619896737,
			"toSolCount": 277735,
			"fromSolCount": 170532,
			"swaps": 617677,
			"activeTraders": 248313
		}
	}`

	// Create a new reader with the JSON string
	r := io.NopCloser(bytes.NewReader([]byte(jsonData)))

	// Create a new HTTP response
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       r,
		Header:     make(http.Header),
	}

	return resp, nil
}
