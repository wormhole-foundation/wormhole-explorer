package consumer

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/processor"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consume queue.ConsumeFunc
	process processor.ProcessorFunc
	metrics metrics.Metrics
	logger  *zap.Logger
}

// New creates a new vaa consumer.
func New(consume queue.ConsumeFunc, process processor.ProcessorFunc, metrics metrics.Metrics, logger *zap.Logger) *Consumer {
	return &Consumer{consume: consume, process: process, metrics: metrics, logger: logger}
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for msg := range c.consume(ctx) {
			event := msg.Data()

			emitterChainID := sdk.ChainID(event.ChainID).String()

			// check id message is expired.
			if msg.IsExpired() {
				c.metrics.IncExpiredMessage(emitterChainID, event.Source)
				c.logger.Warn("Event expired", zap.String("id", event.ID))
				msg.Failed()
				continue
			}

			params := &processor.Params{
				TrackID: event.TrackID,
				Vaa:     event.Vaa,
			}
			_, err := c.process(ctx, params)
			if err != nil {
				c.metrics.IncUnprocessedMessage(emitterChainID, event.Source)
				c.logger.Error("Error processing event",
					zap.String("trackId", event.TrackID),
					zap.String("id", event.ID),
					zap.Error(err))
				msg.Failed()
				continue
			} else {
				c.metrics.IncProcessedMessage(emitterChainID, event.Source)
				c.logger.Debug("Event processed",
					zap.String("trackId", event.TrackID),
					zap.String("id", event.ID))
			}
			c.metrics.VaaProcessingDuration(emitterChainID, msg.SentTimestamp())
			msg.Done()
		}
	}()
}
