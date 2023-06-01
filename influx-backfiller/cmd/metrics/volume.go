package metrics

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/analytic/metric"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"github.com/xlabs/influx-backfiller/prices"
)

type LineParser struct {
	MissingTokens        map[sdk.Address]sdk.ChainID
	MissingTokensCounter map[sdk.Address]int
	PriceCache           *prices.CoinPricesCache
}

// read a csv file with VAAs and convert into a decoded csv file
// ready to upload to the database
func RunVaaVolume(inputFile, outputFile string) {

	// open input file
	f, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// create missing tokens file
	missingTokensFile := "missing_tokens.csv"
	fmissingTokens, err := os.Create(missingTokensFile)
	if err != nil {
		panic(err)
	}
	defer fmissingTokens.Close()

	//open output file for writing
	fout, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	// init price cache!
	fmt.Println("loading historical prices...")
	lp := NewLineParser()
	lp.PriceCache.InitCache()
	fmt.Println("done!")

	r := bufio.NewReader(f)

	c := 0
	i := 0
	// read file line by line and send to workpool
	for {
		line, _, err := r.ReadLine() //loading chunk into buffer
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("a real error happened here: %v\n", err)
		}
		nl, err := lp.ParseLine(line)
		if err != nil {
			//fmt.Printf(",")
		} else {
			c++
			fout.Write([]byte(nl))

			if c == 10000 {
				fmt.Printf(".")
				c = 0
				i := i + 1
				if i == 10 {
					fmt.Printf("\n")
					i = 0
				}
			}

		}

	}

	for k, v := range lp.MissingTokensCounter {
		fmt.Printf("missing token %s %s %d\n", k.String(), lp.MissingTokens[k], v)
		fmissingTokens.WriteString(fmt.Sprintf("%s,%s,%d\n", k.String(), lp.MissingTokens[k], lp.MissingTokensCounter[k]))
	}

	fmt.Println("done!")

}

func NewLineParser() *LineParser {
	priceCache := prices.NewCoinPricesCache("prices.csv")
	return &LineParser{
		MissingTokens:        make(map[sdk.Address]sdk.ChainID),
		MissingTokensCounter: make(map[sdk.Address]int),
		PriceCache:           priceCache,
	}
}

// ParseLine takes a CSV line as input, and generates a line protocol entry as output.
//
// The format for InfluxDB line protocol is: vaa,tags fields timestamp
func (lp *LineParser) ParseLine(line []byte) (string, error) {

	// Parse the VAA and payload
	var vaa *sdk.VAA
	var payload *sdk.TransferPayloadHdr
	{
		tt := strings.Split(string(line), ",")
		if len(tt) != 2 {
			return "", fmt.Errorf("expected line to have two tokens, but has %d: %s", len(tt), line)
		}
		vaaBytes, err := hex.DecodeString(tt[1])
		if err != nil {
			return "", fmt.Errorf("error decoding: %v", err)
		}
		vaa, err = sdk.Unmarshal(vaaBytes)
		if err != nil {
			return "", fmt.Errorf("error unmarshaling vaa: %v", err)
		}
		payload, err = sdk.DecodeTransferPayloadHdr(vaa.Payload)
		if err != nil {
			return "", fmt.Errorf("error decoding payload: %v", err)
		}
	}

	// Look up token metadata
	tokenMetadata, ok := domain.GetTokenByAddress(vaa.EmitterChain, payload.OriginAddress.String())
	if !ok {

		// if not found, add to missing tokens
		lp.MissingTokens[payload.OriginAddress] = vaa.EmitterChain
		lp.MissingTokensCounter[payload.OriginAddress] = lp.MissingTokensCounter[payload.OriginAddress] + 1

		return "", fmt.Errorf("unknown token: %s %s", payload.OriginChain.String(), payload.OriginAddress.String())
	}

	// Generate a data point for the VAA volume metric
	var point *write.Point
	{
		p := metric.MakePointForVaaVolumeParams{
			Vaa: vaa,
			TokenPriceFunc: func(_ domain.Symbol, timestamp time.Time) (decimal.Decimal, error) {

				// fetch the historic price from cache
				price, err := lp.PriceCache.GetPriceByTime(tokenMetadata.CoingeckoID, timestamp)
				if err != nil {
					return decimal.NewFromInt(0), err
				}

				return price, nil
			},
		}

		var err error
		point, err = metric.MakePointForVaaVolume(&p)
		if err != nil {
			return "", fmt.Errorf("failed to create data point for VAA volume metric: %v", err)
		}
		if point == nil {
			// Some VAAs don't generate any data points for this metric (e.g.: PythNet, non-token-bridge VAAs)
			return "", errors.New("can't generate point for VAA volume metric")
		}
	}

	// Convert the data point to line protocol
	result := convertPointToLineProtocol(point)
	return result, nil
}
