package consumer

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/analytics/metric"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consume    queue.VAAConsumeFunc
	pushMetric metric.MetricPushFunc
	logger     *zap.Logger
	p2pNetwork string
}

// New creates a new vaa consumer.
func New(consume queue.VAAConsumeFunc, pushMetric metric.MetricPushFunc, logger *zap.Logger, p2pNetwork string) *Consumer {
	return &Consumer{consume: consume, pushMetric: pushMetric, logger: logger, p2pNetwork: p2pNetwork}
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for msg := range c.consume(ctx) {
			event := msg.Data()

			// check id message is expired.
			if msg.IsExpired() {
				c.logger.Warn("Message with vaa expired", zap.String("id", event.ID))
				msg.Failed()
				continue
			}

			// unmarshal vaa.
			vaa, err := sdk.Unmarshal(event.Vaa)
			if err != nil {
				c.logger.Error("Invalid vaa", zap.String("id", event.ID), zap.Error(err))
				msg.Failed()
				continue
			}

			// push vaa metrics.
			err = c.pushMetric(ctx, vaa)
			if err != nil {
				msg.Failed()
				continue
			}

			msg.Done()
			c.logger.Debug("Pushed vaa metric", zap.String("id", event.ID))
		}
	}()
}
