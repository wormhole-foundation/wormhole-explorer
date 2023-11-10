package consumer

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/parser/processor"
	"github.com/wormhole-foundation/wormhole-explorer/parser/queue"
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

			// check id message is expired.
			if msg.IsExpired() {
				c.logger.Warn("Event expired", zap.String("id", event.ID))
				msg.Failed()
				continue
			}
			c.metrics.IncVaaUnexpired(event.ChainID)

			params := &processor.Params{
				TrackID: event.TrackID,
				Vaa:     event.Vaa,
			}
			_, err := c.process(ctx, params)
			if err != nil {
				c.logger.Error("Error processing event",
					zap.String("trackId", event.TrackID),
					zap.String("id", event.ID),
					zap.Error(err))
				msg.Failed()
				continue
			} else {
				c.logger.Debug("Event processed",
					zap.String("trackId", event.TrackID),
					zap.String("id", event.ID))
			}
			msg.Done()
		}
	}()
}
