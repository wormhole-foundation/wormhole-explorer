package metrics

import (
	"context"
	"fmt"
	"os"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"github.com/xlabs/influx-backfiller/parser"
)

func RunVaaCount(inputFile, outputFile string) {
	fout, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	csvParser := parser.NewVaaCsvParser(
		func(vaa *sdk.VAA) error {
			line := fmt.Sprintf("vaa_count,chain_id=%d count=1 %d\n", vaa.EmitterChain, vaa.Timestamp.UnixNano())
			if _, err := fout.Write([]byte(line)); err != nil {
				return err
			}
			return nil
		},
		inputFile)

	csvParser.Start(context.Background())
}
