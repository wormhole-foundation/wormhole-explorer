package consumer

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	"go.uber.org/zap"
)

const (
	maxAttempts = 5
	retryDelay  = 60 * time.Second
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc                queue.VAAConsumeFunc
	rpcServiceProviderSettings *config.RpcProviderSettings
	logger                     *zap.Logger
	repository                 *Repository
	workerPool                 *WorkerPool
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.VAAConsumeFunc,
	rpcServiceProviderSettings *config.RpcProviderSettings,
	ctx context.Context,
	logger *zap.Logger,
	repository *Repository,
	metrics metrics.Metrics,
) *Consumer {

	workerPool := NewWorkerPool(ctx, logger, rpcServiceProviderSettings, repository, metrics)

	c := Consumer{
		consumeFunc:                consumeFunc,
		rpcServiceProviderSettings: rpcServiceProviderSettings,
		logger:                     logger,
		repository:                 repository,
		workerPool:                 workerPool,
	}

	return &c
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go c.producerLoop(ctx)
}

func (c *Consumer) producerLoop(ctx context.Context) {

	ch := c.consumeFunc(ctx)

	for msg := range ch {

		// Send the VAA to the worker pool.
		//
		// The worker pool is responsible for calling `msg.Done()`
		err := c.workerPool.Push(ctx, msg)
		if err != nil {
			c.logger.Warn("failed to push message into worker pool",
				zap.String("vaaId", msg.Data().ID),
				zap.Error(err),
			)
			msg.Failed()
		}
	}
}
