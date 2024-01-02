package prices

import (
	"fmt"
	"os"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/coingecko"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"go.uber.org/zap"
)

// go througth the symbol list provided by wormhole
// and fetch the history from coingecko
// and save it to a file
func RunPrices(output, p2pNetwork, coingeckoUrl, coingeckoHeaderKey, coingeckoApiKey string) {

	// build logger
	logger := logger.New("wormhole-explorer-analytics")

	logger.Info("starting wormhole-explorer-analytics ...")

	cg := coingecko.NewCoinGeckoAPI(coingeckoUrl, coingeckoHeaderKey, coingeckoApiKey)

	pricesOutput, err := os.Create(output)
	if err != nil {
		logger.Fatal("creating file", zap.Error(err))
	}
	defer pricesOutput.Close()

	// create token provider
	tokenProvider := domain.NewTokenProvider(p2pNetwork)
	tokens := tokenProvider.GetAllTokens()
	logger.Info("found tokens", zap.Int("count", len(tokens)))
	for index, token := range tokens {
		logger.Info("processing token",
			zap.String("coingeckoID", token.CoingeckoID),
			zap.Stringer("symbol", token.Symbol),
			zap.Int("index", index+1), zap.Int("count", len(tokens)))

		r, err := cg.GetSymbolDailyPrice(token.CoingeckoID, "max")
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, p := range r.Prices {
			pricesOutput.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s\n", token.TokenChain, token.CoingeckoID, token.Symbol, p[0], p[1]))
		}

		time.Sleep(5 * time.Second) // 10 requests per second

	}

	logger.Info("finished wormhole-explorer-analytics")

}
