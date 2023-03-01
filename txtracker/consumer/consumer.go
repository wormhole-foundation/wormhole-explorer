package consumer

import (
	"context"

	//"github.com/wormhole-foundation/wormhole-explorer/analytic/metric"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consume queue.VAAConsumeFunc
	logger  *zap.Logger
}

// New creates a new vaa consumer.
func New(consume queue.VAAConsumeFunc, logger *zap.Logger) *Consumer {
	return &Consumer{consume: consume, logger: logger}
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for msg := range c.consume(ctx) {
			event := msg.Data()

			// check id message is expired.
			if msg.IsExpired() {
				c.logger.Warn("Message with VAA expired", zap.String("id", event.ID))
				msg.Failed()
				continue
			}

			// unmarshal vaa.
			_, err := vaa.Unmarshal(event.Vaa)
			if err != nil {
				c.logger.Error("Invalid VAA", zap.String("id", event.ID), zap.Error(err))
				msg.Failed()
				continue
			}

			//TODO: process message
			msg.Done()

			c.logger.Info("Processed VAA", zap.String("id", event.ID))
		}
	}()
}
