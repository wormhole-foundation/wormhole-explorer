package metrics

import (
	"fmt"
	"strings"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// convertPointToLineProtocol transforms a given data point into the format that InfluxDB uses for dumps.
//
// See https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/
func convertPointToLineProtocol(point *write.Point) string {

	// Collect tags
	var tags string
	for _, t := range point.TagList() {
		tags += fmt.Sprintf(",%s=%s", t.Key, t.Value)
	}

	// Collect fields
	if len(point.FieldList()) == 0 {
		panic("expected at least one point in metric")
	}
	var tmp []string
	for _, f := range point.FieldList() {
		tmp = append(tmp, fmt.Sprintf("%s=%v", f.Key, f.Value))
	}
	fields := strings.Join(tmp, ",")

	// Build a line for the dump file
	return fmt.Sprintf("%s%s %s %d\n", point.Name(), tags, fields, point.Time().UnixNano())
}
