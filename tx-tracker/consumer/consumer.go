package consumer

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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
) *Consumer {

	workerPool := NewWorkerPool(ctx, logger, rpcServiceProviderSettings, repository)

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

		event := msg.Data()

		// Check if message is expired.
		if msg.IsExpired() {
			c.logger.Warn("Message with VAA expired", zap.String("id", event.ID))
			msg.Failed()
			continue
		}

		// Do not process messages from PythNet
		if event.ChainID == sdk.ChainIDPythNet {
			msg.Done()
			continue
		}

		// Send the VAA to the worker pool.
		p := ProcessSourceTxParams{
			VaaId:    event.ID,
			ChainId:  event.ChainID,
			Emitter:  event.EmitterAddress,
			Sequence: event.Sequence,
			TxHash:   event.TxHash,
		}
		err := c.workerPool.Push(ctx, &p)
		if err != nil {
			c.logger.Warn("failed to push message into worker pool",
				zap.String("vaaId", event.ID),
				zap.Error(err),
			)
			msg.Failed()
		}

		msg.Done()
	}
}
