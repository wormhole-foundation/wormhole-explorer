package metrics

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/shopspring/decimal"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"github.com/xlabs/influx-backfiller/coingecko"
	"github.com/xlabs/influx-backfiller/prices"
	"github.com/xlabs/influx-backfiller/tokens"
)

type LineParser struct {
	hasher               hash.Hash32
	MissingTokens        map[sdk.Address]sdk.ChainID
	MissingTokensCounter map[sdk.Address]int
	Coingecko            coingecko.CoinGeckoAPI
	includeVaaID         bool
	TokeList             *[]tokens.TokenConfigEntry
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
			fout.Write([]byte(nl + "\n"))

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
	tokenList := tokens.TokenList()
	priceCache := prices.NewCoinPricesCache("prices.csv")
	return &LineParser{
		hasher:               fnv.New32a(),
		MissingTokens:        make(map[sdk.Address]sdk.ChainID),
		MissingTokensCounter: make(map[sdk.Address]int),
		Coingecko:            *coingecko.NewCoinGeckoAPI(""),
		includeVaaID:         false,
		TokeList:             &tokenList,
		PriceCache:           priceCache,
	}
}

// generate influxdb line protocol format
// vaa,tags fields timestamp
// vaa_count,emitter_chain=solana,emitter_token=ETH,destination_address=0x123123123123 amount=10221,notional=1.230 timestamp
func (lp *LineParser) ParseLine(line []byte) (string, error) {
	tt := strings.Split(string(line), ",")
	//fmt.Printf("bcid %s, emmiter %s, seq %s\n", header[0], header[1], header[2])

	if len(tt) != 2 {
		return "", fmt.Errorf("invalid line: %s", line)
	}

	data, err := hex.DecodeString(tt[1])
	if err != nil {
		return "", fmt.Errorf("error decoding: %v", err)
	}

	vaa, err := sdk.Unmarshal(data)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling vaa: %v", err)
	}

	payload, err := sdk.DecodeTransferPayloadHdr(vaa.Payload)
	if err != nil {
		return "", fmt.Errorf("error decoding payload: %v", err)
	}
	//lp.hasher.Reset()
	//lp.hasher.Write([]byte(v.MessageID()))
	//fmt.Println(string(v.Payload))

	ochain := vaa.EmitterChain.String()
	if strings.HasPrefix(ochain, "unknown") {
		ochain = "unknown"
	}

	tchain := payload.TargetChain.String()
	if strings.HasPrefix(tchain, "unknown") {
		tchain = "unknown"
	}

	var symbol string
	var formatedAmount string
	var price *decimal.Decimal
	// look for the token by address
	token := tokens.TokenLookup(lp.TokeList, uint16(vaa.EmitterChain), payload.OriginAddress.String())
	if token == nil {
		// if not found, add to missing tokens
		lp.MissingTokens[payload.OriginAddress] = vaa.EmitterChain
		lp.MissingTokensCounter[payload.OriginAddress] = lp.MissingTokensCounter[payload.OriginAddress] + 1
		symbol = "none"
		formatedAmount = formatAmount(payload.Amount)
		tmp := decimal.NewFromInt(0)
		price = &tmp
	} else {
		if token.Symbol == "" {
			symbol = token.CoingeckoId
		} else {
			symbol = token.Symbol
		}
		// format amount using decimals from table
		formatedAmount = formatAmountWithDecimals(payload.Amount, int(token.Decimals))

		// get price from cache
		price, err = lp.PriceCache.GetPriceByTime(int16(vaa.GetEmitterChain()), token.CoingeckoId, vaa.Timestamp)
		if err != nil {
			// nothing to do
			return "", fmt.Errorf("error getting price: %v", err)
		}
	}

	famount, err := decimal.NewFromString(formatedAmount)
	if err != nil {
		return "", fmt.Errorf("error parsing amount: %v", err)
	}

	// build the metric
	newline := fmt.Sprintf("vaa,origin_chain=%s,target_chain=%s,token=%s,origin_address=%s amount=%s,notional=%s,volume=%s %d",
		//v.MessageID(),
		//payload.Type,
		ochain,
		tchain,
		symbol,
		vaa.EmitterAddress.String(),
		//payload.OriginAddress.String(),
		//payload.TargetAddress.String(),
		formatedAmount,
		price,
		price.Mul(famount),
		vaa.Timestamp.UnixNano())

	return newline, nil

}

// if we dont know the token, we can try to infer the decimals
func inferPrecision(amount *big.Int) int {
	l := len(amount.String())
	if l > 8 && l < 18 {
		return 8
	}
	if l < 9 {
		return 6
	}
	if l > 18 {
		return 18
	}
	return l
}

func formatAmount(amount *big.Int) string {

	p := inferPrecision(amount)
	s := amount.String()
	// put a comma in the right place
	if p < len(s) {
		s = s[:len(s)-p] + "." + s[len(s)-p:]
	} else {
		s = "0." + strings.Repeat("0", p-len(s)) + s
	}

	return s

}

func formatAmountWithDecimals(amount *big.Int, decimals int) string {

	s := amount.String()
	if decimals < len(s) {
		s = s[:len(s)-decimals] + "." + s[len(s)-decimals:]
	} else {
		s = "0." + strings.Repeat("0", decimals-len(s)) + s
	}
	// remove trailing zeros
	s = strings.TrimRight(s, "0")

	// add a zero if the number ends with a dot
	if s[len(s)-1] == '.' {
		s = s + "0"
	}

	return s
}
