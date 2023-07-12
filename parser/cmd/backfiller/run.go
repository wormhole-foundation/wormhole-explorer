package backfiller

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/parser/config"
	"github.com/wormhole-foundation/wormhole-explorer/parser/http/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"github.com/wormhole-foundation/wormhole-explorer/parser/processor"
	"go.uber.org/zap"
)

func Run(config *config.BackfillerConfiguration) {

	rootCtx := context.Background()

	logger := logger.New("wormhole-explorer-parser", logger.WithLevel(config.LogLevel))

	logger.Info("Starting wormhole-explorer-parser  as backfiller ...")

	startTime, err := time.Parse(time.RFC3339, config.StartTime)
	if err != nil {
		logger.Fatal("failed to parse start time", zap.Error(err))
	}

	endTime := time.Now()
	if config.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, config.EndTime)
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
	db, err := db.New(rootCtx, logger, config.MongoURI, config.MongoDatabase)
	if err != nil {
		logger.Fatal("Failed to connect MongoDB", zap.Error(err))
	}

	parserVAAAPIClient, err := parser.NewParserVAAAPIClient(config.VaaPayloadParserTimeout, config.VaaPayloadParserURL, logger)
	if err != nil {
		logger.Fatal("Failed to create parse vaa api client")
	}

	parserRepository := parser.NewRepository(db.Database, logger)
	vaaRepository := vaa.NewRepository(db.Database, logger)

	//create a processor
	processor := processor.New(parserVAAAPIClient, parserRepository, alert.NewDummyClient(), metrics.NewDummyMetrics(), logger)

	logger.Info("Started wormhole-explorer-parser as backfiller")

	//start backfilling
	page := int64(0)
	for {
		logger.Info("Processing page", zap.Int64("page", page),
			zap.String("start_time", startTime.Format(time.RFC3339)),
			zap.String("end_time", endTime.Format(time.RFC3339)))

		vaas, err := vaaRepository.FindPageByTimeRange(rootCtx, startTime, endTime, page, config.PageSize, config.SortAsc)
		if err != nil {
			logger.Error("Failed to get vaas", zap.Error(err))
			break
		}

		if len(vaas) == 0 {
			logger.Info("Empty page", zap.Int64("page", page))
			break
		}
		for _, v := range vaas {
			logger.Debug("Processing vaa", zap.String("id", v.ID))
			_, err := processor.Process(rootCtx, v.Vaa)
			if err != nil {
				logger.Error("Failed to process vaa", zap.String("id", v.ID), zap.Error(err))
			}
		}
		page++
	}
	logger.Info("Closing database connections ...")
	db.Close()

	logger.Info("Finish wormhole-explorer-parser as backfiller")
}
