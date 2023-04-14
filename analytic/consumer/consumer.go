package consumer

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/analytic/metric"
	"github.com/wormhole-foundation/wormhole-explorer/analytic/queue"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	vaa_sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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
			vaa, err := vaa.Unmarshal(event.Vaa)
			if err != nil {
				c.logger.Error("Invalid vaa", zap.String("id", event.ID), zap.Error(err))
				msg.Failed()
				continue
			}

			// filter vaa from pythnet.
			if c.p2pNetwork == domain.P2pMainNet && vaa_sdk.ChainIDPythNet == vaa.EmitterChain {
				c.logger.Debug("Skip vaa from pythnet", zap.String("id", event.ID))
				msg.Done()
				continue
			}

			// push vaa metrics.
			err = c.pushMetric(ctx, vaa)
			if err != nil {
				msg.Failed()
				continue
			}
			msg.Done()
			c.logger.Info("Pushed vaa metric", zap.String("id", event.ID))
		}
	}()
}
