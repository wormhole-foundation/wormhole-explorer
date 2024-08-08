package metrics

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/analytics/builder"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/token"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/prices"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/storage"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"go.uber.org/zap"
)

type VaasVolume struct {
	DbLayer string //mongo, postgres, both

	// Mongo database configuration
	MongoUri string
	MongoDb  string

	// Postgres database configuration
	DbURL       string
	DbLogEnable bool

	PageSize            int64
	P2PNetwork          string
	OutputFile          string
	PricesFile          string
	VaaPayloadParserUrl string
	StartTime           *time.Time
	EndTime             *time.Time
}

// read a csv file with VAAs and convert into a decoded csv file
// ready to upload to the database
func RunVaaVolumeFromDb(cfg VaasVolume) {

	rootCtx := context.Background()

	// build logger
	logger := logger.New("wormhole-explorer-analytics")

	logger.Info("starting wormhole-explorer-analytics ...")

	// setup DB connection
	storageLayer, err := builder.NewStorageLayer(rootCtx, cfg.DbLayer, cfg.MongoDb, cfg.MongoDb, cfg.DbURL, cfg.DbLogEnable, logger)
	if err != nil {
		logger.Fatal("failed to create to storage layer", zap.Error(err))
	}

	// create a new VAA repository
	vaaRepository := storageLayer.VaaRepository()

	// create a parserVAAAPIClient
	parserVAAAPIClient, err := parser.NewParserVAAAPIClient(10, cfg.VaaPayloadParserUrl, logger)
	if err != nil {
		logger.Fatal("failed to create parse vaa api client")
	}

	// create a token resolver
	tokenResolver := token.NewTokenResolver(parserVAAAPIClient, logger)

	// create a token provider
	tokenProvider := domain.NewTokenProvider(cfg.P2PNetwork)

	// create missing tokens file
	missingTokensFile := "missing_tokens.csv"
	fmissingTokens, err := os.Create(missingTokensFile)
	if err != nil {
		logger.Fatal("creating missing tokens file", zap.Error(err))
	}
	defer fmissingTokens.Close()

	//open output file for writing
	fout, err := os.Create(cfg.OutputFile)
	if err != nil {
		logger.Fatal("creating output file", zap.Error(err))
	}
	defer fout.Close()

	// init price cache!
	logger.Info("loading historical prices...")
	priceCache := prices.NewCoinPricesCache(cfg.PricesFile)
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

		query := storage.VaaPageQuery{
			StartTime: &startTime,
			EndTime:   &endTime,
		}

		pagination := storage.Pagination{
			Page:     page,
			PageSize: cfg.PageSize,
			SortAsc:  true,
		}

		vaas, err := vaaRepository.FindPage(rootCtx, query, pagination)
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

	logger.Info("closing db connection...")
	storageLayer.Close()

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
