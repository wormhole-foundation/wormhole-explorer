package prices

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/builder"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/token"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/metric"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/storage"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	apiPrices "github.com/wormhole-foundation/wormhole-explorer/common/prices"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type VaasPrices struct {
	DbLayer string //mongo, postgres, both

	// Mongo database configuration
	MongoUri string
	MongoDb  string

	// Postgres database configuration
	DbURL       string
	DbLogEnable bool

	PageSize            int64
	P2PNetwork          string
	NotionalUrl         string
	VaaPayloadParserUrl string
	StartTime           *time.Time
	EndTime             *time.Time
	EmitterChainID      *sdk.ChainID
	EmitterAddress      *string
	Sequence            *string
}

func RunVaasPrices(cfg VaasPrices) {

	ctx := context.Background()

	// build logger
	logger := logger.New("wormhole-explorer-analytics")

	logger.Info("starting wormhole-explorer-analytics ...")

	// init dummy metrics
	metrics := metrics.NewNoopMetrics()

	// setup DB connection
	storageLayer, err := builder.NewStorageLayer(ctx, cfg.DbLayer, cfg.MongoDb, cfg.MongoDb, cfg.DbURL, cfg.DbLogEnable, metrics, logger)
	if err != nil {
		logger.Fatal("failed to create to storage layer", zap.Error(err))
	}

	// create a parserVAAAPIClient
	parserVAAAPIClient, err := parser.NewParserVAAAPIClient(5, cfg.VaaPayloadParserUrl, logger)
	if err != nil {
		logger.Fatal("failed to create parse vaa api client")
	}

	// create a new prices repository
	pricesRepository := storageLayer.PricesRepository()

	// create a new VAA repository
	vaaRepository := storageLayer.VaaRepository()

	// create a token resolver
	tokenResolver := token.NewTokenResolver(parserVAAAPIClient, logger)

	// create a token provider
	tokenProvider := domain.NewTokenProvider(cfg.P2PNetwork)

	// create a price api
	api := apiPrices.NewPricesApi(cfg.NotionalUrl, logger)

	query := storage.VaaPageQuery{
		StartTime:      cfg.StartTime,
		EndTime:        cfg.EndTime,
		EmitterChainID: cfg.EmitterChainID,
		EmitterAddress: cfg.EmitterAddress,
		Sequence:       cfg.Sequence,
	}

	pagination := storage.Pagination{
		Page:     0,
		PageSize: cfg.PageSize,
		SortAsc:  true,
	}

	// start backfilling
	for {
		logger.Info("Processing page", zap.Any("pagination", pagination), zap.Any("query", query))

		vaas, err := vaaRepository.FindPage(ctx, query, pagination)
		if err != nil {
			logger.Error("Failed to get vaas", zap.Error(err))
			break
		}

		if len(vaas) == 0 {
			logger.Info("Empty page", zap.Int64("page", pagination.Page))
			break
		}
		for _, v := range vaas {
			logger.Debug("Processing vaa", zap.String("id", v.ID))
			vaa, err := sdk.Unmarshal(v.Vaa)
			if err != nil {
				logger.Error("Failed to unmarshal VAA", zap.Error(err))
				continue
			}

			transferredToken, err := tokenResolver.GetTransferredTokenByVaa(ctx, vaa)
			if err != nil {
				if !token.IsUnknownTokenErr(err) {
					logger.Error("Failed to obtain transferred token for this VAA",
						zap.String("vaaId", vaa.MessageID()),
						zap.Error(err))
				}
				continue
			}

			if err := metric.UpsertTransferPrices(
				ctx,
				logger,
				vaa,
				pricesRepository,
				func(tokenID, coinGeckoID string, timestamp time.Time) (decimal.Decimal, error) {
					price, err := api.GetPriceByTime(ctx, coinGeckoID, timestamp)
					if err != nil {
						return decimal.NewFromInt(0), err
					}
					return price, nil
				},
				transferredToken,
				tokenProvider,
				"backfiller",
				fmt.Sprintf("backfiller-%s", vaa.MessageID()),
			); err != nil {
				logger.Error("Failed to upsert transfer prices", zap.String("id", v.ID), zap.Error(err))
			}

		}
		pagination.Page++
	}

	logger.Info("finished wormhole-explorer-analytics")

}
