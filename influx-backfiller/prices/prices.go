package prices

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type CoinPricesCache struct {
	filename string
	Prices   map[string]decimal.Decimal
}

func NewCoinPricesCache(priceFile string) *CoinPricesCache {
	return &CoinPricesCache{
		filename: priceFile,
		Prices:   make(map[string]decimal.Decimal),
	}
}

func (c *CoinPricesCache) GetPriceByTime(chainID int16, symbol string, day time.Time) (*decimal.Decimal, error) {

	// remove hours and minutes
	// times are in UTC
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)

	// generate key
	key := fmt.Sprintf("%d%s%d", chainID, symbol, day.UnixMilli())
	if price, ok := c.Prices[key]; ok {
		return &price, nil
	}
	return nil, fmt.Errorf("price not found for %s", key)
}

// load the csv file with prices into a map
func (cpc *CoinPricesCache) InitCache() {
	// open prices file
	file := cpc.filename
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// read line by line
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		row := scanner.Text()
		// split line by comma
		tokens := strings.Split(row, ",")
		if len(tokens) != 5 {
			panic(fmt.Errorf("invalid line: %s", row))
		}
		// build map key: chainid+coingecko_id+timestamp
		key := fmt.Sprintf("%s%s%s", tokens[0], tokens[1], tokens[3])

		price, err := decimal.NewFromString(tokens[4])
		if err != nil {
			msg := fmt.Sprintf("failed to parse price err=%v line=%s", err, row)
			panic(msg)
		}
		cpc.Prices[key] = price
	}

}
