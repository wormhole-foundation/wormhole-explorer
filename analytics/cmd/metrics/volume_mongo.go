package metrics

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/token"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/prices"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"go.uber.org/zap"
)

// read a csv file with VAAs and convert into a decoded csv file
// ready to upload to the database
func RunVaaVolumeFromMongo(mongoUri, mongoDb, outputFile, pricesFile, vaaPayloadParserURL, p2pNetwork string) {

	rootCtx := context.Background()

	// build logger
	logger := logger.New("wormhole-explorer-analytics")

	logger.Info("starting wormhole-explorer-analytics ...")

	//setup DB connection
	db, err := dbutil.Connect(rootCtx, logger, mongoUri, mongoDb, false)
	if err != nil {
		logger.Fatal("Failed to connect MongoDB", zap.Error(err))
	}

	// create a new VAA repository
	vaaRepository := repository.NewVaaRepository(db.Database, logger)

	// create a parserVAAAPIClient
	parserVAAAPIClient, err := parser.NewParserVAAAPIClient(10, vaaPayloadParserURL, logger)
	if err != nil {
		logger.Fatal("failed to create parse vaa api client")
	}

	// create a token resolver
	tokenResolver := token.NewTokenResolver(parserVAAAPIClient, logger)

	// create a token provider
	tokenProvider := domain.NewTokenProvider(p2pNetwork)

	// create missing tokens file
	missingTokensFile := "missing_tokens.csv"
	fmissingTokens, err := os.Create(missingTokensFile)
	if err != nil {
		logger.Fatal("creating missing tokens file", zap.Error(err))
	}
	defer fmissingTokens.Close()

	//open output file for writing
	fout, err := os.Create(outputFile)
	if err != nil {
		logger.Fatal("creating output file", zap.Error(err))
	}
	defer fout.Close()

	// init price cache!
	logger.Info("loading historical prices...")
	priceCache := prices.NewCoinPricesCache(pricesFile)
	priceCache.InitCache()
	converter := NewVaaConverter(priceCache, tokenResolver.GetTransferredTokenByVaa, tokenProvider)
	logger.Info("loaded historical prices")

	endTime := time.Now()
	startTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	// start backfilling
	page := int64(0)
	c := 0
	for {
		logger.Info("Processing page", zap.Int64("page", page),
			zap.String("start_time", startTime.Format(time.RFC3339)),
			zap.String("end_time", endTime.Format(time.RFC3339)))

		vaas, err := vaaRepository.FindPageByTimeRange(rootCtx, startTime, endTime, page, 1000, true)
		if err != nil {
			logger.Error("Failed to get vaas", zap.Error(err))
			break
		}

		if len(vaas) == 0 {
			logger.Info("Empty page", zap.Int64("page", page))
			break
		}
		for i, v := range vaas {
			logger.Debug("Processing vaa", zap.String("id", v.ID))
			_, _, nl, err := converter.Convert(rootCtx, v.Vaa)
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
		page++
	}

	logger.Info("closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	missingTokensCount := 0
	converter.MissingTokensCounter.Range(
		func(key, value interface{}) bool {
			fmissingTokens.WriteString(fmt.Sprintf("%s,%d\n", key.(string), value.(uint64)))
			missingTokensCount++
			return true
		})

	logger.Info("missing tokens", zap.Int("count", missingTokensCount))

	logger.Info("finished wormhole-explorer-analytics")

}
