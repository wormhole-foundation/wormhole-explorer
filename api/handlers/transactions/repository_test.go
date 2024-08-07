package transactions

import (
	"github.com/stretchr/testify/assert"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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
