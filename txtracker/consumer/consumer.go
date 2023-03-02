package consumer

import (
	"context"

	//"github.com/wormhole-foundation/wormhole-explorer/analytic/metric"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Consumer consumer struct definition.
type Consumer struct {
	consume queue.VAAConsumeFunc
	cfg     *config.Settings
	logger  *zap.Logger
}

// New creates a new vaa consumer.
func New(
	consumeFunc queue.VAAConsumeFunc,
	cfg *config.Settings,
	logger *zap.Logger,
) *Consumer {

	c := Consumer{
		consume: consumeFunc,
		cfg:     cfg,
		logger:  logger,
	}

	return &c
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

			// do not process messages from PythNet
			if event.ChainID == sdk.ChainIDPythNet {
				msg.Done()
				continue
			}

			// process message
			txDetail, err := chains.FetchTx(ctx, c.cfg, event.ChainID, event.TxHash)
			if err != nil {
				c.logger.Warn("Failed to fetch source transaction details from VAA",
					zap.String("id", event.ID),
					zap.String("chain", event.ChainID.String()),
					zap.Error(err),
				)
			} else {
				c.logger.Debug("Successfuly obtained source transaction details from VAA",
					zap.String("id", event.ID),
					zap.Any("details", txDetail),
				)
			}
			msg.Done()
		}
	}()
}
