package metrics

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/analytics/metric"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/parser"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func RunVaaCount(inputFile, outputFile string) {

	// Create the output file
	fout, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	time30DaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	fmt.Println(time30DaysAgo)

	// Define a processor function that will be called for each input VAA
	processorFunc := func(vaa *sdk.VAA) error {

		// Call the analytics module to generate the data point for this VAA
		point, err := metric.MakePointForVaaCount(vaa)
		if err != nil {
			return err
		}
		if point == nil {
			// Some VAAs don't generate any data points for this metric (e.g.: PythNet)
			return nil
		}

		if point.Time().Before(time30DaysAgo) {
			// Ignore VAAs that are older than 30 days
			return nil
		}

		// Write a new in the dump file
		line := convertPointToLineProtocol(point)
		if _, err := fout.Write([]byte(line)); err != nil {
			return err
		}

		return nil
	}

	csvParser := parser.NewVaaCsvParser(processorFunc, inputFile)

	csvParser.Start(context.Background())
}
