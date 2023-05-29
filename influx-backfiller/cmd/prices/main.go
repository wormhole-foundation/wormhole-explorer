package prices

import (
	"fmt"
	"os"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/xlabs/influx-backfiller/coingecko"
)

// go througth the symbol list provided by wormhole
// and fetch the history from coingecko
// and save it to a file
func RunPrices(output string) {

	cg := coingecko.NewCoinGeckoAPI("")

	pricesOutput, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer pricesOutput.Close()

	for _, token := range domain.GetAllTokens() {
		fmt.Printf("%s [%s]\n", token.CoingeckoID, token.Symbol)
		r, err := cg.GetSymbolDailyPrice(token.CoingeckoID)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, p := range r.Prices {
			pricesOutput.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s\n", token.TokenChain, token.CoingeckoID, token.Symbol, p[0], p[1]))
		}

		time.Sleep(5 * time.Second) // 10 requests per second

	}

}
