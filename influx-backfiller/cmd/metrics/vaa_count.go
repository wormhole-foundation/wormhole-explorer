package metrics

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/wormhole-foundation/wormhole-explorer/analytic/metric"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"github.com/xlabs/influx-backfiller/parser"
)

func RunVaaCount(inputFile, outputFile string) {

	// Create the output file
	fout, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	// Define a processor function that will be called for each input VAA
	processorFunc := func(vaa *sdk.VAA) error {

		// Call the analytics module to generate the data point for this VAA
		point, err := metric.GenerateDataPointForVaaCount(vaa)
		if err != nil {
			return err
		}
		if point == nil {
			return nil
		}

		// Write the data point in the format that InfluxDB uses for dumps.
		// See https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/
		{
			// Collect tags
			var tags string
			for _, t := range point.TagList() {
				tags += fmt.Sprintf(",%s=%s", t.Key, t.Value)
			}

			// Collect fields
			var tmp []string
			for _, f := range point.FieldList() {
				tmp = append(tmp, fmt.Sprintf("%s=%v", f.Key, f.Value))
			}
			fields := strings.Join(tmp, ",")

			// Build a line for the dump file
			line := fmt.Sprintf("%s%s %s %d\n", point.Name(), tags, fields, point.Time().UnixNano())
			if _, err := fout.Write([]byte(line)); err != nil {
				return err
			}
		}

		return nil
	}

	csvParser := parser.NewVaaCsvParser(processorFunc, inputFile)

	csvParser.Start(context.Background())
}
