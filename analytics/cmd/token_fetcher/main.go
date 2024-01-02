package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mr-tron/base58"
	"github.com/wormhole-foundation/wormhole-explorer/common/coingecko"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("usage: %s <input file> <coingecko-url>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	f, err := os.Open(filename)
	coingeckoURL := os.Args[2]
	cg := coingecko.NewCoinGeckoAPI(coingeckoURL, "", "")

	if err != nil {
		panic(err)
	}
	defer f.Close()

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

		_, err = ParseLine(cg, line)
		time.Sleep(6 * time.Second) // 10 requests per second

		if err != nil {
			//fmt.Printf("%s: %s\n", x, err)
		} else {
			c++

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

	fmt.Println("done!")
}

func ParseLine(cg *coingecko.CoinGeckoAPI, line []byte) (string, error) {
	tokens := strings.Split(string(line), ",")
	if len(tokens) != 3 {
		return "", fmt.Errorf("invalid line: %s", string(line))
	}

	address := normalizeAddress(tokens[0])

	//chain := convertionMap[tokens[1]]

	_, err := cg.GetSymbolByContract(tokens[1], address)
	if err != nil {
		return address, err
	}

	//fmt.Printf("%s,%s\n", tokens[0], ti.Symbol)

	return "", nil
}

// TODO: add special rules for solana/terra/etc
func normalizeAddress(address string) string {

	// remove first 24 characters from address
	// 0x000000000000000000000000
	if len(address) > 24 {
		if strings.HasPrefix(address, "00000000000000") {
			return "0x" + address[24:]
		}
	} else {
		return base58encode(address)
	}

	return address
}

func base58encode(address string) string {
	//
	// Encode a byte slice into a base58-encoded string.
	ds, _ := hex.DecodeString(address)
	encoded := base58.Encode(ds)

	return encoded

}
