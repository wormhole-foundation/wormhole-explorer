package backfiller

import (
	"context"
	"fmt"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/builder"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/config"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

func Run(cfg *config.Backfiller) {

	ctx := context.Background()

	logger := logger.New("wormhole-explorer-pipeline", logger.WithLevel(cfg.LogLevel))

	logger.Info("Starting wormhole-explorer-pipeline as backfiller ...")

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
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	// get alert client.
	alertClient := alert.NewDummyClient()

	// get metrics.
	metrics := metrics.NewDummyMetrics()

	// get publish function.
	pushFunc, err := builder.NewTopicProducer(ctx, cfg.AwsRegion, cfg.SNSUrl, cfg.AwsAccessKeyID,
		cfg.AwsSecretAccessKey, cfg.AwsEndpoint, alertClient, metrics, logger)
	if err != nil {
		logger.Fatal("failed to create publish function", zap.Error(err))
	}

	// create a vaa repository.
	vaaRepository := repository.NewVaaRepository(db.Database, logger)

	query := repository.VaaQuery{
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	limiter := ratelimit.New(int(cfg.RequestsPerSecond), ratelimit.Per(time.Second))

	pagination := repository.Pagination{
		Page:     0,
		PageSize: cfg.PageSize,
		SortAsc:  true,
	}

	queue := make(chan *repository.VaaDoc, 5*cfg.PageSize)

	var quantityProduced, quantityConsumed atomic.Uint64

	go getVaas(ctx, logger, pagination, query, vaaRepository, queue, &quantityProduced)

	var wg sync.WaitGroup
	wg.Add(cfg.NumWorkers)

	for i := 0; i < cfg.NumWorkers; i++ {
		name := fmt.Sprintf("worker-%d", i)
		log := logger.With(zap.String("worker", name))
		go publishVaa(ctx, pushFunc, queue, log, &wg, limiter, &quantityConsumed)
	}

	logger.Info("Waiting for all workers to finish...")
	wg.Wait()

	logger.Info("closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	logger.Info("Finish wormhole-explorer-pipeline as backfiller",
		zap.Uint64("produced", quantityProduced.Load()),
		zap.Uint64("consumed", quantityConsumed.Load()))
}

func getVaas(ctx context.Context, logger *zap.Logger, pagination repository.Pagination, query repository.VaaQuery,
	vaaRepository *repository.VaaRepository, queue chan *repository.VaaDoc, quantityProduced *atomic.Uint64) {
	defer close(queue)
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

		for _, vaa := range vaas {
			queue <- vaa
			quantityProduced.Add(1)
		}

		pagination.Page++
	}
	for {
		select {
		case <-time.After(10 * time.Second):
			if len(queue) == 0 {
				logger.Info("Closing, queue is empty")
				return
			}
		case <-ctx.Done():
			logger.Info("Closing due to cancelled context")
			return
		}
	}
}

func publishVaa(ctx context.Context, push topic.PushFunc, queue chan *repository.VaaDoc, logger *zap.Logger, wg *sync.WaitGroup,
	limiter ratelimit.Limiter, quantityConsumed *atomic.Uint64) {
	// Main loop: fetch global txs and process them
	defer wg.Done()
	for {
		select {

		// Try to pop a globalTransaction from the queue
		case vaa, ok := <-queue:
			// If the channel was closed, exit immediately
			if !ok {
				logger.Info("Closing, channel was closed")
				return
			}

			limiter.Take()

			if err := push(ctx, &topic.Event{
				ID:               vaa.ID,
				ChainID:          sdk.ChainID(vaa.ChainID),
				EmitterAddress:   vaa.EmitterAddress,
				Sequence:         vaa.Sequence,
				GuardianSetIndex: vaa.GuardianSetIndex,
				Vaa:              vaa.Vaa,
				IndexedAt:        vaa.IndexedAt,
				Timestamp:        vaa.Timestamp,
				UpdatedAt:        vaa.UpdatedAt,
				TxHash:           vaa.TxHash,
				Version:          uint16(vaa.Version),
				Revision:         uint16(vaa.Revision),
			}); err != nil {
				logger.Error("Failed to push vaa", zap.Error(err))
			} else {
				quantityConsumed.Add(1)
				logger.Debug("VAA pushed", zap.String("vaa_id", vaa.ID))
			}

			// If the context was cancelled, exit immediately
		case <-ctx.Done():
			logger.Info("Closing due to cancelled context")
			return
		}
	}
}
