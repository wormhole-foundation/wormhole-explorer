package stats

import (
	"fmt"
	"time"
)

const queryTemplateSymbolWithAssets = `
from(bucket: "%s")
    |> range(start: %s)
    |> filter(fn: (r) => r._measurement == "%s" and r._field == "txs_volume")
    |> last()
    |> group()
`

func buildSymbolWithAssets(bucket string, t time.Time, measurement string) string {
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	return fmt.Sprintf(queryTemplateSymbolWithAssets, bucket, start, measurement)
}

const queryTemplateTopCorridors = `
from(bucket: "%s")
    |> range(start: %s)
    |> filter(fn: (r) => r._measurement == "%s" and r._field == "count")
    |> last()
    |> group()
`

func buildTopCorridors(bucket string, t time.Time, measurement string) string {
	start := t.Truncate(time.Hour * 24).Format(time.RFC3339Nano)
	return fmt.Sprintf(queryTemplateTopCorridors, bucket, start, measurement)
}
