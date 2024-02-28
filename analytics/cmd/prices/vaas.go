package prices

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/token"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/metric"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	apiPrices "github.com/wormhole-foundation/wormhole-explorer/common/prices"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type VaasPrices struct {
	MongoUri            string
	MongoDb             string
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

	//setup DB connection
	db, err := dbutil.Connect(ctx, logger, cfg.MongoUri, cfg.MongoDb, false)
	if err != nil {
		logger.Fatal("Failed to connect MongoDB", zap.Error(err))
	}

	// get transfer prices collection
	transferPricesCollection := db.Database.Collection(repository.TransferPrices)

	// create a parserVAAAPIClient
	parserVAAAPIClient, err := parser.NewParserVAAAPIClient(5, cfg.VaaPayloadParserUrl, logger)
	if err != nil {
		logger.Fatal("failed to create parse vaa api client")
	}

	// create a new VAA repository
	vaaRepository := repository.NewVaaRepository(db.Database, logger)

	// create a token resolver
	tokenResolver := token.NewTokenResolver(parserVAAAPIClient, logger)

	// create a token provider
	tokenProvider := domain.NewTokenProvider(cfg.P2PNetwork)

	// create a price api
	api := apiPrices.NewPricesApi(cfg.NotionalUrl, logger)

	query := repository.VaaQuery{
		StartTime:      cfg.StartTime,
		EndTime:        cfg.EndTime,
		EmitterChainID: cfg.EmitterChainID,
		EmitterAddress: cfg.EmitterAddress,
		Sequence:       cfg.Sequence,
	}

	pagination := repository.Pagination{
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
				transferPricesCollection,
				func(tokenID, coinGeckoID string, timestamp time.Time) (decimal.Decimal, error) {
					price, err := api.GetPriceByTime(ctx, coinGeckoID, timestamp)
					if err != nil {
						return decimal.NewFromInt(0), err
					}
					return price, nil
				},
				transferredToken,
				tokenProvider,
			); err != nil {
				logger.Error("Failed to upsert transfer prices", zap.String("id", v.ID), zap.Error(err))
			}

		}
		pagination.Page++
	}

	logger.Info("finished wormhole-explorer-analytics")

}
