package consumer

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consumeFunc         queue.VAAConsumeFunc
	rpcProviderSettings *config.RpcProviderSettings
	logger              *zap.Logger
	repository          *Repository
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.VAAConsumeFunc,
	rpcProviderSettings *config.RpcProviderSettings,
	ctx context.Context,
	logger *zap.Logger,
	repository *Repository,
) *Consumer {

	c := Consumer{
		consumeFunc:         consumeFunc,
		rpcProviderSettings: rpcProviderSettings,
		logger:              logger,
		repository:          repository,
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
		c.logger.Debug("Received message", zap.String("vaaId", msg.Data().ID))
		c.process(ctx, msg)
	}
}

func (c *Consumer) process(ctx context.Context, msg queue.ConsumerMessage) {

	event := msg.Data()

	// Do not process messages from PythNet
	if event.ChainID == sdk.ChainIDPythNet {
		if !msg.IsExpired() {
			c.logger.Debug("Deleting PythNet message", zap.String("vaaId", event.ID))
			msg.Done()
		} else {
			c.logger.Debug("Skipping expired PythNet message", zap.String("vaaId", event.ID))
		}
		return
	}

	// Skip non-processed, expired messages
	if msg.IsExpired() {
		c.logger.Warn("Message expired - skipping",
			zap.String("vaaId", event.ID),
			zap.Bool("isExpired", msg.IsExpired()),
		)
		return
	}

	// Process the VAA
	p := ProcessSourceTxParams{
		VaaId:     event.ID,
		ChainId:   event.ChainID,
		Emitter:   event.EmitterAddress,
		Sequence:  event.Sequence,
		TxHash:    event.TxHash,
		Overwrite: false, // avoid processing the same transaction twice
	}
	err := ProcessSourceTx(ctx, c.logger, c.rpcProviderSettings, c.repository, &p)

	// Log a message informing the processing status
	if err == chains.ErrChainNotSupported {
		c.logger.Info("Skipping VAA - chain not supported",
			zap.String("vaaId", event.ID),
		)
	} else if err == ErrAlreadyProcessed {
		c.logger.Warn("Message already processed - skipping",
			zap.String("vaaId", event.ID),
		)
	} else if err == chains.ErrTransactionNotFound {
		c.logger.Warn("Transaction not found - will retry after SQS visibilityTimeout",
			zap.String("vaaId", event.ID),
		)
		return
	} else if err != nil {
		c.logger.Error("Failed to process originTx",
			zap.String("vaaId", event.ID),
			zap.Error(err),
		)
	} else {
		c.logger.Info("Transaction processed successfully",
			zap.String("id", event.ID),
		)
	}

	// Mark the message as done
	//
	// If the message is expired, it will be put back into the queue.
	if !msg.IsExpired() {
		msg.Done()
	}
}
