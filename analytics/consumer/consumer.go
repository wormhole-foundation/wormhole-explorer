package consumer

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/analytics/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/metric"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consume    queue.ConsumeFunc
	pushMetric metric.MetricPushFunc
	logger     *zap.Logger
	metrics    metrics.Metrics
	p2pNetwork string
}

// New creates a new vaa consumer.
func New(consume queue.ConsumeFunc, pushMetric metric.MetricPushFunc, logger *zap.Logger, metrics metrics.Metrics, p2pNetwork string) *Consumer {
	return &Consumer{consume: consume, pushMetric: pushMetric, logger: logger, metrics: metrics, p2pNetwork: p2pNetwork}
}

// Start consumes messages from VAA queue, parse and store those messages in a repository.
func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for msg := range c.consume(ctx) {
			event := msg.Data()

			chainID := sdk.ChainID(event.ChainID).String()

			// check id message is expired.
			if msg.IsExpired() {
				msg.Failed()
				c.logger.Warn("Message with vaa expired", zap.String("id", event.ID))
				c.metrics.IncExpiredMessage(chainID, event.Source, msg.Retry())
				continue
			}

			// unmarshal vaa.
			vaa, err := sdk.Unmarshal(event.Vaa)
			if err != nil {
				msg.Done()
				c.logger.Error("Invalid vaa", zap.String("id", event.ID), zap.Error(err))
				c.metrics.IncInvalidMessage(chainID, event.Source, msg.Retry())
				continue
			}

			// push vaa metrics.
			err = c.pushMetric(
				ctx,
				&metric.Params{Source: event.Source, TrackID: event.TrackID, Vaa: vaa, VaaIsSigned: event.VaaIsSigned},
			)
			if err != nil {
				msg.Failed()
				c.metrics.IncUnprocessedMessage(chainID, event.Source, msg.Retry())
				continue
			}

			msg.Done()
			c.logger.Debug("Pushed vaa metric", zap.String("id", event.ID))
			c.metrics.IncProcessedMessage(chainID, event.Source, msg.Retry())
			c.metrics.VaaProcessingDuration(chainID, msg.SentTimestamp())
		}
	}()
}
