package backfiller

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	vaaPayloadParser "github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"github.com/wormhole-foundation/wormhole-explorer/parser/config"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole-explorer/parser/processor"
	"go.uber.org/zap"
)

func Run(cfg *config.BackfillerConfiguration) {
	rootCtx := context.Background()
	logger := logger.New("wormhole-explorer-parser", logger.WithLevel(cfg.LogLevel))

	if cfg.DbLayer == config.DbLayerMongo {
		runMongoBackfiller(rootCtx, logger, cfg)
	} else if cfg.DbLayer == config.DbLayerPostgres {
		runPostgresBackfiller(rootCtx, logger, cfg)
	} else {
		logger.Fatal("Invalid db layer", zap.String("db_layer", cfg.DbLayer))
	}

}

// runMongoBackfiller run parser backfiller for mongo database.
func runMongoBackfiller(ctx context.Context, logger *zap.Logger,
	cfg *config.BackfillerConfiguration) {

	logger.Info("Starting wormhole-explorer-parser as backfiller [mongo] ...")

	startTime, err := time.Parse(time.RFC3339, cfg.StartTime)
	if err != nil {
		logger.Fatal("failed to parse start time", zap.Error(err))
	}

	endTime := time.Now()
	if cfg.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, cfg.EndTime)
		if err != nil {
			logger.Fatal("Failed to parse end time", zap.Error(err))
		}
	}

	if startTime.After(endTime) {
		logger.Fatal("Start time should be before end time",
			zap.String("start_time", startTime.Format(time.RFC3339)),
			zap.String("end_time", endTime.Format(time.RFC3339)))
	}

	//setup DB connection
	db, err := dbutil.Connect(ctx, logger, cfg.MongoURI, cfg.MongoDatabase, false)
	if err != nil {
		logger.Fatal("Failed to connect MongoDB", zap.Error(err))
	}

	parserVAAAPIClient, err := vaaPayloadParser.NewParserVAAAPIClient(cfg.VaaPayloadParserTimeout, cfg.VaaPayloadParserURL, logger)
	if err != nil {
		logger.Fatal("Failed to create parse vaa api client")
	}

	query := repository.VaaQuery{
		StartTime:      &startTime,
		EndTime:        &endTime,
		EmitterChainID: cfg.EmitterChainID,
		EmitterAddress: cfg.EmitterAddress,
		Sequence:       cfg.Sequence,
	}

	pagination := repository.Pagination{
		Page:     0,
		PageSize: cfg.PageSize,
		SortAsc:  cfg.SortAsc,
	}

	parserRepository := parser.NewMongoRepository(db.Database, logger)
	vaaRepository := repository.NewVaaRepository(db.Database, logger)

	// create a token provider
	tokenProvider := domain.NewTokenProvider(cfg.P2pNetwork)

	//create a processor
	eventProcessor := processor.New(parserVAAAPIClient, cfg.DbLayer, parserRepository, nil,
		alert.NewDummyClient(), metrics.NewDummyMetrics(), tokenProvider, logger)

	logger.Info("Started wormhole-explorer-parser as backfiller [mongo]")

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
			p := &processor.Params{Vaa: v.Vaa, TrackID: fmt.Sprintf("backfiller-%s", v.ID)}
			_, err := eventProcessor.Process(ctx, p)
			if err != nil {
				logger.Error("Failed to process vaa", zap.String("id", v.ID), zap.Error(err))
			}
		}
		pagination.Page++
	}

	logger.Info("closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	logger.Info("Finish wormhole-explorer-parser as backfiller [mongo]")
}

// runPostgresBackfiller run parser backfiller for postgres database.
func runPostgresBackfiller(ctx context.Context, logger *zap.Logger,
	cfg *config.BackfillerConfiguration) {

	logger.Info("Starting wormhole-explorer-parser as backfiller [postgres] ...")

	// Parse start and end time.
	startTime, err := time.Parse(time.RFC3339, cfg.StartTime)
	if err != nil {
		logger.Fatal("failed to parse start time", zap.Error(err))
	}

	endTime := time.Now()
	if cfg.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, cfg.EndTime)
		if err != nil {
			logger.Fatal("Failed to parse end time", zap.Error(err))
		}
	}

	if startTime.After(endTime) {
		logger.Fatal("Start time should be before end time",
			zap.String("start_time", startTime.Format(time.RFC3339)),
			zap.String("end_time", endTime.Format(time.RFC3339)))
	}

	// create vaa query and pagination.
	query := repository.VaaQuery{
		StartTime:      &startTime,
		EndTime:        &endTime,
		EmitterChainID: cfg.EmitterChainID,
		EmitterAddress: cfg.EmitterAddress,
		Sequence:       cfg.Sequence,
	}

	pagination := repository.Pagination{
		Page:     0,
		PageSize: cfg.PageSize,
		SortAsc:  cfg.SortAsc,
	}

	// create postgres db.
	db, err := db.NewDB(ctx, cfg.PostgresDbURL)
	if err != nil {
		logger.Fatal("Failed to connect Postgres",
			zap.Error(err))
	}

	// create parser repository.
	parserRepository := parser.NewPostgresRepository(db, logger)

	// create vaa repositor.
	vaaRepository := repository.NewPostgresVaaRepository(db, logger)

	// create vaa-payload-parser http client.
	parserVAAAPIClient, err := vaaPayloadParser.NewParserVAAAPIClient(cfg.VaaPayloadParserTimeout,
		cfg.VaaPayloadParserURL, logger)
	if err != nil {
		logger.Fatal("Failed to create parse vaa api client")
	}

	// create a token provider.
	tokenProvider := domain.NewTokenProvider(cfg.P2pNetwork)

	// create parser processor.
	eventProcessor := processor.New(parserVAAAPIClient, cfg.DbLayer, nil, parserRepository,
		alert.NewDummyClient(), metrics.NewDummyMetrics(), tokenProvider, logger)

	logger.Info("Started wormhole-explorer-parser as backfiller [postgres]")

	for {
		logger.Info("Processing page", zap.Any("pagination", pagination), zap.Any("query", query))

		attestationVaas, err := vaaRepository.FindPage(ctx, query, pagination)
		if err != nil {
			logger.Error("Failed to get attestation vaas", zap.Error(err))
			break
		}

		if len(attestationVaas) == 0 {
			logger.Info("Empty page", zap.Int64("page", pagination.Page))
			break
		}

		for _, attestationVaa := range attestationVaas {
			logger.Debug("Processing attestation vaa", zap.String("id", attestationVaa.ID),
				zap.String("vaaId", attestationVaa.VaaID))
			p := &processor.Params{Vaa: attestationVaa.Raw, TrackID: fmt.Sprintf("backfiller-%s", attestationVaa.ID)}
			_, err := eventProcessor.Process(ctx, p)
			if err != nil {
				logger.Error("Failed to process attestation vaa",
					zap.String("id", attestationVaa.ID),
					zap.Error(err))
			}
		}
		pagination.Page++
	}

	logger.Info("closing Postgres connection...")
	db.Close()

	logger.Info("Finish wormhole-explorer-parser as backfiller [postgres]")
}
